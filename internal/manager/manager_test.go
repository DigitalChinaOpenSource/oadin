package manager

import (
	"context"
	"testing"
	"time"

	"oadin/internal/types"
)

// MockProvider 模拟的模型服务提供商
type MockProvider struct {
	loadedModels map[string]bool
	loadError    error
	unloadError  error
	loadDelay    time.Duration
	unloadDelay  time.Duration
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		loadedModels: make(map[string]bool),
	}
}

func (m *MockProvider) LoadModel(ctx context.Context, req *types.LoadRequest) error {
	if m.loadDelay > 0 {
		time.Sleep(m.loadDelay)
	}
	if m.loadError != nil {
		return m.loadError
	}
	m.loadedModels[req.Model] = true
	return nil
}

func (m *MockProvider) UnloadModel(ctx context.Context, req *types.UnloadModelRequest) error {
	if m.unloadDelay > 0 {
		time.Sleep(m.unloadDelay)
	}
	if m.unloadError != nil {
		return m.unloadError
	}
	for _, model := range req.Models {
		delete(m.loadedModels, model)
	}
	return nil
}

// 实现其他必需的接口方法（空实现）
func (m *MockProvider) InstallEngine() error          { return nil }
func (m *MockProvider) StartEngine(mode string) error { return nil }
func (m *MockProvider) StopEngine() error             { return nil }
func (m *MockProvider) HealthCheck() error            { return nil }
func (m *MockProvider) InitEnv() error                { return nil }
func (m *MockProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	return nil, nil
}

func (m *MockProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	return nil, nil
}
func (m *MockProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error { return nil }
func (m *MockProvider) ListModels(ctx context.Context) (*types.ListResponse, error)     { return nil, nil }
func (m *MockProvider) GetRunningModels(ctx context.Context) (*types.ListResponse, error) {
	return nil, nil
}
func (m *MockProvider) GetConfig() *types.EngineRecommendConfig { return nil }
func (m *MockProvider) GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error) {
	return nil, nil
}

func TestServiceTypeJudgment(t *testing.T) {
	// 测试embed服务判断
	if !isEmbedService(types.ServiceEmbed) {
		t.Error("ServiceEmbed should be identified as embed service")
	}

	if isEmbedService(types.ServiceChat) {
		t.Error("ServiceChat should not be identified as embed service")
	}

	// 测试本地非embed服务判断
	if !isLocalNonEmbedService(types.ServiceSourceLocal, types.ServiceChat) {
		t.Error("Local chat service should need queuing")
	}

	if isLocalNonEmbedService(types.ServiceSourceLocal, types.ServiceEmbed) {
		t.Error("Local embed service should not need queuing")
	}

	if isLocalNonEmbedService(types.ServiceSourceRemote, types.ServiceChat) {
		t.Error("Remote service should not need queuing")
	}

	// 测试排队需求判断
	if !NeedsQueuing(types.ServiceSourceLocal, types.ServiceChat) {
		t.Error("Local chat service should need queuing")
	}

	if NeedsQueuing(types.ServiceSourceLocal, types.ServiceEmbed) {
		t.Error("Local embed service should not need queuing")
	}

	if NeedsQueuing(types.ServiceSourceRemote, types.ServiceChat) {
		t.Error("Remote service should not need queuing")
	}
}

func TestManager_BasicOperations(t *testing.T) {
	// 创建一个新的管理器实例（不使用全局单例）
	manager := &Manager{
		modelStates: make(map[string]bool),
	}

	// 先创建loader
	manager.loader = NewLoader(manager)

	// 创建queue时注入接口依赖
	manager.queue = NewQueue(manager, manager.loader)

	// 创建cleaner时注入接口依赖
	manager.cleaner = NewCleaner(manager.queue, manager.loader)

	// 启动管理器
	manager.Start(1 * time.Minute)
	defer manager.Stop()

	// 测试统计信息
	stats := manager.GetStats()
	if stats["queue"] == nil {
		t.Error("Expected queue stats to be present")
	}

	if stats["loader"] == nil {
		t.Error("Expected loader stats to be present")
	}

	if stats["cleaner"] == nil {
		t.Error("Expected cleaner stats to be present")
	}
}

func TestManager_ModelOperations(t *testing.T) {
	// 创建一个新的管理器实例
	manager := &Manager{
		modelStates: make(map[string]bool),
	}

	// 先创建loader
	manager.loader = NewLoader(manager)

	// 创建queue时注入接口依赖
	manager.queue = NewQueue(manager, manager.loader)

	// 创建cleaner时注入接口依赖
	manager.cleaner = NewCleaner(manager.queue, manager.loader)

	manager.Start(1 * time.Minute)
	defer manager.Stop()

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 测试模型加载
	err := manager.loader.EnsureModelLoaded(ctx, "test-model", mockProvider, "test-provider", "test-type")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// 测试模型状态查询
	modelState, exists := manager.GetModelState("test-model")
	if !exists {
		t.Error("Model should exist after loading")
	}

	if modelState.ModelName != "test-model" {
		t.Errorf("Expected model name to be 'test-model', got %s", modelState.ModelName)
	}

	// 测试标记模型为使用中
	err = manager.MarkModelInUse("test-model")
	if err != nil {
		t.Fatalf("Failed to mark model in use: %v", err)
	}

	if modelState.GetStatus() != ModelStatusInUse {
		t.Errorf("Expected model status to be in use, got %v", modelState.GetStatus())
	}

	// 测试标记模型为空闲
	err = manager.MarkModelIdle("test-model")
	if err != nil {
		t.Fatalf("Failed to mark model idle: %v", err)
	}

	if modelState.GetRefCount() != 0 {
		t.Errorf("Expected ref count to be 0, got %d", modelState.GetRefCount())
	}
}

