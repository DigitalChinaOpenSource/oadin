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

package provider

import (
	"context"

	"oadin/internal/provider/engine"
	"oadin/internal/types"
)

// ModelServiceProvider model service provider interface
type ModelServiceProvider interface {
	// engine lifecycle management
	InstallEngine() error
	StartEngine(mode string) error
	StopEngine(ctx context.Context) error
	HealthCheck() error
	InitEnv() error

	// model management
	PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error)
	PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error)
	DeleteModel(ctx context.Context, req *types.DeleteRequest) error
	ListModels(ctx context.Context) (*types.ListResponse, error)

	// config and version
	GetConfig() *types.EngineRecommendConfig
	GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error)

	CopyModel(ctx context.Context, req *types.CopyModelRequest) error
	LoadModel(ctx context.Context, req *types.LoadRequest) error
	UnloadModel(ctx context.Context, req *types.UnloadModelRequest) error
	GetRunningModels(ctx context.Context) (*types.ListResponse, error)
	GetOperateStatus() int
	SetOperateStatus(status int)
}

// GetModelEngine get model service provider by engine name
func GetModelEngine(engineName string) ModelServiceProvider {
	switch engineName {
	case "ollama":
		return engine.NewOllamaProvider(nil)
	case "openvino":
		return engine.NewOpenvinoProvider(nil)
	default:
		return engine.NewOllamaProvider(nil)
	}
}
