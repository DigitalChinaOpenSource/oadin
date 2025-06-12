package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"byze/internal/types"

	"github.com/google/uuid"
)

func (o *OllamaProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	c := o.GetDefaultClient()

	// 处理思考模式
	var thoughts string

	// 检查options中是否启用了thinking
	thinkingEnabled := false
	if req.Options != nil {
		if thinking, ok := req.Options["thinking"].(bool); ok && thinking {
			thinkingEnabled = true
		}
	}

	if thinkingEnabled {
		// 创建一个思考请求
		thinkingReq := *req

		// 在options中设置thinking为true
		if thinkingReq.Options == nil {
			thinkingReq.Options = make(map[string]any)
		}
		thinkingReq.Options["thinking"] = true

		// 执行思考请求
		var thinkResponse map[string]interface{}
		if err := c.Do(ctx, http.MethodPost, "/api/chat", &thinkingReq, &thinkResponse); err != nil {
			slog.Error("Failed to perform thinking", "error", err)
			// 思考失败不影响主流程，继续执行
		} else {
			// 提取思考内容
			thinking, ok := thinkResponse["thinking"].(string)
			if ok && thinking != "" {
				thoughts = thinking
				slog.Info("Got thinking response", "thinking", thoughts)
			} else {
				slog.Warn("No thinking content in response", "response", thinkResponse)
			}
		}
	}

	// 执行主请求
	var response map[string]interface{}
	if err := c.Do(ctx, http.MethodPost, "/api/chat", req, &response); err != nil {
		slog.Error("Failed to call Ollama chat API", "error", err)
		return nil, fmt.Errorf("failed to call model API: %v", err)
	}

	// 提取并转换响应
	content, ok := response["message"].(map[string]interface{})["content"].(string)
	if !ok {
		slog.Error("Invalid response format from Ollama", "response", response)
		return nil, fmt.Errorf("invalid response format from model")
	}

	res := &types.ChatResponse{
		ID:         uuid.New().String(),
		Object:     "chat.completion",
		Model:      req.Model,
		Content:    content,
		IsComplete: true,
		Thoughts:   thoughts,
	}

	toolCalls, ok := response["message"].(map[string]interface{})["tool_calls"].([]any)
	if ok {
		res.ToolCalls = toolCalls
	}
	return res, nil
}

func (o *OpenvinoProvider) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return nil, fmt.Errorf("chat not implemented for OpenVINO provider")
}
