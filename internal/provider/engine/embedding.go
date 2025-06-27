package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"
)

// Ollama embedding API 响应结构，兼容 data 和 embeddings 两种格式
// 以及解析逻辑兼容

type ollamaEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	// 兼容 embeddings 字段
	Embeddings [][]float32 `json:"embeddings"`
}

func (e *Engine) GenerateEmbedding(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {

	originalModel := req.Model
	modelInfo := e.GetModelById(ctx, req.Model)
	modelName := modelInfo.Name

	// 需要将embed模型加装到数据库中 service (这个的作用是记录运行中的各类模型)
	err := e.restoreCurrentModel(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to restore model: %w", err)
	}

	fmt.Printf("[Embedding] Embed 模型: %s -> %s\n", originalModel, modelName)

	// 这里重新设置model为modelName因为与ollama交互只认name
	req.Model = modelName
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	serviceReq := &types.ServiceRequest{
		Service:    "embed",
		Model:      modelName,
		FromFlavor: "ollama",
		HTTP: types.HTTPContent{
			Body: body,
		},
	}
	_, ch := schedule.GetScheduler().Enqueue(serviceReq)
	select {
	case result := <-ch:
		if result.Error != nil {
			return nil, result.Error
		}
		var apiResp ollamaEmbeddingResponse
		if err := json.Unmarshal(result.HTTP.Body, &apiResp); err != nil {
			return nil, err
		}
		resp := &types.EmbeddingResponse{
			Object: apiResp.Object,
			Model:  apiResp.Model,
			Usage: types.EmbeddingUsage{
				PromptTokens: apiResp.Usage.PromptTokens,
				TotalTokens:  apiResp.Usage.TotalTokens,
			},
		}
		// 兼容 data 字段
		for _, d := range apiResp.Data {
			resp.Embeddings = append(resp.Embeddings, d.Embedding)
		}

		if len(resp.Embeddings) == 0 && len(apiResp.Embeddings) > 0 {
			resp.Embeddings = append(resp.Embeddings, apiResp.Embeddings...)
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (e *Engine) restoreCurrentModel(ctx context.Context, modelId string) error {
	model := e.GetModelById(ctx, modelId)
	// 检查模式是否已下载安
	if model == nil {
		return fmt.Errorf("model with ID %s not found", modelId)
	}
	// 检查是否已存在于数据库中
	modelRecord := new(types.Model)
	modelRecord.ModelName = model.Name
	err := e.ds.Get(ctx, modelRecord)
	if err != nil {
		return fmt.Errorf("failed to check model existence: %w", err)
	}

	// 设置模型上下文到 service
	if modelRecord.ModelName != "" {
		service := types.Service{
			Name: model.ServiceName,
		}

		err := e.ds.Get(ctx, &service)
		if err != nil {
			return bcode.ErrServiceRecordNotFound
		}

		if model.ServiceSource == "local" {
			service.LocalProvider = modelRecord.ProviderName
		} else {
			service.RemoteProvider = model.ServiceSource
		}

		service.HybridPolicy = "default"
		err = e.ds.Put(ctx, &service)
		if err != nil {
			return bcode.ErrServiceRecordNotFound
		}

	}
	return nil

}
