package model_engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"oadin/extension/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/logger"

	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"
)

// oadinStreamChunk 用于解析 Oadin 流式响应格式
type oadinStreamChunk struct {
	CreatedAt    string `json:"created_at"`
	FinishReason string `json:"finish_reason,omitempty"`
	Finished     bool   `json:"finished"`
	Id           string `json:"id"`
	Message      struct {
		Content   string         `json:"content"`
		Role      string         `json:"role"`
		Thinking  string         `json:"thinking"`
		ToolCalls []dto.ToolCall `json:"tool_calls,omitempty"`
	} `json:"message"`
	Model string `json:"model"`
}

type EngineService struct {
	js datastore.JsonDatastore
	ds datastore.Datastore
}

// NewEngineService 创建新的引擎服务实例
func NewEngineService() EngineServiceProvider {
	return &EngineService{
		js: datastore.GetDefaultJsonDatastore(),
		ds: datastore.GetDefaultDatastore(),
	}
}

// GetModelById 根据ID获取模型信息
func (e *EngineService) GetModelById(ctx context.Context, modelId string) *types.SupportModel {
	if modelId == "" {
		return &types.SupportModel{}
	}

	model := &types.SupportModel{Id: modelId}
	queryOpList := []datastore.FuzzyQueryOption{}
	queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
		Key:   "id",
		Query: modelId,
	})
	res, err := e.js.List(ctx, model, &datastore.ListOptions{FilterOptions: datastore.FilterOptions{Queries: queryOpList}})
	if err != nil {
		return &types.SupportModel{}
	}
	if len(res) == 0 {
		return &types.SupportModel{}
	}

	return res[0].(*types.SupportModel)
}

