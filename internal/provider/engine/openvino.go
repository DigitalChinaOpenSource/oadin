package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"byze/internal/types"
	"byze/internal/utils/client"
)

type OpenvinoProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func NewOpenvinoProvider(config *types.EngineRecommendConfig) *OpenvinoProvider {
	return nil
}

func (o *OpenvinoProvider) GetDefaultClient() *client.Client {
	// default host
	host := "127.0.0.1:16666"
	if o.EngineConfig.Host != "" {
		host = o.EngineConfig.Host
	}

	// default scheme
	scheme := "http"
	if o.EngineConfig.Scheme == "https" {
		scheme = "https"
	}

	return client.NewClient(&url.URL{
		Scheme: scheme,
		Host:   host,
	}, http.DefaultClient)
}

func (o *OpenvinoProvider) StartEngine() error {
	// todo

	return nil
}

func (o *OpenvinoProvider) StopEngine() error {
	// todo

	return nil
}

func (o *OpenvinoProvider) GetConfig() *types.EngineRecommendConfig {
	// todo
	return nil
}

func (o *OpenvinoProvider) HealthCheck() error {
	// todo
	return nil
}

func (o *OpenvinoProvider) GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error) {
	return &types.EngineVersionResponse{
		Version: "1.0",
	}, nil
}

func (o *OpenvinoProvider) InstallEngine() error {
	// todo
	return nil
}

func (o *OpenvinoProvider) InitEnv() error {
	// todo  set env
	return nil
}

func (o *OpenvinoProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	return nil, nil
}

func (o *OpenvinoProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	dataCh := make(chan []byte)
	errCh := make(chan error)
	return dataCh, errCh
}

func (o *OpenvinoProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	fmt.Printf("Ollama: Deleting model %s\n", req.Model)
	return nil
}

func (o *OpenvinoProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	return nil, nil
}

func (o *OpenvinoProvider) CopyModel(ctx context.Context, req *types.CopyModelRequest) error {
	return nil
}
