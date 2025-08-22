package manager

import (
	"context"
	"fmt"
	"time"

	"os"
	"sync"

	"oadin/internal/logger"
	"oadin/internal/provider"
	"oadin/internal/types"
)

// ModelStatus 模型状态枚举
type ModelStatus int

const (
	ModelStatusUnloaded  ModelStatus = iota // 未加载
	ModelStatusLoading                      // 加载中
	ModelStatusIdle                         // 空闲
	ModelStatusInUse                        // 使用中
	ModelStatusUnloading                    // 卸载中
)

func (s ModelStatus) String() string {
	switch s {
	case ModelStatusUnloaded:
		return "unloaded"
	case ModelStatusLoading:
		return "loading"
	case ModelStatusIdle:
		return "idle"
	case ModelStatusInUse:
		return "in_use"
	case ModelStatusUnloading:
		return "unloading"
	default:
		return "unknown"
	}
}

// ModelState 模型状态信息
type ModelState struct {
	ModelName    string                        // 模型名称
	ProviderName string                        // 提供商名称
	ProviderType string                        // 提供商类型
	Status       ModelStatus                   // 当前状态
	LastUsedTime time.Time                     // 最后使用时间
	LoadedTime   time.Time                     // 加载时间
	RefCount     int                           // 引用计数（并发使用数）
	Provider     provider.ModelServiceProvider // 提供商实例
	mutex        sync.RWMutex                  // 状态锁
}

// UpdateLastUsedTime 更新最后使用时间
func (ms *ModelState) UpdateLastUsedTime() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.LastUsedTime = time.Now()
}

// IncrementRef 增加引用计数并设置为使用中
func (ms *ModelState) IncrementRef() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.RefCount++
	if ms.RefCount > 0 {
		ms.Status = ModelStatusInUse
	}
	ms.LastUsedTime = time.Now()
}

// DecrementRef 减少引用计数，如果为0则设置为空闲
func (ms *ModelState) DecrementRef() {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	if ms.RefCount > 0 {
		ms.RefCount--
	}
	if ms.RefCount == 0 {
		ms.Status = ModelStatusIdle
		ms.LastUsedTime = time.Now()
	}
}

// GetStatus 获取当前状态（线程安全）
func (ms *ModelState) GetStatus() ModelStatus {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.Status
}

// SetStatus 设置状态（线程安全）
func (ms *ModelState) SetStatus(status ModelStatus) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.Status = status
}

// GetRefCount 获取引用计数（线程安全）
func (ms *ModelState) GetRefCount() int {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.RefCount
}

// IsIdle 检查是否空闲且超过指定时间
func (ms *ModelState) IsIdle(idleTimeout time.Duration) bool {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.Status == ModelStatusIdle &&
		ms.RefCount == 0 &&
		time.Since(ms.LastUsedTime) > idleTimeout
}

// Loader 模型加载管理器
type Loader struct {
	stateManager ModelStateManager      // 使用接口而非具体类型
	models       map[string]*ModelState // 模型状态映射 key: modelName
	mutex        sync.RWMutex           // 读写锁
	loadCond     *sync.Cond             // 用于等待模型加载完成的条件变量
}

// NewLoader 创建新的加载器
func NewLoader(stateManager ModelStateManager) *Loader {
	l := &Loader{
		stateManager: stateManager,
		models:       make(map[string]*ModelState),
	}
	l.loadCond = sync.NewCond(&l.mutex)
	return l
}

// EnsureModelLoaded 确保模型已加载，如果未加载则自动加载
func (l *Loader) EnsureModelLoaded(ctx context.Context, modelName string, providerInstance provider.ModelServiceProvider, providerName, providerType string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	modelState, exists := l.models[modelName]
	if !exists {
		// 创建新的模型状态
		modelState = &ModelState{
			ModelName:    modelName,
			ProviderName: providerName,
			ProviderType: providerType,
			Status:       ModelStatusUnloaded,
			Provider:     providerInstance,
		}
		l.models[modelName] = modelState
	}

	// 检查当前状态
	switch modelState.GetStatus() {
	case ModelStatusInUse, ModelStatusIdle:
		// 模型已加载，直接返回
		return nil
	case ModelStatusLoading:
		// 模型正在加载中，等待加载完成
		return l.waitForModelLoaded(ctx, modelName)
	case ModelStatusUnloading:
		// 模型正在卸载中，等待卸载完成后重新加载
		if err := l.waitForModelUnloaded(ctx, modelName); err != nil {
			return err
		}
		fallthrough
	case ModelStatusUnloaded:
		// 需要加载模型
		return l.loadModel(ctx, modelState)
	default:
		return fmt.Errorf("unknown model status: %v", modelState.GetStatus())
	}
}

