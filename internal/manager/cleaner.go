package manager

import (
	"sync"
	"time"

	"oadin/internal/logger"
)

// Cleaner 模型自动清理服务
// 负责定期检查并卸载空闲超时的模型
type Cleaner struct {
	stateManager  ModelStateManager
	queueChecker  QueueStatusChecker
	interval      time.Duration // 清理检查间隔
	cleanupTicker *time.Ticker  // 清理定时器
	stopChan      chan struct{} // 停止信号
	started       bool          // 是否已启动
	mutex         sync.Mutex    // 互斥锁
}

// NewCleaner 创建新的清理器
func NewCleaner(stateManager ModelStateManager, queueChecker QueueStatusChecker) *Cleaner {
	return &Cleaner{
		stateManager: stateManager,
		queueChecker: queueChecker,
		interval:     5 * time.Minute, // 默认5分钟清理一次
		stopChan:     make(chan struct{}),
	}
}

// Start 启动清理服务
func (c *Cleaner) Start(cleanupInterval time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.started {
		logger.LogicLogger.Warn("[Cleaner] Already started")
		return
	}

	if cleanupInterval > 0 {
		c.interval = cleanupInterval
	}
	c.cleanupTicker = time.NewTicker(c.interval)
	c.started = true

	// 启动后台清理goroutine
	go c.cleanupLoop()

	logger.LogicLogger.Info("[Cleaner] Started",
		"cleanup_interval", c.interval)
}

// Stop 停止清理服务
func (c *Cleaner) Stop() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.started {
		return
	}

	close(c.stopChan)
	if c.cleanupTicker != nil {
		c.cleanupTicker.Stop()
	}
	c.started = false

	logger.LogicLogger.Info("[Cleaner] Stopped")
}

// SetIdleTimeout 设置空闲超时时间（保持向后兼容）
func (c *Cleaner) SetIdleTimeout(timeout time.Duration) {
	logger.LogicLogger.Info("[Cleaner] Set idle timeout", "timeout", timeout)
	// 简化实现，不再存储idle timeout
}

// GetIdleTimeout 获取空闲超时时间（保持向后兼容）
func (c *Cleaner) GetIdleTimeout() time.Duration {
	return 10 * time.Minute // 返回默认值
}

// cleanupLoop 后台清理循环
func (c *Cleaner) cleanupLoop() {
	logger.LogicLogger.Debug("[Cleaner] Cleanup loop started")

	for {
		select {
		case <-c.cleanupTicker.C:
			c.performCleanup()
		case <-c.stopChan:
			logger.LogicLogger.Debug("[Cleaner] Cleanup loop stopped")
			return
		}
	}
}

// performCleanup 执行清理操作
func (c *Cleaner) performCleanup() {
	// 如果没有请求在处理且有模型加载
	if !c.queueChecker.HasActiveRequests() {
		currentModel := c.stateManager.GetCurrentModel()
		if currentModel != "" {
			// 检查空闲时间
			if c.shouldUnloadModel(currentModel) {
				logger.LogicLogger.Info("[Cleaner] Unloading idle model",
					"model", currentModel)
				c.unloadIdleModel(currentModel)
			}
		}
	}
}

// shouldUnloadModel 检查是否应该卸载模型（简化实现）
func (c *Cleaner) shouldUnloadModel(modelName string) bool {
	// 简化实现：总是返回false，避免意外卸载
	// 实际项目中可以根据需要实现更复杂的逻辑
	return false
}

// unloadIdleModel 卸载空闲模型（简化实现）
func (c *Cleaner) unloadIdleModel(modelName string) {
	// 简化实现：只记录日志
	logger.LogicLogger.Info("[Cleaner] Would unload idle model", "model", modelName)
	// 实际项目中可以调用相应的卸载逻辑
}

// ForceCleanup 强制执行一次清理
func (c *Cleaner) ForceCleanup() {
	logger.LogicLogger.Info("[Cleaner] Force cleanup triggered")
	c.performCleanup()
}

// GetStats 获取清理器统计信息
func (c *Cleaner) GetStats() map[string]interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	stats := map[string]interface{}{
		"cleanup_interval": c.interval.String(),
		"started":          c.started,
	}

	if c.cleanupTicker != nil {
		stats["has_ticker"] = true
	} else {
		stats["has_ticker"] = false
	}

	return stats
}

// IsStarted 检查清理服务是否已启动
func (c *Cleaner) IsStarted() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.started
}
