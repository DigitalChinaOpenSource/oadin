//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package client

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
)

// WebSocketConnection 表示单个WebSocket连接
type WebSocketConnection struct {
	ID        string
	Conn      *websocket.Conn
	TaskID    uint64
	Service   string
	Flavor    string
	CreatedAt time.Time
	mu        sync.Mutex // 用于保护TaskID等字段的并发访问

	// 最后处理的任务ID
	LastTaskID uint64

	// 活跃的任务ID列表，用于跟踪未完成的任务
	ActiveTasks map[uint64]bool

	// 会话数据
	SessionData *WebSocketSessionData
}

// WebSocketSessionData 存储WebSocket会话数据
type WebSocketSessionData struct {
	// 任务映射，key为msgTaskID，value为任务信息
	Tasks map[uint64]*TaskInfo

	// 服务特定数据
	STTParams *types.SpeechToTextParams // 语音识别参数
}

// TaskInfo 存储每个任务的基本信息
type TaskInfo struct {
	TaskType    string // 任务类型
	TaskStarted bool   // 任务是否已启动
	StartTime   int64  // 任务开始时间
	EndTime     int64  // 任务结束时间
}

// WebSocketManager 管理所有活跃的WebSocket连接
type WebSocketManager struct {
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex
}

// 全局WebSocket连接管理器
var wsManager = &WebSocketManager{
	connections: make(map[string]*WebSocketConnection),
}

// NewWebSocketConnection 创建新的WebSocket连接
func NewWebSocketConnection(conn *websocket.Conn, taskID uint64, flavor, service string) *WebSocketConnection {
	return &WebSocketConnection{
		ID:          uuid.New().String(),
		Conn:        conn,
		TaskID:      taskID,
		LastTaskID:  0,
		Service:     service,
		Flavor:      flavor,
		CreatedAt:   time.Now(),
		ActiveTasks: make(map[uint64]bool),
		SessionData: &WebSocketSessionData{
			Tasks:     make(map[uint64]*TaskInfo),
			STTParams: types.NewSpeechToTextParams(),
		},
	}
}

// RegisterConnection 注册一个新的WebSocket连接
func (m *WebSocketManager) RegisterConnection(conn *websocket.Conn, taskID uint64, flavor, service string) *WebSocketConnection {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建新的WebSocket连接
	wsConn := NewWebSocketConnection(conn, taskID, flavor, service)

	// 存储连接
	m.connections[wsConn.ID] = wsConn
	logger.LogicLogger.Info("[WebSocketManager] Registered new connection",
		"connID", wsConn.ID,
		"taskID", taskID,
		"flavor", flavor,
		"service", service)

	return wsConn
}

// UnregisterConnection 注销一个WebSocket连接
func (m *WebSocketManager) UnregisterConnection(connID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if conn, exists := m.connections[connID]; exists {
		logger.LogicLogger.Info("[WebSocketManager] Unregistered connection",
			"connID", connID,
			"taskID", conn.TaskID,
			"flavor", conn.Flavor,
			"service", conn.Service)
		delete(m.connections, connID)
	}
}

// GetConnection 获取指定ID的连接
func (m *WebSocketManager) GetConnection(connID string) (*WebSocketConnection, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[connID]
	return conn, exists
}

// GetAllConnections 获取所有活跃连接
func (m *WebSocketManager) GetAllConnections() []*WebSocketConnection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conns := make([]*WebSocketConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}
	return conns
}

// CloseAllConnections 关闭所有连接
func (m *WebSocketManager) CloseAllConnections() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, conn := range m.connections {
		conn.Conn.Close()
		delete(m.connections, id)
	}
}

// GetActiveConnectionCount 获取活跃连接数
func (m *WebSocketManager) GetActiveConnectionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.connections)
}

// GetGlobalWebSocketManager 获取全局WebSocket管理器实例
func GetGlobalWebSocketManager() *WebSocketManager {
	return wsManager
}

// WriteJSON 安全地向连接写入JSON数据
func (c *WebSocketConnection) WriteJSON(v interface{}) error {
	return c.Conn.WriteJSON(v)
}

// WriteMessage 安全地向连接写入消息
func (c *WebSocketConnection) WriteMessage(messageType int, data []byte) error {
	return c.Conn.WriteMessage(messageType, data)
}

// Close 关闭连接并从管理器中注销
func (c *WebSocketConnection) Close() {
	// 先关闭关联的GRPC流（如果有）
	if c.ID != "" {
		grpcStreamManager := GetGlobalGRPCStreamManager()
		grpcStreamManager.CloseSessionByWSConnID(c.ID)
		logger.LogicLogger.Info("[WebSocketConnection] Closed associated GRPC stream", "connID", c.ID)
	}

	wsManager.UnregisterConnection(c.ID)
	c.Conn.Close()
}