// loadModel 加载模型
func (l *Loader) loadModel(ctx context.Context, modelState *ModelState) error {
	modelState.SetStatus(ModelStatusLoading)

	logger.LogicLogger.Info("[Loader] Loading model",
		"model", modelState.ModelName,
		"provider", modelState.ProviderName)

	// 释放锁，避免在加载过程中阻塞其他操作
	l.mutex.Unlock()
	defer l.mutex.Lock()

	// 调用provider的LoadModel方法
	loadReq := &types.LoadRequest{
		Model: modelState.ModelName,
	}

	if err := modelState.Provider.LoadModel(ctx, loadReq); err != nil {
		modelState.SetStatus(ModelStatusUnloaded)
		l.loadCond.Broadcast() // 通知等待的goroutine
		return fmt.Errorf("failed to load model %s: %w", modelState.ModelName, err)
	}

	// 更新状态
	modelState.SetStatus(ModelStatusIdle)
	modelState.LoadedTime = time.Now()
	modelState.LastUsedTime = time.Now()

	logger.LogicLogger.Info("[Loader] Model loaded successfully",
		"model", modelState.ModelName)

	l.loadCond.Broadcast() // 通知等待的goroutine
	return nil
}

// unloadModel 卸载模型
func (l *Loader) unloadModel(modelState *ModelState) error {
	modelState.SetStatus(ModelStatusUnloading)

	logger.LogicLogger.Info("[Loader] Unloading model",
		"model", modelState.ModelName)

	// 调用provider的UnloadModel方法
	unloadReq := &types.UnloadModelRequest{
		Models: []string{modelState.ModelName},
	}

	ctx := context.Background()
	if err := modelState.Provider.UnloadModel(ctx, unloadReq); err != nil {
		modelState.SetStatus(ModelStatusIdle) // 恢复到空闲状态
		return fmt.Errorf("failed to unload model %s: %w", modelState.ModelName, err)
	}

	// 从管理器中移除
	l.mutex.Lock()
	delete(l.models, modelState.ModelName)

	// 如果这是当前模型，清空当前模型
	if l.stateManager.GetCurrentModel() == modelState.ModelName {
		l.stateManager.SetCurrentModel("")
	}
	l.mutex.Unlock()

	logger.LogicLogger.Info("[Loader] Model unloaded successfully",
		"model", modelState.ModelName)

	return nil
}

// waitForModelLoaded 等待模型加载完成
func (l *Loader) waitForModelLoaded(ctx context.Context, modelName string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			modelState, exists := l.models[modelName]
			if !exists {
				return fmt.Errorf("model %s not found", modelName)
			}

			status := modelState.GetStatus()
			if status == ModelStatusIdle || status == ModelStatusInUse {
				return nil
			}
			if status == ModelStatusUnloaded {
				return fmt.Errorf("model %s failed to load", modelName)
			}

			// 等待状态变化
			l.loadCond.Wait()
		}
	}
}

// waitForModelUnloaded 等待模型卸载完成
func (l *Loader) waitForModelUnloaded(ctx context.Context, modelName string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, exists := l.models[modelName]
			if !exists {
				return nil // 模型已被移除，卸载完成
			}

			// 等待状态变化
			l.loadCond.Wait()
		}
	}
}

