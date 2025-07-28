package model_engine

import (
	"context"
	"oadin/extension/api/dto"
	"oadin/internal/types"
)

// EngineServiceInterface 定义了引擎服务的核心接口
type EngineServiceInterface interface {
	// Chat 处理非流式聊天请求
	Chat(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error)

	// ChatStream 处理流式聊天请求
	ChatStream(ctx context.Context, req *dto.ChatRequest) (<-chan *dto.ChatResponse, <-chan error)

	// GenerateEmbedding 生成向量嵌入
	GenerateEmbedding(ctx context.Context, req *dto.EmbeddingRequest) (*dto.EmbeddingResponse, error)
}

// ModelManager 定义了模型管理的接口
type ModelManager interface {
	// GetModelById 根据ID获取模型信息
	GetModelById(ctx context.Context, modelId string) *types.SupportModel

	// RestoreCurrentModel 恢复当前模型状态
	RestoreCurrentModel(ctx context.Context, modelId string) error
}

// EngineServiceProvider 组合接口，提供完整的引擎服务能力
type EngineServiceProvider interface {
	EngineServiceInterface
	ModelManager
}
