package engine

import (
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
	"fmt"
)

// Ollama embedding API 响应结构
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
}

func (e *Engine) GenerateEmbedding(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	originalModel := req.Model
	modelName := getModelNameById(req.Model)

	fmt.Printf("[Embedding] Embed 模型: %s -> %s\n", originalModel, modelName)

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
		// 转换 Data
		for _, d := range apiResp.Data {
			resp.Data = append(resp.Data, types.EmbeddingData{
				Object:     d.Object,
				Embedding:  d.Embedding,
				EmbedIndex: d.Index,
			})
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
