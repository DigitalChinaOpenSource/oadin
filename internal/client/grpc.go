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
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"intel.com/aog/internal/client/grpc/grpc_client"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/utils"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func NewGRPCClient(target string) (*GRPCClient, error) {
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &GRPCClient{conn: conn},
		nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *GRPCClient) ServerLive() (*grpc_client.ServerLiveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	serverLiveRequest := grpc_client.ServerLiveRequest{}
	serverLiveResponse, err := client.ServerLive(ctx, &serverLiveRequest)
	if err != nil {
		return nil, err
	}
	return serverLiveResponse, nil
}

func (c *GRPCClient) ServerReady() (*grpc_client.ServerReadyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	serverReadyRequest := grpc_client.ServerReadyRequest{}
	serverReadyResponse, err := client.ServerReady(ctx, &serverReadyRequest)
	if err != nil {
		return nil, err
	}
	return serverReadyResponse, nil
}

func (c *GRPCClient) ModelMetadata(modelName string, modelVersion string) (*grpc_client.ModelMetadataResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := grpc_client.NewGRPCInferenceServiceClient(c.conn)
	modelMetadataRequest := grpc_client.ModelMetadataRequest{
		Name:    modelName,
		Version: modelVersion,
	}
	modelMetadataResponse, err := client.ModelMetadata(ctx, &modelMetadataRequest)
	if err != nil {
		return nil, err
	}
	return modelMetadataResponse, nil
}

// GRPCStreamSession 表示单个GRPC流会话
type GRPCStreamSession struct {
	WSConnID   string                                                  // 关联的WebSocket连接ID
	Client     grpc_client.GRPCInferenceServiceClient                  // GRPC客户端
	Stream     grpc_client.GRPCInferenceService_ModelStreamInferClient // 双向流
	Context    context.Context                                         // 流的上下文
	CancelFunc context.CancelFunc                                      // 取消函数

	ResponseChannel chan interface{}
	Done            chan struct{} // 结束信号
	Active          bool          // 会话是否活跃
	mutex           sync.Mutex    // 同步锁

	Service   string // 服务名称
	Model     string // 模型名称
	CreatedAt int64  // 创建时间戳
}

// GRPCStreamManager 管理所有活跃的GRPC流会话
type GRPCStreamManager struct {
	sessions map[string]*GRPCStreamSession // 以WebSocket连接ID为键
	mutex    sync.RWMutex
}

// 全局GRPC流管理器
var grpcStreamManager = &GRPCStreamManager{
	sessions: make(map[string]*GRPCStreamSession),
}

// CreateSession 创建新的GRPC流会话
func (m *GRPCStreamManager) CreateSession(
	wsConnID string,
	client grpc_client.GRPCInferenceServiceClient,
	stream grpc_client.GRPCInferenceService_ModelStreamInferClient,
	ctx context.Context,
	cancelFunc context.CancelFunc,
	service string,
	model string,
) *GRPCStreamSession {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果已存在会话，先关闭它
	if existing, exists := m.sessions[wsConnID]; exists {
		existing.CancelFunc()
		logger.LogicLogger.Info("[GRPCStreamManager] Closed existing session for reconnection",
			"wsConnID", wsConnID)
	}

	session := &GRPCStreamSession{
		WSConnID:   wsConnID,
		Client:     client,
		Stream:     stream,
		Context:    ctx,
		CancelFunc: cancelFunc,
		Service:    service,
		Model:      model,
		CreatedAt:  utils.NowUnixMilli(),
	}

	m.sessions[wsConnID] = session
	logger.LogicLogger.Info("[GRPCStreamManager] Created new session",
		"wsConnID", wsConnID,
		"service", service,
		"model", model)

	return session
}

// GetSessionByWSConnID 通过WebSocket连接ID获取会话
func (m *GRPCStreamManager) GetSessionByWSConnID(wsConnID string) *GRPCStreamSession {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[wsConnID]
	if !exists {
		return nil
	}

	// 检查会话是否仍然有效
	select {
	case <-session.Context.Done():
		// 会话已关闭
		logger.LogicLogger.Info("[GRPCStreamManager] Found expired session",
			"wsConnID", wsConnID)
		return nil
	default:
		// 会话有效
		return session
	}
}

// CloseSessionByWSConnID 关闭并删除指定WebSocket连接ID的会话
func (m *GRPCStreamManager) CloseSessionByWSConnID(wsConnID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if session, exists := m.sessions[wsConnID]; exists {
		session.CancelFunc()
		delete(m.sessions, wsConnID)
		logger.LogicLogger.Info("[GRPCStreamManager] Closed session",
			"wsConnID", wsConnID)
	}
}

// CloseAllSessions 关闭所有会话
func (m *GRPCStreamManager) CloseAllSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for wsConnID, session := range m.sessions {
		session.CancelFunc()
		delete(m.sessions, wsConnID)
	}

	logger.LogicLogger.Info("[GRPCStreamManager] Closed all sessions")
}

// GetActiveSessionCount 获取活跃会话数
func (m *GRPCStreamManager) GetActiveSessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

// GetGlobalGRPCStreamManager 获取全局GRPC流管理器实例
func GetGlobalGRPCStreamManager() *GRPCStreamManager {
	return grpcStreamManager
}
