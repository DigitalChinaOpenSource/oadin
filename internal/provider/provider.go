package provider

import (
	"context"

	"byze/internal/provider/engine"
	"byze/internal/types"
	"byze/internal/utils/client"
)

// ModelServiceProvider local model engine
type ModelServiceProvider interface {
	GetDefaultClient() *client.Client
	InstallEngine() error
	StartEngine() error
	StopEngine() error
	HealthCheck() error
	InitEnv() error
	PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error)
	PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error)
	DeleteModel(ctx context.Context, req *types.DeleteRequest) error
	ListModels(ctx context.Context) (*types.ListResponse, error)
	GetConfig() *types.EngineRecommendConfig
	GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error)
	CopyModel(ctx context.Context, req *types.CopyModelRequest) error
	GetRunModels(ctx context.Context) (*types.ListResponse, error)

	Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error)
	ChatStream(ctx context.Context, req *types.ChatRequest) (chan *types.ChatResponse, chan error)
	GenerateEmbedding(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error)
}

func GetModelEngine(engineName string) ModelServiceProvider {
	switch engineName {
	case "ollama":
		return engine.NewOllamaProvider(nil)
	case "openvino":
		// todo
		return engine.NewOpenvinoProvider(nil)
	default:
		return engine.NewOllamaProvider(nil)
	}
}
