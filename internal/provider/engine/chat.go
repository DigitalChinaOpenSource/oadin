package engine

import (
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
	"io/ioutil"
)

type Engine struct{}

type ollamaAPIResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

func NewEngine() *Engine {
	return &Engine{}
}

// getModelNameById converts a model ID to its corresponding name
func getModelNameById(modelId string) string {
	// 1. First check local model json
	data, err := ioutil.ReadFile("internal/provider/template/local_model.json")
	if err == nil {
		var local struct {
			Chat []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"chat"`
			Embed []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"embed"`
		}
		if json.Unmarshal(data, &local) == nil {
			for _, m := range local.Chat {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
			for _, m := range local.Embed {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
		}
	}
	// 2. Check support_model.json
	data, err = ioutil.ReadFile("internal/datastore/jsonds/data/support_model.json")
	if err == nil {
		var arr []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &arr) == nil {
			for _, m := range arr {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
		}
	}
	return modelId // Return the original ID if no name is found
}

func (e *Engine) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Convert model ID to model name for ServiceRequest
	modelName := getModelNameById(req.Model)

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
		var apiResp ollamaAPIResponse
		if err := json.Unmarshal(result.HTTP.Body, &apiResp); err != nil {
			return nil, err
		}
		resp := &types.ChatResponse{
			Content: apiResp.Message.Content,
		}
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