func TestManager_EnqueueLocalModelRequest(t *testing.T) {
	// 创建一个新的管理器实例
	manager := &Manager{
		modelStates: make(map[string]bool),
	}

	// 先创建loader
	manager.loader = NewLoader(manager)

	// 创建queue时注入接口依赖
	manager.queue = NewQueue(manager, manager.loader)

	// 创建cleaner时注入接口依赖
	manager.cleaner = NewCleaner(manager.queue, manager.loader)

	manager.Start(1 * time.Minute)
	defer manager.Stop()

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 测试入队请求
	readyChan, errorChan, err := manager.EnqueueLocalModelRequest(ctx, "test-model", mockProvider, "test-provider", "test-type", 1)
	if err != nil {
		t.Fatalf("Failed to enqueue request: %v", err)
	}

	// 等待模型准备完成
	select {
	case <-readyChan:
		// 检查是否有错误
		select {
		case queueErr := <-errorChan:
			t.Fatalf("Queue processing failed: %v", queueErr)
		default:
			// 模型准备完成，无错误
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("Timeout waiting for model preparation")
	}

	// 等待处理
	time.Sleep(200 * time.Millisecond)

	// 完成请求
	manager.CompleteLocalModelRequest(1)

	// 等待完成
	time.Sleep(100 * time.Millisecond)

	// 验证模型已加载
	_, exists := manager.GetModelState("test-model")
	if !exists {
		t.Error("Model should exist after processing request")
	}
}

func TestCleaner_ModelUnloading(t *testing.T) {
	// 创建一个新的管理器实例，使用很短的空闲超时时间
	manager := &Manager{
		modelStates: make(map[string]bool),
	}

	// 先创建loader
	manager.loader = NewLoader(manager)

	// 创建queue时注入接口依赖
	manager.queue = NewQueue(manager, manager.loader)

	// 创建cleaner时注入接口依赖，设置很短的空闲超时
	manager.cleaner = NewCleaner(manager.queue, manager.loader)
	manager.cleaner.SetIdleTimeout(100 * time.Millisecond)

	manager.Start(50 * time.Millisecond) // 50ms清理间隔
	defer manager.Stop()

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 加载一个模型
	err := manager.loader.EnsureModelLoaded(ctx, "test-model", mockProvider, "test-provider", "test-type")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// 验证模型已加载
	_, exists := manager.GetModelState("test-model")
	if !exists {
		t.Error("Model should exist after loading")
	}

	// 验证模型在mock provider中已加载
	if !mockProvider.loadedModels["test-model"] {
		t.Error("Model should be loaded in mock provider")
	}

	// 等待超过空闲超时时间，让清理器运行
	time.Sleep(200 * time.Millisecond)

	// 验证模型已被清理（从管理器中移除）
	_, exists = manager.GetModelState("test-model")
	if exists {
		t.Error("Model should be cleaned up after idle timeout")
	}

	// 验证模型在mock provider中已被卸载
	if mockProvider.loadedModels["test-model"] {
		t.Error("Model should be unloaded from mock provider")
	}
}

func TestCleaner_ForceCleanup(t *testing.T) {
	// 创建一个新的管理器实例
	manager := &Manager{
		modelStates: make(map[string]bool),
	}

	// 先创建loader
	manager.loader = NewLoader(manager)

	// 创建queue时注入接口依赖
	manager.queue = NewQueue(manager, manager.loader)

	// 创建cleaner时注入接口依赖，设置很短的空闲超时
	manager.cleaner = NewCleaner(manager.queue, manager.loader)
	manager.cleaner.SetIdleTimeout(100 * time.Millisecond)

	manager.Start(1 * time.Hour) // 很长的清理间隔，不会自动清理
	defer manager.Stop()

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 加载一个模型
	err := manager.loader.EnsureModelLoaded(ctx, "test-model", mockProvider, "test-provider", "test-type")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// 等待超过空闲超时时间
	time.Sleep(150 * time.Millisecond)

	// 手动触发清理
	manager.cleaner.ForceCleanup()

	// 验证模型已被清理
	_, exists := manager.GetModelState("test-model")
	if exists {
		t.Error("Model should be cleaned up after force cleanup")
	}

	// 验证模型在mock provider中已被卸载
	if mockProvider.loadedModels["test-model"] {
		t.Error("Model should be unloaded from mock provider")
	}
}