// RestoreCurrentModel 恢复当前模型状态
func (e *EngineService) RestoreCurrentModel(ctx context.Context, modelId string) error {
	model := e.GetModelById(ctx, modelId)
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

func (e *EngineService) ChatStream(ctx context.Context, req *dto.ChatRequest) (<-chan *dto.ChatResponse, <-chan error) {
	respChan := make(chan *dto.ChatResponse)
	errChan := make(chan error, 1)

	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		go func() {
			errChan <- err
			close(respChan)
			close(errChan)
		}()
		return respChan, errChan
	}

	modelName := req.Model

	hybridPolicy := "default"
	ds := datastore.GetDefaultDatastore()
	sp := &types.Service{
		Name:   "chat",
		Status: 1,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		logger.EngineLogger.Error("[Schedule] Failed to get service", "error", err, "service", "embed")
	} else {
		hybridPolicy = sp.HybridPolicy
	}
	hybridPolicy = sp.HybridPolicy

	serviceReq := &types.ServiceRequest{
		Service:       "chat",
		Model:         modelName,
		FromFlavor:    "oadin",
		HybridPolicy:  hybridPolicy,
		AskStreamMode: true,
		Think:         req.Think,
		HTTP: types.HTTPContent{
			Header: http.Header{},
			Body:   body,
		},
	}

	// 请求调度器执行任务
	_, ch := schedule.GetScheduler().Enqueue(serviceReq)

	go func() {
		defer close(respChan)
		defer close(errChan)

		accumulatedContent := ""
		var toolCalls []dto.ToolCall
		var totalDuration int64 // 跟踪总处理时间

		// 处理流式响应
		for result := range ch {
			if result.Error != nil {
				errChan <- result.Error
				return
			}

			// 如果chunk为空，则跳过
			if len(result.HTTP.Body) == 0 {
				continue
			}

			// 处理前缀
			bodyStr := string(result.HTTP.Body)
			if strings.HasPrefix(bodyStr, "data: ") {
				bodyStr = strings.TrimPrefix(bodyStr, "data: ")
				bodyStr = strings.TrimSpace(bodyStr)
			}
			// 转回[]byte
			cleanBody := []byte(bodyStr)
			fmt.Println("[ChatStream] 收到块内容", bodyStr)

			// 每个块都是一个完整的JSON对象
			var content string
			isComplete := false
			var thoughts string
			model := ""
			parseSucceeded := false

			var streamChunk oadinStreamChunk
			if err := json.Unmarshal(cleanBody, &streamChunk); err == nil {
				// 成功解析为直接流式格式
				parseSucceeded = true
				// fmt.Printf("[ChatStream] 解析为直接流式格式成功\n")

				// 提取模型名称
				model = streamChunk.Model

				// 提取内容
				if streamChunk.Message.Content != "" {
					content = streamChunk.Message.Content
				}

				if streamChunk.Message.Thinking != "" {
					thoughts = streamChunk.Message.Thinking
				}

				if len(streamChunk.Message.ToolCalls) > 0 {
					toolCalls = append(toolCalls, streamChunk.Message.ToolCalls...)
				}

				isComplete = streamChunk.Finished || streamChunk.FinishReason != ""
				if isComplete {
					fmt.Printf("[ChatStream] 检测到流式输出完成标记 finished=%v, finish_reason=%s\n",
						streamChunk.Finished, streamChunk.FinishReason)
				}
			}

			if !parseSucceeded {
				var oadinResp dto.OadinAPIResponse
				if err := json.Unmarshal(cleanBody, &oadinResp); err == nil && oadinResp.BusinessCode == 10000 {
					dataBytes, err := json.Marshal(oadinResp.Data)
					if err == nil {
						var streamResp dto.OadinChatStreamResponse
						if err := json.Unmarshal(dataBytes, &streamResp); err == nil {
							parseSucceeded = true
							fmt.Printf("[ChatStream] 解析为标准Oadin响应格式成功\n")

							model = streamResp.Model

							if len(streamResp.Choices) > 0 {
								if streamResp.Choices[0].Delta.Content != "" {
									content = streamResp.Choices[0].Delta.Content
									fmt.Printf("[ChatStream] 从Oadin delta提取内容，长度: %d\n", len(content))
								}

								if streamResp.Choices[0].Delta.Thinking != "" {
									thoughts = streamResp.Choices[0].Delta.Thinking
									fmt.Printf("[ChatStream] 从Oadin delta提取思考内容，长度: %d\n", len(thoughts))
								}

								isComplete = streamResp.Choices[0].FinishReason != ""

								if len(streamResp.Choices[0].Delta.ToolCalls) > 0 {
									toolCalls = streamResp.Choices[0].Delta.ToolCalls
									fmt.Printf("[ChatStream] 从Oadin delta提取到工具调用，数量: %d\n", len(toolCalls))
									logger.EngineLogger.Info("[ChatStream] 从Oadin delta提取到工具调用", "数量", len(toolCalls))
								}
							}
						}
					}
				}
			}

			if !parseSucceeded {
				fmt.Printf("[ChatStream] Oadin API响应解析失败，尝试通用格式解析\n")

				var data map[string]interface{}
				if err := json.Unmarshal(cleanBody, &data); err == nil {
					extractContentLocal := func(data map[string]interface{}) (string, bool) {
						if msg, ok := data["message"].(map[string]interface{}); ok {
							if content, ok := msg["content"].(string); ok && content != "" {
								return content, true
							}
						}

						if response, ok := data["response"].(string); ok && response != "" {
							return response, true
						}

						if content, ok := data["content"].(string); ok && content != "" {
							return content, true
						}

						return "", false
					}

					// 提取消息内容
					extractedContent, found := extractContentLocal(data)
					if found {
						content = extractedContent
						fmt.Printf("[ChatStream] 提取到内容，长度: %d\n", len(content))
					}

					// 检查是否完�?
					if done, ok := data["done"].(bool); ok {
						isComplete = done
					}

					// 提取模型名称
					if m, ok := data["model"].(string); ok {
						model = m
					}

					// 提取思考内�?
					if msg, ok := data["message"].(map[string]interface{}); ok {
						if th, ok := msg["thinking"].(string); ok && th != "" {
							thoughts = th
							fmt.Printf("[ChatStream] 从通用格式message.thinking中提取到思考内容，长度: %d\n", len(thoughts))
						}
					}
					// 如果没有在message中找到thinking，尝试从顶层查找
					if thoughts == "" {
						if th, ok := data["thinking"].(string); ok && th != "" {
							thoughts = th
							fmt.Printf("[ChatStream] 从顶层thinking中提取到思考内容，长度: %d\n", len(thoughts))
						}
					}

					// 提取工具调用
					if msg, ok := data["message"].(map[string]interface{}); ok {
						if tc, ok := msg["tool_calls"].([]dto.ToolCall); ok && len(tc) > 0 {
							toolCalls = tc
							fmt.Printf("[ChatStream] 提取到工具调用，数量: %d\n", len(toolCalls))
							logger.EngineLogger.Info("[ChatStream] 从message中提取到工具调用", "数量", len(toolCalls))
						}
					}

					// 如果没有在message中找到，尝试从顶层查�?
					if len(toolCalls) == 0 {
						if tc, ok := data["tool_calls"].([]dto.ToolCall); ok && len(tc) > 0 {
							toolCalls = tc
							fmt.Printf("[ChatStream] 从顶层提取到工具调用，数�? %d\n", len(toolCalls))
						}
					}
				} else {
					fmt.Printf("[ChatStream] JSON解析完全失败: %v，跳过此块\n", err)
					continue
				}
			}

			// 处理提取到的内容
			if content != "" {
				accumulatedContent += content
			}

			// 创建响应对象
			resp := &dto.ChatResponse{
				Content:       content,
				Model:         model,
				IsComplete:    isComplete,
				ToolCalls:     toolCalls,
				Object:        "chat.completion.chunk",
				TotalDuration: totalDuration,
				Thoughts:      thoughts,
			}

			if isComplete {
				fmt.Printf("[ChatStream] 收到完成标记，当前块内容长度: %d，累积内容长�? %d\n",
					len(content), len(accumulatedContent))
			}

			// 发送响�?
			// 只发送有内容或是最后一个块的响�?
			if (resp.Content != "" || resp.Thoughts != "") || resp.IsComplete {
				// 如果是最后一个块，发送完整累积的内容
				if resp.IsComplete {
					var tempData map[string]interface{}
					if err := json.Unmarshal(cleanBody, &tempData); err == nil {
						if duration, ok := tempData["total_duration"].(float64); ok {
							totalDuration = int64(duration)
							resp.TotalDuration = totalDuration
							fmt.Printf("[ChatStream] 提取到总时�? %dms\n", totalDuration)
						}
					}

					// 确保最后一块还包含之前累积的内�?
					resp.Content = accumulatedContent
					resp.Object = "chat.completion"
					fmt.Printf("[ChatStream] 发送最终完整响应，内容长度: %d\n", len(accumulatedContent))

					if len(accumulatedContent) == 0 {
						fmt.Printf("[ChatStream] 警告：累积内容为空，尝试使用最后接收的非空内容\n")
					}
				}
				respChan <- resp
			}
		}
	}()
	return respChan, errChan
}

