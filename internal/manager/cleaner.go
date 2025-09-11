package manager

import (
	"sync"
	"time"

	"oadin/config"
	"oadin/internal/logger"
)

// Cleaner 模型自动清理服务
// 负责定期检查并卸载空闲超时的模型
type Cleaner struct {
	queueChecker  QueueStatusChecker // 队列状态检查器
	modelLoader   ModelLoader        // 模型加载器，用于实际卸载操作
	interval      time.Duration      // 清理检查间隔
	idleTimeout   time.Duration      // 空闲超时时间
	cleanupTicker *time.Ticker       // 清理定时器
	stopChan      chan struct{}      // 停止信号
	started       bool               // 是否已启动
	mutex         sync.Mutex         // 互斥锁
}

// NewCleaner 创建新的清理器
func NewCleaner(queueChecker QueueStatusChecker, modelLoader ModelLoader) *Cleaner {
	// 获取默认的清理间隔和空闲超时时间
	defaultInterval := 1 * time.Minute    // 默认清理间隔
	defaultIdleTimeout := 5 * time.Minute // 默认空闲超时

	// 如果全局配置已初始化，使用配置中的值
	if config.GlobalEnvironment != nil {
		if config.GlobalEnvironment.ModelCleanupInterval > 0 {
			defaultInterval = config.GlobalEnvironment.ModelCleanupInterval
		}
		if config.GlobalEnvironment.ModelIdleTimeout > 0 {
			defaultIdleTimeout = config.GlobalEnvironment.ModelIdleTimeout
		}
	}

	return &Cleaner{
		queueChecker: queueChecker,
		modelLoader:  modelLoader,
		interval:     defaultInterval,
		idleTimeout:  defaultIdleTimeout,
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
func (c *Cleaner) performCleanup() {
	// 如果有活跃请求正在处理，跳过清理
	if c.queueChecker.HasActiveRequests() {
		logger.LogicLogger.Debug("[Cleaner] Skipping cleanup - active requests in progress")
		return
	}

	// 获取当前空闲超时时间
	idleTimeout := c.GetIdleTimeout()

	// 获取所有空闲超时的模型
	idleModels := c.getIdleModels(idleTimeout)

	if len(idleModels) == 0 {
		logger.LogicLogger.Debug("[Cleaner] No idle models found for cleanup")
		return
	}

	logger.LogicLogger.Info("[Cleaner] Found idle models for cleanup",
		"count", len(idleModels), "idle_timeout", idleTimeout)

	// 卸载空闲模型
	for _, modelState := range idleModels {
		c.unloadIdleModel(modelState)
	}
}

// getIdleModels 获取所有空闲超时的模型
func (c *Cleaner) getIdleModels(idleTimeout time.Duration) []*ModelState {
	// 通过ModelLoader接口获取空闲模型
	if loader, ok := c.modelLoader.(interface {
		GetIdleModels(time.Duration) []*ModelState
	}); ok {
		return loader.GetIdleModels(idleTimeout)
	}

	logger.LogicLogger.Warn("[Cleaner] ModelLoader does not support GetIdleModels method")
	return nil
}

// unloadIdleModel 卸载空闲模型
func (c *Cleaner) unloadIdleModel(modelState *ModelState) {
	if modelState == nil {
		logger.LogicLogger.Warn("[Cleaner] Cannot unload nil model state")
		return
	}

	logger.LogicLogger.Info("[Cleaner] Unloading idle model",
		"model", modelState.ModelName,
		"idle_time", time.Since(modelState.LastUsedTime),
		"status", modelState.GetStatus())

	// 调用loader的ForceUnloadModel方法
	err := c.modelLoader.ForceUnloadModel(modelState.ModelName)
	if err != nil {
		logger.LogicLogger.Error("[Cleaner] Failed to unload idle model",
			"model", modelState.ModelName, "error", err)
		return
	}

	logger.LogicLogger.Info("[Cleaner] Successfully unloaded idle model",
		"model", modelState.ModelName)
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

	return map[string]interface{}{
		"cleanup_interval": c.interval.String(),
		"idle_timeout":     c.idleTimeout.String(),
		"started":          c.started,
	}
}

// IsStarted 检查清理服务是否已启动
func (c *Cleaner) IsStarted() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.started
}