// GetTaskType 获取指定taskID的任务类型
func (c *WebSocketConnection) GetTaskType(taskID uint64) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if taskInfo, exists := c.SessionData.Tasks[taskID]; exists {
		return taskInfo.TaskType
	}
	return ""
}

// SetTaskType 设置指定taskID的任务类型
func (c *WebSocketConnection) SetTaskType(taskID uint64, taskType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if taskInfo, exists := c.SessionData.Tasks[taskID]; exists {
		taskInfo.TaskType = taskType
	} else {
		c.SessionData.Tasks[taskID] = &TaskInfo{
			TaskType: taskType,
		}
	}

	// 更新最后处理的任务ID
	c.LastTaskID = taskID
}

// GetSTTParams 获取语音识别参数
func (c *WebSocketConnection) GetSTTParams() *types.SpeechToTextParams {
	return c.SessionData.STTParams
}

// SetTaskStatus 设置指定taskID的任务状态（启动/结束）
func (c *WebSocketConnection) SetTaskStatus(taskID uint64, started bool, timestamp int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	taskInfo, exists := c.SessionData.Tasks[taskID]
	if !exists {
		taskInfo = &TaskInfo{}
		c.SessionData.Tasks[taskID] = taskInfo
	}

	taskInfo.TaskStarted = started
	if started {
		taskInfo.StartTime = timestamp
	} else {
		taskInfo.EndTime = timestamp
	}

	// 更新最后处理的任务ID
	c.LastTaskID = taskID
}

// SetConnectionTaskStatus 设置连接基础任务状态（向后兼容）
func (c *WebSocketConnection) SetConnectionTaskStatus(started bool, timestamp int64) {
	if c.LastTaskID > 0 {
		c.SetTaskStatus(c.LastTaskID, started, timestamp)
	} else if c.TaskID > 0 {
		c.SetTaskStatus(c.TaskID, started, timestamp)
	}
}

// IsTaskStarted 检查指定taskID的任务是否已启动
func (c *WebSocketConnection) IsTaskStarted(taskID uint64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if taskInfo, exists := c.SessionData.Tasks[taskID]; exists {
		return taskInfo.TaskStarted
	}
	return false
}

// GetTaskTimes 获取指定taskID的任务开始和结束时间
func (c *WebSocketConnection) GetTaskTimes(taskID uint64) (startTime, endTime int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if taskInfo, exists := c.SessionData.Tasks[taskID]; exists {
		return taskInfo.StartTime, taskInfo.EndTime
	}
	return 0, 0
}

// SetTaskFinished 设置指定taskID的任务完成状态
func (c *WebSocketConnection) SetTaskFinished(taskID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if taskInfo, exists := c.SessionData.Tasks[taskID]; exists {
		taskInfo.TaskStarted = false
		taskInfo.EndTime = time.Now().Unix()
	}

	// 更新最后处理的任务ID
	c.LastTaskID = taskID
}

// SetConnectionTaskFinished 设置连接基础任务完成状态（向后兼容）
func (c *WebSocketConnection) SetConnectionTaskFinished() {
	if c.LastTaskID > 0 {
		c.SetTaskFinished(c.LastTaskID)
	} else if c.TaskID > 0 {
		c.SetTaskFinished(c.TaskID)
	}
}

// AddActiveTask 添加一个活跃任务到跟踪列表
func (c *WebSocketConnection) AddActiveTask(taskID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActiveTasks[taskID] = true
	logger.LogicLogger.Debug("[WebSocketConnection] Added active task",
		"connID", c.ID,
		"taskID", taskID,
		"activeCount", len(c.ActiveTasks))
}

// RemoveActiveTask 从跟踪列表中移除一个活跃任务
func (c *WebSocketConnection) RemoveActiveTask(taskID uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.ActiveTasks, taskID)
	logger.LogicLogger.Debug("[WebSocketConnection] Removed active task",
		"connID", c.ID,
		"taskID", taskID,
		"activeCount", len(c.ActiveTasks))
}

// HasActiveTasks 检查是否有任何活跃的任务
func (c *WebSocketConnection) HasActiveTasks() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.ActiveTasks) > 0
}

// GetActiveTaskCount 获取当前活跃任务数量
func (c *WebSocketConnection) GetActiveTaskCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.ActiveTasks)
}

// ClearActiveTasks 清除所有活跃任务
func (c *WebSocketConnection) ClearActiveTasks() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ActiveTasks = make(map[uint64]bool)
	logger.LogicLogger.Debug("[WebSocketConnection] Cleared all active tasks", "connID", c.ID)
}
