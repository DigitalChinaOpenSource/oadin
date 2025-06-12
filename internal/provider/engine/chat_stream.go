package engine

import (
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
)

func (e *Engine) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, <-chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)
	body, err := json.Marshal(req)
	if err != nil {
		go func() {
			errChan <- err
			close(respChan)
			close(errChan)
		}()
		return respChan, errChan
	}

	// Convert model ID to model name if needed
	modelName := getModelNameById(req.Model)

	serviceReq := &types.ServiceRequest{
		Service:       "generate",
		Model:         modelName, // Using model name instead of ID
		FromFlavor:    "ollama",
		AskStreamMode: true,
		HTTP: types.HTTPContent{
			Body: body,
		},
	}
	_, ch := schedule.GetScheduler().Enqueue(serviceReq)
	go func() {
		defer close(respChan)
		defer close(errChan)
		for result := range ch {
			if result.Error != nil {
				errChan <- result.Error
				return
			}
			var apiResp ollamaAPIResponse
			if err := json.Unmarshal(result.HTTP.Body, &apiResp); err != nil {
				errChan <- err
				return
			}
			resp := &types.ChatResponse{
				Content: apiResp.Message.Content,
			}
			respChan <- resp
		}
	}()
	return respChan, errChan
}
