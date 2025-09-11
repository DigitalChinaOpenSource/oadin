package manager

import (
	"context"
	"testing"
	"time"
)

// MockStateManager 模拟状态管理器
type MockStateManager struct {
	currentModel string
}

func (m *MockStateManager) GetCurrentModel() string {
	return m.currentModel
}

func (m *MockStateManager) SetCurrentModel(modelName string) {
	m.currentModel = modelName
}

func (m *MockStateManager) MarkModelInUse(modelName string) error {
	return nil
}

func (m *MockStateManager) MarkModelIdle(modelName string) error {
	return nil
}

// MockQueueChecker 模拟队列检查器
type MockQueueChecker struct {
	hasActiveRequests bool
}

func (m *MockQueueChecker) HasPendingRequests() bool {
	return false
}

func (m *MockQueueChecker) HasActiveRequests() bool {
	return m.hasActiveRequests
}

func (m *MockQueueChecker) GetCurrentRequest() *QueuedRequest {
	return nil
}

func TestCleaner_BasicOperations(t *testing.T) {
	// 创建mock状态管理器
	stateManager := &MockStateManager{}

	// 创建mock队列检查器
	queueChecker := &MockQueueChecker{}

	// 创建loader
	loader := NewLoader(stateManager)

	// 创建cleaner
	cleaner := NewCleaner(queueChecker, loader)

	// 测试初始状态
	if cleaner.IsStarted() {
		t.Error("Cleaner should not be started initially")
	}

	// 启动清理器
	cleaner.Start(1 * time.Minute)
	defer cleaner.Stop()

	// 测试启动状态
	if !cleaner.IsStarted() {
		t.Error("Cleaner should be started after Start()")
	}

	// 测试统计信息
	stats := cleaner.GetStats()
	if stats["idle_timeout"] == nil {
		t.Error("Expected idle_timeout in stats")
	}

	if stats["cleanup_interval"] == nil {
		t.Error("Expected cleanup_interval in stats")
	}

	if stats["started"] != true {
		t.Error("Expected started to be true")
	}
}

func TestCleaner_ModelCleanup(t *testing.T) {
	// 创建mock状态管理器
	stateManager := &MockStateManager{}

	// 创建mock队列检查器
	queueChecker := &MockQueueChecker{}

	// 创建loader
	loader := NewLoader(stateManager)

	// 创建cleaner，设置很短的空闲超时
	cleaner := NewCleaner(queueChecker, loader)
	cleaner.SetIdleTimeout(100 * time.Millisecond)

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 加载一个模型
	err := loader.EnsureModelLoaded(ctx, "test-model", mockProvider, "test-provider", "test-type")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// 验证模型已加载
	_, exists := loader.GetModelState("test-model")
	if !exists {
		t.Error("Model should exist after loading")
	}

	// 验证模型在mock provider中已加载
	if !mockProvider.loadedModels["test-model"] {
		t.Error("Model should be loaded in mock provider")
	}

	// 等待超过空闲超时时间
	time.Sleep(150 * time.Millisecond)

	// 手动触发清理
	cleaner.ForceCleanup()

	// 验证模型已被清理
	_, exists = loader.GetModelState("test-model")
	if exists {
		t.Error("Model should be cleaned up after force cleanup")
	}

	// 验证模型在mock provider中已被卸载
	if mockProvider.loadedModels["test-model"] {
		t.Error("Model should be unloaded from mock provider")
	}
}

func TestCleaner_NoCleanupWhenInUse(t *testing.T) {
	// 创建mock状态管理器
	stateManager := &MockStateManager{}

	// 创建mock队列检查器
	queueChecker := &MockQueueChecker{}

	// 创建loader
	loader := NewLoader(stateManager)

	// 创建cleaner，设置很短的空闲超时
	cleaner := NewCleaner(queueChecker, loader)
	cleaner.SetIdleTimeout(100 * time.Millisecond)

	mockProvider := NewMockProvider()
	ctx := context.Background()

	// 加载一个模型
	err := loader.EnsureModelLoaded(ctx, "test-model", mockProvider, "test-provider", "test-type")
	if err != nil {
		t.Fatalf("Failed to load model: %v", err)
	}

	// 标记模型为使用中
	err = loader.MarkModelInUse("test-model")
	if err != nil {
		t.Fatalf("Failed to mark model in use: %v", err)
	}

	// 等待超过空闲超时时间
	time.Sleep(150 * time.Millisecond)

	// 手动触发清理
	cleaner.ForceCleanup()

	// 验证模型没有被清理（因为正在使用中）
	_, exists := loader.GetModelState("test-model")
	if !exists {
		t.Error("Model should not be cleaned up when in use")
	}

	// 验证模型在mock provider中仍然加载
	if !mockProvider.loadedModels["test-model"] {
		t.Error("Model should still be loaded in mock provider when in use")
	}

	// 标记模型为空闲
	err = loader.MarkModelIdle("test-model")
	if err != nil {
		t.Fatalf("Failed to mark model idle: %v", err)
	}

	// 等待超过空闲超时时间
	time.Sleep(150 * time.Millisecond)

	// 再次触发清理
	cleaner.ForceCleanup()

	// 现在模型应该被清理
	_, exists = loader.GetModelState("test-model")
	if exists {
		t.Error("Model should be cleaned up after marking idle")
	}
}

func TestCleaner_SetIdleTimeout(t *testing.T) {
	// 创建mock状态管理器
	stateManager := &MockStateManager{}

	// 创建mock队列检查器
	queueChecker := &MockQueueChecker{}

	// 创建loader
	loader := NewLoader(stateManager)

	// 创建cleaner
	cleaner := NewCleaner(queueChecker, loader)

	// 测试设置空闲超时时间
	newTimeout := 10 * time.Minute
	cleaner.SetIdleTimeout(newTimeout)

	if cleaner.GetIdleTimeout() != newTimeout {
		t.Errorf("Expected idle timeout to be %v, got %v", newTimeout, cleaner.GetIdleTimeout())
	}

	// 验证统计信息中的超时时间也更新了
	stats := cleaner.GetStats()
	if stats["idle_timeout"] != newTimeout.String() {
		t.Errorf("Expected idle_timeout in stats to be %v, got %v", newTimeout.String(), stats["idle_timeout"])
	}
}
