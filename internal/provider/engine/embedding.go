package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"oadin/internal/datastore"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"
)

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

	err := e.restoreCurrentModel(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to restore model: %w", err)
	}

	fmt.Printf("[Embedding] Embed 模型: %s -> %s\n", originalModel, modelName)

	req.Model = modelName

	if len(req.Input) == 0 {
		return nil, fmt.Errorf("embedding input cannot be empty")
	}

	slog.Info("[Embedding] 正在生成向量",
		"model", req.Model,
		"input_count", len(req.Input),
		"first_input_length", len(req.Input[0]))

	body, err := json.Marshal(req)
	if err != nil {
		slog.Error("[Embedding] 请求体序列化失败", "error", err)
		return nil, err
	}

	inputSample := ""
	if len(req.Input) > 0 {
		sampleLength := 30
		if len(req.Input[0]) < sampleLength {
			sampleLength = len(req.Input[0])
		}
		inputSample = req.Input[0][:sampleLength] + "..."
	}

	slog.Info("[Embedding] Request body for embedding",
		"model", req.Model,
		"input_count", len(req.Input),
		"input_sample", inputSample,
		"body_size", len(body))

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	hybridPolicy := "default"
	ds := datastore.GetDefaultDatastore()
	sp := &types.Service{
		Name:   "embed",
		Status: 1,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		slog.Error("[Schedule] Failed to get service", "error", err, "service", "embed")
	} else {
		hybridPolicy = sp.HybridPolicy
	}
	hybridPolicy = sp.HybridPolicy

	serviceReq := &types.ServiceRequest{
		Service:      "embed",
		Model:        modelName,
		FromFlavor:   "oadin",
		HybridPolicy: hybridPolicy,
		HTTP: types.HTTPContent{
			Header: headers,
			Body:   body,
		},
	}

	slog.Info("[Embedding] 发送向量生成请求到调度器",
		"model", modelName,
		"body_length", len(body))

	_, ch := schedule.GetScheduler().Enqueue(serviceReq)
	select {
	case result := <-ch:
		if result.Error != nil {
			slog.Error("[Embedding] Error from service provider", "error", result.Error)
			return nil, result.Error
		}

		responsePreview := "empty"
		if len(result.HTTP.Body) > 0 {
			previewLength := 100
			if len(result.HTTP.Body) < previewLength {
				previewLength = len(result.HTTP.Body)
			}
			responsePreview = string(result.HTTP.Body[:previewLength]) + "..."
		}

		slog.Info("[Embedding] Received response from service provider",
			"response_size", len(result.HTTP.Body),
			"response_preview", responsePreview)

		var directData []map[string]interface{}
		if err := json.Unmarshal(result.HTTP.Body, &directData); err == nil && len(directData) > 0 {
			slog.Info("[Embedding] 成功解析为直接数据数组",
				"array_length", len(directData))

			resp := &types.EmbeddingResponse{
				Object: "list", // 默认对象类型
				Model:  req.Model,
				Usage: types.EmbeddingUsage{
					PromptTokens: 0,
					TotalTokens:  0,
				},
			}

			for _, item := range directData {
				if embVector, ok := item["embedding"].([]interface{}); ok {
					embedding := make([]float32, len(embVector))
					for i, val := range embVector {
						if floatVal, ok := val.(float64); ok {
							embedding[i] = float32(floatVal)
						}
					}
					resp.Embeddings = append(resp.Embeddings, embedding)
				}
			}

			if len(resp.Embeddings) > 0 {
				slog.Info("[Embedding] 成功从直接数据数组提取向量",
					"embedding_count", len(resp.Embeddings),
					"vector_dim", len(resp.Embeddings[0]))
				return resp, nil
			} else {
				slog.Warn("[Embedding] 数据数组中未找到有效的向量数据")
			}
		}

		var oadinResp types.OadinAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &oadinResp); err != nil {

			slog.Warn("[Embedding] 无法解析为 Oadin 响应格式，尝试兼容模式",
				"error", err.Error(),
				"response_preview", responsePreview)

			var ollamaResp ollamaEmbeddingResponse
			if err := json.Unmarshal(result.HTTP.Body, &ollamaResp); err != nil {
				slog.Error("[Embedding] 响应解析失败，既不符合 Oadin 格式也不符合兼容格式",
					"error", err.Error(),
					"response", string(result.HTTP.Body))
				return nil, fmt.Errorf("embedding 响应解析失败: %w", err)
			}

			resp := &types.EmbeddingResponse{
				Object: ollamaResp.Object,
				Model:  ollamaResp.Model,
				Usage: types.EmbeddingUsage{
					PromptTokens: ollamaResp.Usage.PromptTokens,
					TotalTokens:  ollamaResp.Usage.TotalTokens,
				},
			}

			for _, d := range ollamaResp.Data {
				resp.Embeddings = append(resp.Embeddings, d.Embedding)
			}

			if len(resp.Embeddings) == 0 && len(ollamaResp.Embeddings) > 0 {
				resp.Embeddings = append(resp.Embeddings, ollamaResp.Embeddings...)
			}

			slog.Info("[Embedding] 成功从兼容格式解析响应",
				"embedding_count", len(resp.Embeddings),
				"model", resp.Model)
			return resp, nil
		}

		if oadinResp.BusinessCode > 0 && oadinResp.BusinessCode != 10000 {
			slog.Error("[Embedding] Oadin API 返回错误",
				"business_code", oadinResp.BusinessCode,
				"message", oadinResp.Message)
			return nil, fmt.Errorf("Oadin Embedding API 错误: %s (代码: %d)",
				oadinResp.Message, oadinResp.BusinessCode)
		}

		if oadinResp.Data == nil {
			slog.Error("[Embedding] Oadin 响应中 data 字段为空")
			return nil, fmt.Errorf("Oadin embedding 响应中 data 字段为空")
		}

		dataBytes, err := json.Marshal(oadinResp.Data)
		if err != nil {
			slog.Error("[Embedding] Oadin data 字段序列化失败", "error", err)
			return nil, fmt.Errorf("Oadin data 字段序列化失败: %w", err)
		}

		slog.Debug("[Embedding] Oadin 响应 data 字段内容",
			"data_json", string(dataBytes))

		var dataArray []map[string]interface{}
		parseArrayOk := false
		resp := &types.EmbeddingResponse{
			Object: "list",
			Model:  req.Model,
			Usage: types.EmbeddingUsage{
				PromptTokens: 0,
				TotalTokens:  0,
			},
		}
		if err := json.Unmarshal(dataBytes, &dataArray); err == nil && len(dataArray) > 0 {
			slog.Info("[Embedding] data 字段是数组格式", "length", len(dataArray))
			for _, item := range dataArray {
				if embVector, ok := item["embedding"].([]interface{}); ok {
					embedding := make([]float32, len(embVector))
					for i, val := range embVector {
						if floatVal, ok := val.(float64); ok {
							embedding[i] = float32(floatVal)
						}
					}
					resp.Embeddings = append(resp.Embeddings, embedding)
				}
			}
			if len(resp.Embeddings) > 0 {
				slog.Info("[Embedding] 成功从 data 数组中提取向量",
					"embedding_count", len(resp.Embeddings),
					"vector_dim", len(resp.Embeddings[0]))
				parseArrayOk = true
				return resp, nil
			}
		}

		if !parseArrayOk {
			var embedResp types.OadinEmbeddingResponse
			if err := json.Unmarshal(dataBytes, &embedResp); err == nil {
				resp.Object = embedResp.Object
				resp.Model = embedResp.Model
				if embedResp.Usage != nil {
					resp.Usage.PromptTokens = embedResp.Usage["prompt_tokens"]
					resp.Usage.TotalTokens = embedResp.Usage["total_tokens"]
				}
				for _, d := range embedResp.Data {
					resp.Embeddings = append(resp.Embeddings, d.Embedding)
				}
				if len(resp.Embeddings) > 0 {
					slog.Info("[Embedding] 成功解析 Oadin embedding 响应",
						"embedding_count", len(resp.Embeddings),
						"vector_dim", len(resp.Embeddings[0]),
						"model", resp.Model)
					return resp, nil
				}
			} else {
				slog.Warn("[Embedding] 无法解析为标准 OadinEmbeddingResponse",
					"error", err.Error())

				var directEmbeddings [][]float32
				if err := json.Unmarshal(dataBytes, &directEmbeddings); err == nil && len(directEmbeddings) > 0 {
					slog.Info("[Embedding] 从 Oadin data 字段直接解析出 embeddings 数组",
						"embedding_count", len(directEmbeddings))
					resp.Embeddings = directEmbeddings
					return resp, nil
				}
				slog.Error("[Embedding] 所有解析尝试均失败", "data_json", string(dataBytes))
				return nil, fmt.Errorf("Oadin embedding 响应解析失败: 无法从数据中提取向量")
			}
		}
	case <-ctx.Done():
		slog.Error("[Embedding] 上下文已取消", "error", ctx.Err())
		return nil, fmt.Errorf("向量生成请求被取消: %w", ctx.Err())
	}

	return nil, fmt.Errorf("embedding 处理过程中未能生成有效响应")
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