// SwitchModel 切换模型
func (l *Loader) SwitchModel(newModel string, provider provider.ModelServiceProvider) error {
	currentModel := l.stateManager.GetCurrentModel()

	logger.LogicLogger.Info("[Loader] Starting model switch operation",
		"from", currentModel, "to", newModel)

	// 如果是同一个模型，直接返回
	if currentModel == newModel {
		logger.LogicLogger.Debug("[Loader] Model already loaded, no switch needed",
			"model", newModel)
		return nil
	}

	// 步骤1：卸载当前模型（如果存在）
	if currentModel != "" {
		logger.LogicLogger.Info("[Loader] Step 1: Unloading current model",
			"model", currentModel)

		if err := l.UnloadModel(currentModel, provider); err != nil {
			logger.LogicLogger.Error("[Loader] Failed to unload current model",
				"model", currentModel, "error", err)
			// 卸载失败，但继续尝试加载新模型
			// 这可能导致资源冲突，但至少尝试恢复
		} else {
			logger.LogicLogger.Info("[Loader] Current model unloaded successfully",
				"model", currentModel)
		}

		// 确保当前模型状态已清空
		l.stateManager.SetCurrentModel("")
		logger.LogicLogger.Debug("[Loader] Current model state cleared")
	}

	// 步骤2：加载新模型
	logger.LogicLogger.Info("[Loader] Step 2: Loading new model",
		"model", newModel)

	if err := l.LoadModel(newModel, provider); err != nil {
		logger.LogicLogger.Error("[Loader] Failed to load new model",
			"model", newModel, "error", err)
		return fmt.Errorf("failed to load model %s: %w", newModel, err)
	}

	// 步骤3：更新当前模型状态
	l.stateManager.SetCurrentModel(newModel)
	logger.LogicLogger.Info("[Loader] Model switch completed successfully",
		"from", currentModel, "to", newModel)

	return nil
}

// LoadModel 加载模型
func (l *Loader) LoadModel(modelName string, provider provider.ModelServiceProvider) error {
	ctx := context.Background()

	logger.LogicLogger.Debug("[Loader] LoadModel called",
		"model", modelName)

	// 获取provider信息
	providerName := "unknown"
	providerType := "unknown"

	// 尝试从provider获取名称和类型信息
	if provider != nil {
		// 这里可以根据实际的provider接口获取信息
		// 暂时使用默认值
	}

	err := l.EnsureModelLoaded(ctx, modelName, provider, providerName, providerType)
	if err != nil {
		logger.LogicLogger.Error("[Loader] LoadModel failed",
			"model", modelName, "error", err)
		return err
	}

	logger.LogicLogger.Info("[Loader] LoadModel completed successfully",
		"model", modelName)
	return nil
}

// UnloadModel 卸载模型
func (l *Loader) UnloadModel(modelName string, provider provider.ModelServiceProvider) error {
	logger.LogicLogger.Debug("[Loader] UnloadModel called",
		"model", modelName)

	err := l.ForceUnloadModel(modelName)
	if err != nil {
		logger.LogicLogger.Error("[Loader] UnloadModel failed",
			"model", modelName, "error", err)
		return err
	}

	logger.LogicLogger.Info("[Loader] UnloadModel completed successfully",
		"model", modelName)
	return nil
}

// MarkModelInUse 标记模型为使用中
func (l *Loader) MarkModelInUse(modelName string) error {
	l.mutex.RLock()
	modelState, exists := l.models[modelName]
	l.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	modelState.IncrementRef()
	logger.LogicLogger.Debug("[Loader] Model marked as in use",
		"model", modelName, "ref_count", modelState.GetRefCount())

	return nil
}

// MarkModelIdle 标记模型为空闲
func (l *Loader) MarkModelIdle(modelName string) error {
	l.mutex.RLock()
	modelState, exists := l.models[modelName]
	l.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("model %s not found", modelName)
	}

	modelState.DecrementRef()
	logger.LogicLogger.Debug("[Loader] Model marked as idle",
		"model", modelName, "ref_count", modelState.GetRefCount())

	return nil
}

// GetModelState 获取模型状态信息
func (l *Loader) GetModelState(modelName string) (*ModelState, bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	modelState, exists := l.models[modelName]
	return modelState, exists
}

// GetAllModelStates 获取所有模型状态信息
func (l *Loader) GetAllModelStates() map[string]*ModelState {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	result := make(map[string]*ModelState)
	for k, v := range l.models {
		result[k] = v
	}
	return result
}

// ForceUnloadModel 强制卸载指定模型
func (l *Loader) ForceUnloadModel(modelName string) error {
	logger.LogicLogger.Debug("[Loader] ForceUnloadModel called",
		"model", modelName)

	l.mutex.RLock()
	modelState, exists := l.models[modelName]
	l.mutex.RUnlock()

	if !exists {
		logger.LogicLogger.Debug("[Loader] Model not found in loader, assuming already unloaded",
			"model", modelName)
		// 模型不存在，可能已经被卸载了，这不是错误
		return nil
	}

	logger.LogicLogger.Info("[Loader] Force unloading model",
		"model", modelName, "current_status", modelState.GetStatus())

	// 强制设置引用计数为0
	modelState.mutex.Lock()
	modelState.RefCount = 0
	modelState.Status = ModelStatusIdle
	modelState.mutex.Unlock()

	err := l.unloadModel(modelState)
	if err != nil {
		logger.LogicLogger.Error("[Loader] ForceUnloadModel failed",
			"model", modelName, "error", err)
		return err
	}

	logger.LogicLogger.Info("[Loader] ForceUnloadModel completed successfully",
		"model", modelName)
	return nil
}

