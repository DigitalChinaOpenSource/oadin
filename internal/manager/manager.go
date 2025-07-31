package manager

import (
	"context"
	"sync"
	"time"

	"oadin/config"
	"oadin/internal/logger"
	"oadin/internal/provider"
	"oadin/internal/types"
)

// isEmbedService 判断是否为embed服务
func isEmbedService(serviceType string) bool {
	return serviceType == types.ServiceEmbed
}

// isLocalNonEmbedService 判断是否为本地非embed服务
func isLocalNonEmbedService(location, serviceType string) bool {
	return location == types.ServiceSourceLocal && !isEmbedService(serviceType)
}

// NeedsQueuing 判断请求是否需要进入排队机制（导出函数）
func NeedsQueuing(location, serviceType string) bool {
	return isLocalNonEmbedService(location, serviceType)
}

// Manager 模型管理器
type Manager struct {
	queue   *Queue   // 请求队列
	loader  *Loader  // 模型加载器
	cleaner *Cleaner // 自动清理器
	mutex   sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

// GetModelManager 获取全局模型管理器实例
func GetModelManager() *Manager {
	once.Do(func() {
		// 从配置中获取参数，如果配置未初始化则使用默认值
		queueSize := 10
		queueTimeout := 30 * time.Second
		idleTimeout := 10 * time.Minute
		cleanupInterval := 1 * time.Minute

		// 尝试从全局配置获取参数
		if config.GlobalOADINEnvironment != nil {
			if config.GlobalOADINEnvironment.LocalModelQueueSize > 0 {
				queueSize = config.GlobalOADINEnvironment.LocalModelQueueSize
			}
			if config.GlobalOADINEnvironment.LocalModelQueueTimeout > 0 {
				queueTimeout = config.GlobalOADINEnvironment.LocalModelQueueTimeout
			}
			if config.GlobalOADINEnvironment.ModelIdleTimeout > 0 {
				idleTimeout = config.GlobalOADINEnvironment.ModelIdleTimeout
			}
			if config.GlobalOADINEnvironment.ModelCleanupInterval > 0 {
				cleanupInterval = config.GlobalOADINEnvironment.ModelCleanupInterval
			}
		}

		// 创建组件
		queue := NewQueue(queueSize, queueTimeout)
		loader := NewLoader()
		cleaner := NewCleaner(loader, idleTimeout)

		instance = &Manager{
			queue:   queue,
			loader:  loader,
			cleaner: cleaner,
		}

		logger.LogicLogger.Info("[Manager] Initialized with config",
			"queue_size", queueSize,
			"queue_timeout", queueTimeout,
			"idle_timeout", idleTimeout,
			"cleanup_interval", cleanupInterval)

		// 简化初始化：不再自动发现运行中的模型
	})
	return instance
}

// GetModelMemoryManager 获取全局模型管理器实例（保持向后兼容）
func GetModelMemoryManager() *Manager {
	return GetModelManager()
}

// Start 启动模型管理器
func (m *Manager) Start(cleanupInterval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果传入的清理间隔为0，使用配置中的默认值
	if cleanupInterval == 0 {
		if config.GlobalOADINEnvironment != nil && config.GlobalOADINEnvironment.ModelCleanupInterval > 0 {
			cleanupInterval = config.GlobalOADINEnvironment.ModelCleanupInterval
		} else {
			cleanupInterval = 1 * time.Minute // 默认1分钟
		}
	}

	// 启动队列
	m.queue.Start()

	// 启动清理器
	m.cleaner.Start(cleanupInterval)

	logger.LogicLogger.Info("[Manager] Started",
		"cleanup_interval", cleanupInterval)
}

// Stop 停止模型管理器
func (m *Manager) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 停止队列
	m.queue.Stop()

	// 停止清理器
	m.cleaner.Stop()

	logger.LogicLogger.Info("[Manager] Stopped")
}

// SetIdleTimeout 设置空闲超时时间
func (m *Manager) SetIdleTimeout(timeout time.Duration) {
	m.cleaner.SetIdleTimeout(timeout)
}

// MarkModelInUse 标记模型为使用中
func (m *Manager) MarkModelInUse(modelName string) error {
	return m.loader.MarkModelInUse(modelName)
}

// MarkModelIdle 标记模型为空闲
func (m *Manager) MarkModelIdle(modelName string) error {
	return m.loader.MarkModelIdle(modelName)
}

// GetModelState 获取模型状态信息
func (m *Manager) GetModelState(modelName string) (*ModelState, bool) {
	return m.loader.GetModelState(modelName)
}

// GetAllModelStates 获取所有模型状态信息
func (m *Manager) GetAllModelStates() map[string]*ModelState {
	return m.loader.GetAllModelStates()
}

// ForceUnloadModel 强制卸载指定模型
func (m *Manager) ForceUnloadModel(modelName string) error {
	return m.loader.ForceUnloadModel(modelName)
}

// EnqueueLocalModelRequest 将本地模型请求加入队列
func (m *Manager) EnqueueLocalModelRequest(ctx context.Context, modelName string, providerInstance provider.ModelServiceProvider, providerName, providerType string, taskID uint64) error {
	// 创建排队请求
	request := &QueuedRequest{
		TaskID:       taskID,
		ModelName:    modelName,
		Provider:     providerInstance,
		ProviderName: providerName,
		ProviderType: providerType,
		Context:      ctx,
		StartTime:    time.Now(),
		CompleteChan: make(chan struct{}),
	}

	// 加入队列（非阻塞）
	if err := m.queue.EnqueueRequest(request); err != nil {
		return err
	}

	// 在后台处理模型切换
	go m.processModelRequest(request)

	logger.LogicLogger.Debug("[Manager] Local model request enqueued",
		"task_id", taskID,
		"model", modelName)

	return nil
}

// processModelRequest 处理模型请求（后台执行）
func (m *Manager) processModelRequest(request *QueuedRequest) {
	logger.LogicLogger.Debug("[Manager] Processing model request",
		"task_id", request.TaskID,
		"model", request.ModelName)

	// 简化处理：只记录请求信息，实际的模型加载由原有流程处理
	logger.LogicLogger.Info("[Manager] Model request queued and ready for processing",
		"task_id", request.TaskID,
		"model", request.ModelName,
		"provider", request.ProviderName)

	// 通知处理完成
	close(request.CompleteChan)
}

// CompleteLocalModelRequest 完成本地模型请求处理
func (m *Manager) CompleteLocalModelRequest(taskID uint64) {
	logger.LogicLogger.Debug("[Manager] Completing local model request", "task_id", taskID)
	m.queue.CompleteRequest(taskID)
}

// GetStats 获取管理器统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"queue":   m.queue.GetStats(),
		"loader":  m.loader.GetStats(),
		"cleaner": m.cleaner.GetStats(),
	}

	return stats
}
