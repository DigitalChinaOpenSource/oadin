package manager

import (
	"sync"
	"time"

	"oadin/internal/logger"
)

// Cleaner 模型自动清理服务
// 负责定期检查并卸载空闲超时的模型
type Cleaner struct {
	loader          *Loader       // 模型加载器
	idleTimeout     time.Duration // 空闲超时时间
	cleanupInterval time.Duration // 清理检查间隔
	cleanupTicker   *time.Ticker  // 清理定时器
	stopChan        chan struct{} // 停止信号
	started         bool          // 是否已启动
	mutex           sync.Mutex    // 互斥锁
}

// NewCleaner 创建新的清理器
func NewCleaner(loader *Loader, idleTimeout time.Duration) *Cleaner {
	return &Cleaner{
		loader:      loader,
		idleTimeout: idleTimeout,
		stopChan:    make(chan struct{}),
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

	c.cleanupInterval = cleanupInterval
	c.cleanupTicker = time.NewTicker(cleanupInterval)
	c.started = true

	// 启动后台清理goroutine
	go c.cleanupLoop()

	logger.LogicLogger.Info("[Cleaner] Started",
		"cleanup_interval", cleanupInterval,
		"idle_timeout", c.idleTimeout)
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

// SetIdleTimeout 设置空闲超时时间
func (c *Cleaner) SetIdleTimeout(timeout time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.idleTimeout = timeout
	logger.LogicLogger.Info("[Cleaner] Set idle timeout", "timeout", timeout)
}

// GetIdleTimeout 获取空闲超时时间
func (c *Cleaner) GetIdleTimeout() time.Duration {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.idleTimeout
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
// 检查所有模型，找出空闲超时的模型并调用provider的UnloadModel方法卸载
func (c *Cleaner) performCleanup() {
	c.mutex.Lock()
	idleTimeout := c.idleTimeout
	c.mutex.Unlock()

	logger.LogicLogger.Debug("[Cleaner] Starting cleanup check",
		"idle_timeout", idleTimeout)

	// 获取需要清理的空闲模型
	idleModels := c.loader.GetIdleModels(idleTimeout)

	if len(idleModels) == 0 {
		logger.LogicLogger.Debug("[Cleaner] No idle models to clean up")
		return
	}

	logger.LogicLogger.Info("[Cleaner] Found idle models to clean up",
		"count", len(idleModels))

	// 逐个卸载空闲模型
	cleanedCount := 0
	for _, modelState := range idleModels {
		idleTime := time.Since(modelState.LastUsedTime)

		logger.LogicLogger.Info("[Cleaner] Attempting to unload idle model",
			"model", modelState.ModelName,
			"idle_time", idleTime,
			"status", modelState.GetStatus().String(),
			"ref_count", modelState.GetRefCount())

		// 调用loader的ForceUnloadModel，它会调用provider的UnloadModel方法
		if err := c.loader.ForceUnloadModel(modelState.ModelName); err != nil {
			logger.LogicLogger.Error("[Cleaner] Failed to unload idle model",
				"model", modelState.ModelName,
				"idle_time", idleTime,
				"error", err)
		} else {
			cleanedCount++
			logger.LogicLogger.Info("[Cleaner] Successfully unloaded idle model",
				"model", modelState.ModelName,
				"idle_time", idleTime)
		}
	}

	logger.LogicLogger.Info("[Cleaner] Cleanup completed",
		"total_checked", len(idleModels),
		"successfully_cleaned", cleanedCount)
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
		"idle_timeout":     c.idleTimeout.String(),
		"cleanup_interval": c.cleanupInterval.String(),
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