// GenerateEmbedding 实现向量嵌入生成功能
func (e *EngineService) GenerateEmbedding(ctx context.Context, req *dto.EmbeddingRequest) (*dto.EmbeddingResponse, error) {
	originalModel := req.Model
	modelInfo := e.GetModelById(ctx, req.Model)
	modelName := modelInfo.Name

	err := e.RestoreCurrentModel(ctx, req.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to restore model: %w", err)
	}

	fmt.Printf("[Embedding] Embed 模型: %s -> %s\n", originalModel, modelName)

	req.Model = modelName

	if len(req.Input) == 0 {
		return nil, fmt.Errorf("embedding input cannot be empty")
	}

	logger.EngineLogger.Info("[Embedding] 正在生成向量",
		"model", req.Model,
		"input_count", len(req.Input),
		"first_input_length", len(req.Input[0]))

	body, err := json.Marshal(req)
	if err != nil {
		logger.EngineLogger.Error("[Embedding] 请求体序列化失败", "error", err)
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

	logger.EngineLogger.Info("[Embedding] Request body for embedding",
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
		logger.EngineLogger.Error("[Schedule] Failed to get service", "error", err, "service", "embed")
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

	logger.EngineLogger.Info("[Embedding] 发送向量生成请求到调度器",
		"model", modelName,
		"body_length", len(body))

	_, ch := schedule.GetScheduler().Enqueue(serviceReq)
	select {
	case result := <-ch:
		if result.Error != nil {
			logger.EngineLogger.Error("[Embedding] Error from service provider", "error", result.Error)
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

		logger.EngineLogger.Info("[Embedding] Received response from service provider",
			"response_size", len(result.HTTP.Body),
			"response_preview", responsePreview)

		var directData []map[string]interface{}
		if err := json.Unmarshal(result.HTTP.Body, &directData); err == nil && len(directData) > 0 {
			logger.EngineLogger.Info("[Embedding] 成功解析为直接数据数组",
				"array_length", len(directData))

			resp := &dto.EmbeddingResponse{
				Object: "list", // 默认对象类型
				Model:  req.Model,
				Usage: dto.EmbeddingUsage{
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
				logger.EngineLogger.Info("[Embedding] 成功从直接数据数组提取向量",
					"embedding_count", len(resp.Embeddings),
					"vector_dim", len(resp.Embeddings[0]))
				return resp, nil
			} else {
				logger.EngineLogger.Warn("[Embedding] 数据数组中未找到有效的向量数据")
			}
		}

		var oadinResp dto.OadinAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &oadinResp); err != nil {

			logger.EngineLogger.Warn("[Embedding] 无法解析�?Oadin 响应格式",
				"error", err.Error(),
				"response_preview", responsePreview)
		}

		if oadinResp.BusinessCode > 0 && oadinResp.BusinessCode != 10000 {
			logger.EngineLogger.Error("[Embedding] Oadin API 返回错误",
				"business_code", oadinResp.BusinessCode,
				"message", oadinResp.Message)
			return nil, fmt.Errorf("Oadin Embedding API 错误: %s (代码: %d)",
				oadinResp.Message, oadinResp.BusinessCode)
		}

		if oadinResp.Data == nil {
			logger.EngineLogger.Error("[Embedding] Oadin 响应�?data 字段为空")
			return nil, fmt.Errorf("Oadin embedding 响应�?data 字段为空")
		}

		dataBytes, err := json.Marshal(oadinResp.Data)
		if err != nil {
			logger.EngineLogger.Error("[Embedding] Oadin data 字段序列化失败", "error", err)
			return nil, fmt.Errorf("Oadin data 字段序列化失败: %w", err)
		}

		logger.EngineLogger.Debug("[Embedding] Oadin 响应 data 字段内容",
			"data_json", string(dataBytes))

		var dataArray []map[string]interface{}
		parseArrayOk := false
		resp := &dto.EmbeddingResponse{
			Object: "list",
			Model:  req.Model,
			Usage: dto.EmbeddingUsage{
				PromptTokens: 0,
				TotalTokens:  0,
			},
		}
		if err := json.Unmarshal(dataBytes, &dataArray); err == nil && len(dataArray) > 0 {
			logger.EngineLogger.Info("[Embedding] data 字段是数组格式", "length", len(dataArray))
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
				logger.EngineLogger.Info("[Embedding] 成功从data 数组中提取向量",
					"embedding_count", len(resp.Embeddings),
					"vector_dim", len(resp.Embeddings[0]))
				parseArrayOk = true
				return resp, nil
			}
		}

		if !parseArrayOk {
			var embedResp dto.OadinEmbeddingResponse
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
					logger.EngineLogger.Info("[Embedding] 成功解析 Oadin embedding 响应",
						"embedding_count", len(resp.Embeddings),
						"vector_dim", len(resp.Embeddings[0]),
						"model", resp.Model)
					return resp, nil
				}
			} else {
				logger.EngineLogger.Warn("[Embedding] 无法解析为标�?OadinEmbeddingResponse",
					"error", err.Error())

				var directEmbeddings [][]float32
				if err := json.Unmarshal(dataBytes, &directEmbeddings); err == nil && len(directEmbeddings) > 0 {
					logger.EngineLogger.Info("[Embedding] �?Oadin data 字段直接解析�?embeddings 数组",
						"embedding_count", len(directEmbeddings))
					resp.Embeddings = directEmbeddings
					return resp, nil
				}
				logger.EngineLogger.Error("[Embedding] 所有解析尝试均失败", "data_json", string(dataBytes))
				return nil, fmt.Errorf("Oadin embedding 响应解析失败: 无法从数据中提取向量")
			}
		}
	case <-ctx.Done():
		logger.EngineLogger.Error("[Embedding] 上下文已取消", "error", ctx.Err())
		return nil, fmt.Errorf("向量生成请求被取�? %w", ctx.Err())
	}

	return nil, fmt.Errorf("embedding 处理过程中未能生成有效响应")
}

