package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"byze/internal/types"
)

// 直接从Go代码调用的嵌入功能。被内部组件如playground功能使用。

func (o *OllamaProvider) GenerateEmbedding(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	c := o.GetDefaultClient()
	// 如果没有输入，返回错误
	if len(req.Input) == 0 {
		return nil, fmt.Errorf("嵌入请求至少需要包含一个输入")
	}

	if len(req.Input) > 1 {
		return o.generateBatchEmbeddings(ctx, req)
	}

	ollamaRequest := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Input[0],
	}

	var response map[string]interface{}
	if err := c.Do(ctx, http.MethodPost, "/api/embeddings", ollamaRequest, &response); err != nil {
		slog.Error("调用Ollama embeddings API失败", "error", err, "model", req.Model)
		return nil, fmt.Errorf("failed to call embeddings API: %v", err)
	}

	// 检查API错误响应
	if errMsg, hasErr := response["error"].(string); hasErr && errMsg != "" {
		slog.Error("Ollama embeddings API返回错误", "error", errMsg)
		return nil, fmt.Errorf("embedding API error: %s", errMsg)
	}
	// 从响应中提取嵌入向量
	var embedding []float32

	// 尝试直接获取float32数组
	embedding, ok := response["embedding"].([]float32)
	if !ok {
		rawEmbedding, ok := response["embedding"].([]interface{})
		if !ok {
			slog.Error("Ollama返回的嵌入格式无效", "response", response)
			return nil, fmt.Errorf("invalid embedding format from model: embedding field not found or not an array")
		}

		embedding = make([]float32, len(rawEmbedding))
		for i, val := range rawEmbedding {
			switch v := val.(type) {
			case float64:
				embedding[i] = float32(v)
			case float32:
				embedding[i] = v
			case int:
				embedding[i] = float32(v)
			case int64:
				embedding[i] = float32(v)
			default:
				slog.Error("嵌入值类型无效", "value_type", fmt.Sprintf("%T", val), "index", i)
				return nil, fmt.Errorf("invalid embedding value type at index %d: %T", i, val)
			}
		}
	}

	// 验证嵌入向量的维度
	if len(embedding) == 0 {
		slog.Error("嵌入向量为空")
		return nil, fmt.Errorf("empty embedding vector returned by model")
	}

	slog.Debug("成功获取嵌入向量", "model", req.Model, "dimensions", len(embedding))
	embeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data: []types.EmbeddingData{
			{
				Object:     "embedding",
				Embedding:  embedding,
				EmbedIndex: 0,
			},
		},
		Model: req.Model,
		Usage: types.EmbeddingUsage{
			PromptTokens: len(req.Input[0]) / 4,
			TotalTokens:  len(req.Input[0]) / 4,
		},
	}
	return embeddingResponse, nil
}

