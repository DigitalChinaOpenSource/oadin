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
	ToolCalls          []types.ToolCall `json:"tool_calls,omitempty"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	Thinking  string           `json:"thinking,omitempty"` // 支持深度思考模式
	ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
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
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Convert model ID to model name for ServiceRequest
	originalModel := req.Model
	modelInfo := e.GetModelById(ctx, req.Model)
	modelName := modelInfo.Name

	if modelName == "" {
		modelName = originalModel
	}

	// Debug log to trace model conversion
	fmt.Printf("[Chat] Model conversion: %s -> %s\n", originalModel, modelName)

	hybridPolicy := "default"
	ds := datastore.GetDefaultDatastore()
	sp := &types.Service{
		Name:   "chat",
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
		Service:      "chat",
		Model:        originalModel,
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

		var oadinResp types.OadinAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &oadinResp); err == nil && oadinResp.BusinessCode == 10000 {

			dataBytes, err := json.Marshal(oadinResp.Data)
			if err != nil {
				return nil, fmt.Errorf("解析Oadin响应data字段失败: %v", err)
			}

			var chatResp types.OadinChatResponse
			if err := json.Unmarshal(dataBytes, &chatResp); err != nil {
				return nil, fmt.Errorf("解析Oadin聊天响应结构失败: %v", err)
			}

			if len(chatResp.Choices) == 0 {
				return nil, fmt.Errorf("Oadin响应中没有choices选项")
			}

			response := &types.ChatResponse{
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
				fmt.Printf("[Chat] Oadin API检测到思考内容，长度：%d\n", len(response.Thoughts))
			}
			return response, nil
		}

		// 尝试直接解析成完整的ChatResponse
		var response types.ChatResponse
		if err := json.Unmarshal(result.HTTP.Body, &response); err == nil && response.Content != "" {
			fmt.Printf("[Chat] 直接解析成功，内容长度：%d\n", len(response.Content))
			return &response, nil
		}

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

			// 提取思考内容(如果有)
			var thoughts string
			if ollamaResp.Message != nil && ollamaResp.Message.Thinking != "" {
				thoughts = ollamaResp.Message.Thinking
				fmt.Printf("[Chat] 提取到思考内容，长度: %d\n", len(thoughts))
			}

			// 提取工具调用(如果有)
			var toolCalls []types.ToolCall
			if ollamaResp.Message != nil && ollamaResp.Message.ToolCalls != nil && len(ollamaResp.Message.ToolCalls) > 0 {
				toolCalls = ollamaResp.Message.ToolCalls
				fmt.Printf("[Chat] 提取到工具调用，数量: %d\n", len(toolCalls))
			} else if ollamaResp.ToolCalls != nil && len(ollamaResp.ToolCalls) > 0 {
				toolCalls = ollamaResp.ToolCalls
				fmt.Printf("[Chat] 提取到工具调用，数量: %d\n", len(toolCalls))
			}

			// 创建响应对象
			resp := &types.ChatResponse{
				Content:       content,
				Model:         model,
				IsComplete:    isComplete,
				Object:        "chat.completion",
				ToolCalls:     toolCalls,
				Thoughts:      thoughts,
				TotalDuration: ollamaResp.TotalDuration,
			}
			fmt.Printf("[Chat] 兼容Ollama格式解析成功，内容长度: %d\n", len(content))
			return resp, nil
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
		var toolCalls []types.ToolCall
		var thoughts string
		var totalDuration int64

		// 尝试提取message.content (Ollama /api/chat)
		if msg, ok := data["message"].(map[string]interface{}); ok {
			// 提取content
			if c, ok := msg["content"].(string); ok {
				content = c
				fmt.Printf("[Chat] 从message.content提取内容，长度: %d\n", len(content))
			}

			// 提取thinking
			if th, ok := msg["thinking"].(string); ok {
				thoughts = th
				fmt.Printf("[Chat] 从message.thinking提取思考内容，长度: %d\n", len(thoughts))
			}

			// 提取tool_calls
			if tc, ok := msg["tool_calls"].([]types.ToolCall); ok && len(tc) > 0 {
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

		// 提取处理时间
		if td, ok := data["total_duration"].(float64); ok {
			totalDuration = int64(td)
		}

		// 若还没有提取到工具调用，尝试从顶层提取
		if len(toolCalls) == 0 {
			if tc, ok := data["tool_calls"].([]types.ToolCall); ok && len(tc) > 0 {
				toolCalls = tc
				fmt.Printf("[Chat] 从顶层提取到tool_calls，数量: %d\n", len(toolCalls))
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
		resp := &types.ChatResponse{
			Content:       content,
			Model:         model,
			IsComplete:    isComplete,
			Object:        "chat.completion",
			ToolCalls:     toolCalls,
			Thoughts:      thoughts,
			TotalDuration: totalDuration,
		}
		fmt.Printf("[Chat] 通用解析成功，内容长度: %d\n", len(content))
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