func cleanModelId(modelId string) string {
	for _, sep := range []string{",", " "} {
		if idx := strings.Index(modelId, sep); idx > 0 {
			modelId = modelId[:idx]
			break
		}
	}
	result := strings.ReplaceAll(modelId, "-", "")
	result = strings.ReplaceAll(result, "/", "_")
	return result
}

// Chat 实现非流式聊天功�?
func (e *EngineService) Chat(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	originalModel := req.Model
	modelInfo := e.GetModelById(ctx, req.Model)
	modelName := modelInfo.Name

	if modelName == "" {
		modelName = originalModel
	}

	hybridPolicy := "default"
	ds := datastore.GetDefaultDatastore()
	sp := &types.Service{
		Name:   "chat",
		Status: 1,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		logger.EngineLogger.Error("[Schedule] Failed to get service", "error", err, "service", "embed")
	} else {
		hybridPolicy = sp.HybridPolicy
	}
	hybridPolicy = sp.HybridPolicy

	serviceReq := &types.ServiceRequest{
		Service:      "chat",
		Model:        modelName,
		FromFlavor:   "oadin",
		HybridPolicy: hybridPolicy,
		Think:        req.Think,
		HTTP: types.HTTPContent{
			Header: http.Header{},
			Body:   body,
		},
	}
	_, ch := schedule.GetScheduler().Enqueue(serviceReq)
	select {
	case result := <-ch:
		if result.Error != nil {
			return nil, result.Error
		}

		// Log the raw response for debugging
		fmt.Printf("[Chat] Raw response received, length: %d\n", len(result.HTTP.Body))
		fmt.Printf("[Chat] Raw response content: %s\n", string(result.HTTP.Body))

		var oadinResp dto.OadinAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &oadinResp); err == nil && oadinResp.BusinessCode == 10000 {

			dataBytes, err := json.Marshal(oadinResp.Data)
			if err != nil {
				return nil, fmt.Errorf("解析Oadin响应data字段失败: %v", err)
			}

			var chatResp dto.OadinChatResponse
			if err := json.Unmarshal(dataBytes, &chatResp); err != nil {
				return nil, fmt.Errorf("解析Oadin聊天响应结构失败: %v", err)
			}

			if len(chatResp.Choices) == 0 {
				return nil, fmt.Errorf("Oadin响应中没有choices选项")
			}

			response := &dto.ChatResponse{
				ID:         chatResp.ID,
				Object:     chatResp.Object,
				Model:      chatResp.Model,
				Content:    chatResp.Choices[0].Message.Content,
				Thoughts:   chatResp.Choices[0].Message.Thinking,
				ToolCalls:  chatResp.Choices[0].Message.ToolCalls,
				IsComplete: chatResp.Choices[0].FinishReason != "",
			}

			fmt.Printf("[Chat] Oadin API解析成功，内容长度：%d\n", len(response.Content))
			if len(response.ToolCalls) > 0 {
				fmt.Printf("[Chat] Oadin API检测到工具调用，数量：%d\n", len(response.ToolCalls))
			}
			if response.Thoughts != "" {
				fmt.Printf("[Chat] Oadin API检测到思考内容，长度�?d\n", len(response.Thoughts))
			}
			return response, nil
		}

		// 尝试直接解析成完整的ChatResponse
		var response dto.ChatResponse
		if err := json.Unmarshal(result.HTTP.Body, &response); err == nil && response.Content != "" {
			fmt.Printf("[Chat] 直接解析成功，内容长度：%d\n", len(response.Content))
			return &response, nil
		}

		// 如果上述方法都失败，回退到使用通用map解析
		fmt.Printf("[Chat] 标准解析方式失败，尝试通用map解析\n")
		var data map[string]interface{}
		if err := json.Unmarshal(result.HTTP.Body, &data); err != nil {
			return nil, fmt.Errorf("无法解析API响应: %v", err)
		}

		content := ""
		isComplete := false
		model := ""
		var toolCalls []dto.ToolCall
		var thoughts string
		var totalDuration int64

		// 尝试提取message.content
		if msg, ok := data["message"].(map[string]interface{}); ok {
			// 提取content
			if c, ok := msg["content"].(string); ok {
				content = c
				fmt.Printf("[Chat] 从message.content提取内容，长�? %d\n", len(content))
			}

			// 提取thinking
			if th, ok := msg["thinking"].(string); ok {
				thoughts = th
				fmt.Printf("[Chat] 从message.thinking提取思考内容，长度: %d\n", len(thoughts))
			}

			// 提取tool_calls
			if tc, ok := msg["tool_calls"].([]dto.ToolCall); ok && len(tc) > 0 {
				toolCalls = tc
				fmt.Printf("[Chat] 提取到tool_calls，数�? %d\n", len(toolCalls))
			}
		}

		// 如果message.content为空，尝试提取response
		if content == "" {
			if resp, ok := data["response"].(string); ok {
				content = resp
				fmt.Printf("[Chat] 从response提取内容，长�? %d\n", len(content))
			}
		}

		// 如果还是为空，尝试直接提取content字段
		if content == "" {
			if c, ok := data["content"].(string); ok {
				content = c
				fmt.Printf("[Chat] 从content字段提取内容，长�? %d\n", len(content))
			}
		}

		// 提取完成标志
		if done, ok := data["done"].(bool); ok {
			isComplete = done
		}

		// 提取模型名称
		if m, ok := data["model"].(string); ok {
			model = m
		}

		// 提取处理时间
		if td, ok := data["total_duration"].(float64); ok {
			totalDuration = int64(td)
		}

		// 若还没有提取到工具调用，尝试从顶层提�?
		if len(toolCalls) == 0 {
			if tc, ok := data["tool_calls"].([]dto.ToolCall); ok && len(tc) > 0 {
				toolCalls = tc
				fmt.Printf("[Chat] 从顶层提取到tool_calls，数�? %d\n", len(toolCalls))
			}
		}
		// 尝试从顶级字段提取thinking
		if thoughts == "" {
			if th, ok := data["thinking"].(string); ok {
				thoughts = th
				fmt.Printf("[Chat] 从顶级thinking字段提取思考内容，长度: %d\n", len(thoughts))
			}
		}

		// 创建响应对象
		resp := &dto.ChatResponse{
			Content:       content,
			Model:         model,
			IsComplete:    isComplete,
			Object:        "chat.completion",
			ToolCalls:     toolCalls,
			Thoughts:      thoughts,
			TotalDuration: totalDuration,
		}
		fmt.Printf("[Chat] 通用解析成功，内容长�? %d\n", len(content))
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
