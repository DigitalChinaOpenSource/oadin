package provider

import (
	"context"

	"intel.com/aog/internal/provider/engine"
	"intel.com/aog/internal/types"
)

// ModelServiceProvider local model engine
type ModelServiceProvider interface {
	InstallEngine() error
	StartEngine(mode string) error
	StopEngine() error
	HealthCheck() error
	InitEnv() error
	PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error)
	PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error)
	DeleteModel(ctx context.Context, req *types.DeleteRequest) error
	ListModels(ctx context.Context) (*types.ListResponse, error)
	GetConfig() *types.EngineRecommendConfig
	GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error)
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