// GetCurrentModel 获取当前加载的模型
func (l *Loader) GetCurrentModel() string {
	return l.stateManager.GetCurrentModel()
}

// InitializeRunningModels 初始化时获取已运行的模型并清理它们
func (l *Loader) InitializeRunningModels() {
	logger.LogicLogger.Info("[Loader] Starting initialization - checking for running models to cleanup")

	// 支持GetRunningModels方法的provider列表
	supportedProviders := []string{
		types.FlavorOllama,
		types.FlavorOpenvino,
	}

	for _, flavor := range supportedProviders {
		providerInstance := provider.GetModelEngine(flavor)
		if providerInstance == nil {
			continue

			execPath := providerInstance.GetConfig().ExecPath
			if _, err := os.Stat(execPath); err != nil {
				continue
			}

			l.cleanupProviderModels(flavor, providerInstance)
		}

		logger.LogicLogger.Info("[Loader] Initialization complete - all running models cleaned up")
	}
}

// cleanupProviderModels 清理指定provider的运行模型
func (l *Loader) cleanupProviderModels(flavor string, providerInstance provider.ModelServiceProvider) {
	// 检查provider是否支持GetRunningModels方法
	runningModelsProvider, ok := providerInstance.(interface {
		GetRunningModels(context.Context) (*types.ListResponse, error)
	})
	if !ok {
		logger.LogicLogger.Debug("[Loader] Provider does not support GetRunningModels",
			"provider", flavor)
		return
	}

	ctx := context.Background()
	listResp, err := runningModelsProvider.GetRunningModels(ctx)
	if err != nil {
		logger.LogicLogger.Warn("[Loader] Failed to get running models from provider",
			"provider", flavor, "error", err)
		return
	}

	if listResp == nil || listResp.Models == nil {
		logger.LogicLogger.Debug("[Loader] No running models found to cleanup",
			"provider", flavor)
		return
	}

	logger.LogicLogger.Info("[Loader] Found running models to cleanup",
		"provider", flavor, "count", len(listResp.Models))

	// 异步清理每个运行中的模型
	for _, model := range listResp.Models {
		if model.Name == "" {
			continue
		}

		logger.LogicLogger.Info("[Loader] Scheduling cleanup for running model",
			"model", model.Name, "provider", flavor)

		// 异步卸载模型
		go l.asyncUnloadModel(model.Name, providerInstance, flavor)
	}
}

// asyncUnloadModel 异步卸载模型
func (l *Loader) asyncUnloadModel(modelName string, providerInstance provider.ModelServiceProvider, flavor string) {
	logger.LogicLogger.Info("[Loader] Starting async unload for model",
		"model", modelName, "provider", flavor)

	// 创建卸载请求
	unloadReq := &types.UnloadModelRequest{
		Models: []string{modelName},
	}

	ctx := context.Background()
	err := providerInstance.UnloadModel(ctx, unloadReq)
	if err != nil {
		logger.LogicLogger.Error("[Loader] Failed to unload running model during cleanup",
			"model", modelName, "provider", flavor, "error", err)
		return
	}

	logger.LogicLogger.Info("[Loader] Successfully unloaded running model during cleanup",
		"model", modelName, "provider", flavor)
}

// GetStats 获取加载器统计信息
func (l *Loader) GetStats() map[string]interface{} {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_models":  len(l.models),
		"current_model": l.stateManager.GetCurrentModel(),
	}

	statusCount := make(map[string]int)
	for _, modelState := range l.models {
		status := modelState.GetStatus().String()
		statusCount[status]++
	}
	stats["status_count"] = statusCount

	return stats
}

// GetIdleModels 获取空闲超时的模型列表
func (l *Loader) GetIdleModels(idleTimeout time.Duration) []*ModelState {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	idleModels := make([]*ModelState, 0)
	for _, modelState := range l.models {
		if modelState.IsIdle(idleTimeout) {
			idleModels = append(idleModels, modelState)
		}
	}

	return idleModels
}
