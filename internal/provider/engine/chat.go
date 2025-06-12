package engine

import (
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	serviceReq := &types.ServiceRequest{
		Service:    "chat",
		Model:      req.Model,
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
		var apiResp ollamaAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &apiResp); err != nil {
			return nil, err
		}
		resp := &types.ChatResponse{
			Content: apiResp.Message.Content,
			// 其它字段可按需补充
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
