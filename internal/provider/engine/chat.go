package engine

import (
	"byze/internal/datastore"
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Engine struct {
	js datastore.JsonDatastore
	ds datastore.Datastore
}

type ollamaAPIResponse struct {
	Model              string           `json:"model"`
	CreatedAt          string           `json:"created_at"`
	Response           string           `json:"response"`
	Content            string           `json:"content,omitempty"`
	Done               bool             `json:"done"`
	DoneReason         string           `json:"done_reason,omitempty"`
	TotalDuration      int64            `json:"total_duration,omitempty"`
	LoadDuration       int64            `json:"load_duration,omitempty"`
	PromptEvalCount    int              `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64            `json:"prompt_eval_duration,omitempty"`
	EvalCount          int              `json:"eval_count,omitempty"`
	EvalDuration       int64            `json:"eval_duration,omitempty"`
	Message            *ollamaMessage   `json:"message,omitempty"` // Used by /api/chat
	ToolCalls          []map[string]any `json:"tool_calls,omitempty"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []map[string]any `json:"tool_calls,omitempty"`
}

func NewEngine() *Engine {
	return &Engine{
		js: datastore.GetDefaultJsonDatastore(),
		ds: datastore.GetDefaultDatastore(),
	}
}

// GetModelById converts a model ID to its corresponding name
func (e *Engine) GetModelById(ctx context.Context, modelId string) *types.SupportModel {
	// If the modelId is empty, return empty string
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

func (e *Engine) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Convert model ID to model name for ServiceRequest
	originalModel := req.Model
	modelName := e.GetModelById(ctx, req.Model).Name

	// Debug log to trace model conversion
	fmt.Printf("[Chat] Model conversion: %s -> %s\n", originalModel, modelName)

	serviceReq := &types.ServiceRequest{
		Service:    "chat",
		Model:      modelName, // Use the model name instead of ID
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

		// Log the raw response for debugging
		fmt.Printf("[Chat] Raw response received, length: %d\n", len(result.HTTP.Body))

		// 尝试直接解析成完整的ChatResponse
		var response types.ChatResponse
		if err := json.Unmarshal(result.HTTP.Body, &response); err == nil && response.Content != "" {
			fmt.Printf("[Chat] 直接解析成功，内容长度：%d\n", len(response.Content))
			return &response, nil
		}

		// 尝试解析为Ollama API标准响应格式
		var ollamaResp ollamaAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &ollamaResp); err == nil {
			// 构建ChatResponse
			content := ""

			// 按优先级提取content
			// 1. 如果存在message字段且有内容，使用message.content (优先 /api/chat 格式)
			if ollamaResp.Message != nil && ollamaResp.Message.Content != "" {
				content = ollamaResp.Message.Content
				fmt.Printf("[Chat] 从message.content提取内容，长度: %d\n", len(content))
			} else if ollamaResp.Response != "" {
				// 2. 如果存在response字段且有内容，使用response (/api/generate 格式)
				content = ollamaResp.Response
				fmt.Printf("[Chat] 从response提取内容，长度: %d\n", len(content))
			} else if ollamaResp.Content != "" {
				// 3. 如果存在content字段且有内容，使用content (备用)
				content = ollamaResp.Content
				fmt.Printf("[Chat] 从content提取内容，长度: %d\n", len(content))
			}

			// 提取完成标志
			isComplete := ollamaResp.Done

			// 提取模型名称
			model := ollamaResp.Model

			// 提取工具调用(如果有)
			var toolCalls []any
			if ollamaResp.Message != nil && ollamaResp.Message.ToolCalls != nil && len(ollamaResp.Message.ToolCalls) > 0 {
				toolCalls = make([]any, len(ollamaResp.Message.ToolCalls))
				for i, tc := range ollamaResp.Message.ToolCalls {
					toolCalls[i] = tc
				}
				fmt.Printf("[Chat] 提取到工具调用，数量: %d\n", len(toolCalls))
			} else if ollamaResp.ToolCalls != nil && len(ollamaResp.ToolCalls) > 0 {
				toolCalls = make([]any, len(ollamaResp.ToolCalls))
				for i, tc := range ollamaResp.ToolCalls {
					toolCalls[i] = tc
				}
				fmt.Printf("[Chat] 提取到工具调用，数量: %d\n", len(toolCalls))
			}

			// 创建响应对象
			resp := &types.ChatResponse{
				Content:    content,
				Model:      model,
				IsComplete: isComplete,
				Object:     "chat.completion",
				ToolCalls:  toolCalls,
			}
			return resp, nil
		}

		// 如果上述方法都失败，回退到使用通用map解析
		var data map[string]interface{}
		if err := json.Unmarshal(result.HTTP.Body, &data); err != nil {
			return nil, fmt.Errorf("无法解析API响应: %v", err)
		}

		content := ""
		isComplete := false
		model := ""
		var toolCalls []any

		// 尝试提取message.content (Ollama /api/chat)
		if msg, ok := data["message"].(map[string]interface{}); ok {
			// 提取content
			if c, ok := msg["content"].(string); ok {
				content = c
				fmt.Printf("[Chat] 从message.content提取内容，长度: %d\n", len(content))
			}

			// 提取tool_calls
			if tc, ok := msg["tool_calls"].([]interface{}); ok && len(tc) > 0 {
				toolCalls = tc
				fmt.Printf("[Chat] 提取到tool_calls，数量: %d\n", len(toolCalls))
			}
		}

		// 如果message.content为空，尝试提取response (Ollama /api/generate)
		if content == "" {
			if resp, ok := data["response"].(string); ok {
				content = resp
				fmt.Printf("[Chat] 从response提取内容，长度: %d\n", len(content))
			}
		}

		// 如果还是为空，尝试直接提取content字段
		if content == "" {
			if c, ok := data["content"].(string); ok {
				content = c
				fmt.Printf("[Chat] 从content字段提取内容，长度: %d\n", len(content))
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

		// 若还没有提取到工具调用，尝试从顶层提取
		if len(toolCalls) == 0 {
			if tc, ok := data["tool_calls"].([]interface{}); ok && len(tc) > 0 {
				toolCalls = tc
				fmt.Printf("[Chat] 从顶层提取到tool_calls，数量: %d\n", len(toolCalls))
			}
		}

		// 创建响应对象
		resp := &types.ChatResponse{
			Content:    content,
			Model:      model,
			IsComplete: isComplete,
			Object:     "chat.completion",
			ToolCalls:  toolCalls,
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