func (o *OllamaProvider) generateBatchEmbeddings(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	c := o.GetDefaultClient()

	ollamaRequest := map[string]interface{}{
		"model": req.Model,
		"texts": req.Input,
	}

	var response map[string]interface{}

	if err := c.Do(ctx, http.MethodPost, "/api/embed", ollamaRequest, &response); err != nil {
		slog.Error("调用Ollama批量嵌入API失败", "error", err, "model", req.Model)

		slog.Info("回退到传统单条嵌入方法")
		return o.fallbackToSingleEmbeddings(ctx, req)
	}

	// 检查API错误响应
	if errMsg, hasErr := response["error"].(string); hasErr && errMsg != "" {

		if strings.Contains(strings.ToLower(errMsg), "not found") ||
			strings.Contains(strings.ToLower(errMsg), "method not allowed") {
			slog.Info("Ollama不支持批量嵌入API，回退到传统单条嵌入方法")
			return o.fallbackToSingleEmbeddings(ctx, req)
		}

		slog.Error("Ollama批量嵌入API返回错误", "error", errMsg)
		return nil, fmt.Errorf("batch embedding API error: %s", errMsg)
	}

	// 从响应中提取嵌入向量
	rawEmbeddings, ok := response["embeddings"].([]interface{})
	if !ok {
		slog.Error("Ollama批量嵌入返回的格式无效", "response", response)
		return nil, fmt.Errorf("invalid batch embedding format from model: embeddings field not found or not an array")
	}

	// 创建响应对象
	embeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   make([]types.EmbeddingData, 0, len(rawEmbeddings)),
		Model:  req.Model,
		Usage: types.EmbeddingUsage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}

	// 处理每个嵌入向量
	for i, rawEmbeddingObj := range rawEmbeddings {
		rawEmbedding, ok := rawEmbeddingObj.([]interface{})
		if !ok {
			slog.Error("嵌入向量格式无效", "index", i)
			continue
		}

		embedding := make([]float32, len(rawEmbedding))
		for j, val := range rawEmbedding {
			switch v := val.(type) {
			case float64:
				embedding[j] = float32(v)
			case float32:
				embedding[j] = v
			case int:
				embedding[j] = float32(v)
			case int64:
				embedding[j] = float32(v)
			default:
				slog.Error("嵌入值类型无效", "value_type", fmt.Sprintf("%T", val), "index", i)
				continue
			}
		}
		if len(embedding) == 0 {
			slog.Error("嵌入向量为空", "index", i)
			continue
		}

		embeddingResponse.Data = append(embeddingResponse.Data, types.EmbeddingData{
			Object:     "embedding",
			Embedding:  embedding,
			EmbedIndex: i,
		})

		// 估算token用量
		embeddingResponse.Usage.PromptTokens += len(req.Input[i]) / 4
		embeddingResponse.Usage.TotalTokens += len(req.Input[i]) / 4
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("no valid embeddings returned by batch API")
	}

	slog.Debug("成功获取批量嵌入向量", "model", req.Model, "count", len(embeddingResponse.Data))
	return embeddingResponse, nil
}

func (o *OllamaProvider) fallbackToSingleEmbeddings(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	c := o.GetDefaultClient()

	slog.Info("使用传统方式逐条生成嵌入", "count", len(req.Input))

	// 创建响应对象
	embeddingResponse := &types.EmbeddingResponse{
		Object: "list",
		Data:   make([]types.EmbeddingData, 0, len(req.Input)),
		Model:  req.Model,
		Usage: types.EmbeddingUsage{
			PromptTokens: 0,
			TotalTokens:  0,
		},
	}

	for i, input := range req.Input {
		ollamaRequest := map[string]interface{}{
			"model":  req.Model,
			"prompt": input,
		}

		var response map[string]interface{}
		if err := c.Do(ctx, http.MethodPost, "/api/embeddings", ollamaRequest, &response); err != nil {
			slog.Error("调用Ollama embeddings API失败", "error", err, "index", i)
			continue
		}

		if errMsg, hasErr := response["error"].(string); hasErr && errMsg != "" {
			slog.Error("Ollama embeddings API返回错误", "error", errMsg, "index", i)
			continue
		}

		var embedding []float32

		rawEmbedding, ok := response["embedding"].([]interface{})
		if !ok {
			slog.Error("Ollama返回的嵌入格式无效", "index", i)
			continue
		}

		embedding = make([]float32, len(rawEmbedding))
		for j, val := range rawEmbedding {

			switch v := val.(type) {
			case float64:
				embedding[j] = float32(v)
			case float32:
				embedding[j] = v
			case int:
				embedding[j] = float32(v)
			case int64:
				embedding[j] = float32(v)
			default:
				slog.Error("嵌入值类型无效", "value_type", fmt.Sprintf("%T", val), "index", i)
				continue
			}
		}
		if len(embedding) > 0 {
			embeddingResponse.Data = append(embeddingResponse.Data, types.EmbeddingData{
				Object:     "embedding",
				Embedding:  embedding,
				EmbedIndex: i,
			})

			// 估算token用量
			embeddingResponse.Usage.PromptTokens += len(input) / 4
			embeddingResponse.Usage.TotalTokens += len(input) / 4
		}
	}

	if len(embeddingResponse.Data) == 0 {
		return nil, fmt.Errorf("failed to generate any valid embeddings")
	}

	slog.Debug("成功逐条生成嵌入向量", "model", req.Model, "count", len(embeddingResponse.Data))
	return embeddingResponse, nil
}

func (o *OpenvinoProvider) GenerateEmbedding(ctx context.Context, req *types.EmbeddingRequest) (*types.EmbeddingResponse, error) {
	return nil, fmt.Errorf("embedding generation not implemented for OpenVINO provider")
}
