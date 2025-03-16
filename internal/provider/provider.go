package provider

import (
	"context"
	"log/slog"

	"intel.com/aog/internal/provider/engine"
	"intel.com/aog/internal/schedule"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils/client"
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
	DeleteModel(ctx context.Context, req *types.DeleteRequest) error
	ListModels(ctx context.Context) (*types.ListResponse, error)
	GetConfig() *types.EngineRecommendConfig
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

type ServiceProviderInfo struct {
	AuthType string
	Endpoint string
}

func GetServiceProviderInfo(flavor string) ServiceProviderInfo {
	def, err := schedule.LoadFlavorDef(flavor, "/")
	if err != nil {
		slog.Error("[Provider]Failed to load file", "provider_name", flavor, "error", err.Error())
		return ServiceProviderInfo{AuthType: "none", Endpoint: "none"}
	}
	return ServiceProviderInfo{
		AuthType: def.AuthType,
		Endpoint: def.Endpoint,
	}
}
