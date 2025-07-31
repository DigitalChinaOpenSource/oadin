package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"oadin/internal/logger"
	"oadin/internal/provider"
)

// QueuedRequest 排队的请求
type QueuedRequest struct {
	TaskID       uint64                        // 任务ID
	ModelName    string                        // 模型名称
	Provider     provider.ModelServiceProvider // 提供商实例
	ProviderName string                        // 提供商名称
	ProviderType string                        // 提供商类型
	Context      context.Context               // 上下文
	StartTime    time.Time                     // 开始时间
	CompleteChan chan struct{}                 // 完成通知通道
}

// Queue 模型请求队列
type Queue struct {
	queue        chan *QueuedRequest // 请求队列
	processing   bool                // 是否正在处理请求
	currentTask  *QueuedRequest      // 当前正在处理的任务
	mutex        sync.Mutex          // 互斥锁
	queueSize    int                 // 队列大小
	queueTimeout time.Duration       // 排队超时时间
	stopChan     chan struct{}       // 停止信号
	started      bool                // 是否已启动
}

// NewQueue 创建新的队列
func NewQueue(queueSize int, queueTimeout time.Duration) *Queue {
	return &Queue{
		queue:        make(chan *QueuedRequest, queueSize),
		queueSize:    queueSize,
		queueTimeout: queueTimeout,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动队列处理
func (q *Queue) Start() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.started {
		logger.LogicLogger.Warn("[Queue] Already started")
		return
	}

	q.started = true
	go q.processLoop()

	logger.LogicLogger.Info("[Queue] Started",
		"queue_size", q.queueSize,
		"queue_timeout", q.queueTimeout)
}

// Stop 停止队列处理
func (q *Queue) Stop() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.started {
		return
	}

	close(q.stopChan)
	q.started = false

	logger.LogicLogger.Info("[Queue] Stopped")
}

// EnqueueRequest 将请求加入队列（非阻塞）
func (q *Queue) EnqueueRequest(request *QueuedRequest) error {
	if !q.started {
		return fmt.Errorf("queue not started")
	}

	logger.LogicLogger.Debug("[Queue] Enqueueing request",
		"task_id", request.TaskID,
		"model", request.ModelName,
		"queue_length", len(q.queue))

	// 非阻塞加入队列
	select {
	case q.queue <- request:
		logger.LogicLogger.Debug("[Queue] Request enqueued successfully",
			"task_id", request.TaskID,
			"model", request.ModelName)
		return nil
	case <-time.After(q.queueTimeout):
		return fmt.Errorf("failed to enqueue request: queue full, timeout after %v", q.queueTimeout)
	case <-request.Context.Done():
		return request.Context.Err()
	}
}

// CompleteRequest 完成请求处理
func (q *Queue) CompleteRequest(taskID uint64) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	logger.LogicLogger.Debug("[Queue] Completing request", "task_id", taskID)

	// 检查当前处理的任务是否匹配
	if q.currentTask != nil && q.currentTask.TaskID == taskID {
		logger.LogicLogger.Debug("[Queue] Request completed",
			"task_id", taskID,
			"model", q.currentTask.ModelName,
			"processing_time", time.Since(q.currentTask.StartTime))

		// 安全地关闭channel（检查是否已经关闭）
		select {
		case <-q.currentTask.CompleteChan:
			// Channel已经关闭，不需要再关闭
			logger.LogicLogger.Debug("[Queue] CompleteChan already closed", "task_id", taskID)
		default:
			// Channel还没关闭，安全关闭它
			close(q.currentTask.CompleteChan)
		}

		q.currentTask = nil
		q.processing = false
	} else {
		logger.LogicLogger.Warn("[Queue] TaskID mismatch in CompleteRequest",
			"expected", func() uint64 {
				if q.currentTask != nil {
					return q.currentTask.TaskID
				}
				return 0
			}(),
			"actual", taskID)
	}
}

// processLoop 队列处理循环
func (q *Queue) processLoop() {
	logger.LogicLogger.Debug("[Queue] Process loop started")

	for {
		select {
		case request := <-q.queue:
			q.processRequest(request)
		case <-q.stopChan:
			logger.LogicLogger.Debug("[Queue] Process loop stopped")
			return
		}
	}
}

// processRequest 处理单个请求
func (q *Queue) processRequest(request *QueuedRequest) {
	q.mutex.Lock()
	q.processing = true
	q.currentTask = request
	q.mutex.Unlock()

	logger.LogicLogger.Info("[Queue] Processing request",
		"task_id", request.TaskID,
		"model", request.ModelName,
		"queue_wait_time", time.Since(request.StartTime))

	// 等待请求完成通知
	select {
	case <-request.CompleteChan:
		logger.LogicLogger.Debug("[Queue] Request processing completed",
			"task_id", request.TaskID,
			"model", request.ModelName)
	case <-request.Context.Done():
		logger.LogicLogger.Warn("[Queue] Request context cancelled",
			"task_id", request.TaskID,
			"model", request.ModelName,
			"error", request.Context.Err())

		// 清理当前任务状态
		q.mutex.Lock()
		q.currentTask = nil
		q.processing = false
		q.mutex.Unlock()
	case <-q.stopChan:
		logger.LogicLogger.Warn("[Queue] Queue stopped while processing request",
			"task_id", request.TaskID,
			"model", request.ModelName)
		return
	}
}

// GetStats 获取队列统计信息
func (q *Queue) GetStats() map[string]interface{} {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	stats := map[string]interface{}{
		"queue_size":    q.queueSize,
		"queue_length":  len(q.queue),
		"queue_timeout": q.queueTimeout.String(),
		"processing":    q.processing,
		"started":       q.started,
	}

	if q.currentTask != nil {
		stats["current_task"] = map[string]interface{}{
			"task_id":    q.currentTask.TaskID,
			"model_name": q.currentTask.ModelName,
			"start_time": q.currentTask.StartTime,
		}
	}

	return stats
}

// IsProcessing 检查是否正在处理请求
func (q *Queue) IsProcessing() bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.processing
}

// GetQueueLength 获取当前队列长度
func (q *Queue) GetQueueLength() int {
	return len(q.queue)
}
