package vega

import (
	"byze/internal/api/dto"
	"context"
	"reflect"
	"testing"
	"time"
)

func TestQueryCloudModelJson(t *testing.T) {
	data, err := QueryCloudModelJson(context.Background(), "")
	if err != nil {
		t.Errorf("QueryCloudModelJson() error = %v", err)
		return
	}
	if len(data) == 0 {
		t.Errorf("QueryCloudModelJson() got = %v, want not empty", data)
	}
}

func TestGetRemoteModels(t *testing.T) {
	tests := []struct {
		name     string
		models   []Model
		expected map[string][]dto.LocalSupportModelData
	}{
		{
			name:     "empty input",
			models:   []Model{},
			expected: map[string][]dto.LocalSupportModelData{},
		},
		{
			name: "single model with known type",
			models: []Model{
				{
					Id:             "model1",
					SupplierId:     "supplier1",
					SupplierName:   "OpenAI",
					Name:           "GPT-4",
					Avatar:         "avatar1",
					Description:    "Advanced AI model",
					Type:           "chat",
					Tags:           []string{"ai", "chat"},
					DeployMode:     1,
					ParameterScale: 175.0,
					FileSize:       1024,
					MaxInput:       2048,
					MaxOutput:      2048,
					Popularity:     100,
					CreateAt:       time.Now(),
					UpdateAt:       time.Now(),
					CreateBy:       "admin",
					UpdateBy:       "admin",
				},
			},
			expected: map[string][]dto.LocalSupportModelData{
				"chat": {
					{
						OllamaId:    "model1",
						Name:        "GPT-4",
						Avatar:      "avatar1",
						Description: "Advanced AI model",
						Class:       "文本生成",
						Flavor:      "OpenAI",
						Size:        "1024",
						ParamsSize:  175.0,
					},
				},
			},
		},
		{
			name: "multiple models with different types",
			models: []Model{
				{
					Id:             "model1",
					Type:           "chat",
					SupplierName:   "OpenAI",
					Name:           "GPT-4",
					FileSize:       1024,
					ParameterScale: 175.0,
				},
				{
					Id:             "model2",
					Type:           "embed",
					SupplierName:   "Cohere",
					Name:           "Embedder",
					FileSize:       512,
					ParameterScale: 100.0,
				},
				{
					Id:             "model3",
					Type:           "unknown",
					SupplierName:   "Unknown",
					Name:           "Unknown",
					FileSize:       256,
					ParameterScale: 50.0,
				},
			},
			expected: map[string][]dto.LocalSupportModelData{
				"chat": {
					{
						OllamaId:   "model1",
						Name:       "GPT-4",
						Class:      "文本生成",
						Flavor:     "OpenAI",
						Size:       "1024",
						ParamsSize: 175.0,
					},
				},
				"embed": {
					{
						OllamaId:   "model2",
						Name:       "Embedder",
						Class:      "文本向量化",
						Flavor:     "Cohere",
						Size:       "512",
						ParamsSize: 100.0,
					},
				},
				"unknown": {
					{
						OllamaId:   "model3",
						Name:       "Unknown",
						Class:      "未知",
						Flavor:     "Unknown",
						Size:       "256",
						ParamsSize: 50.0,
					},
				},
			},
		},
		{
			name: "multiple models with same type",
			models: []Model{
				{
					Id:             "model1",
					Type:           "chat",
					SupplierName:   "OpenAI",
					Name:           "GPT-4",
					FileSize:       1024,
					ParameterScale: 175.0,
				},
				{
					Id:             "model2",
					Type:           "chat",
					SupplierName:   "Anthropic",
					Name:           "Claude",
					FileSize:       2048,
					ParameterScale: 200.0,
				},
			},
			expected: map[string][]dto.LocalSupportModelData{
				"chat": {
					{
						OllamaId:   "model1",
						Name:       "GPT-4",
						Class:      "文本生成",
						Flavor:     "OpenAI",
						Size:       "1024",
						ParamsSize: 175.0,
					},
					{
						OllamaId:   "model2",
						Name:       "Claude",
						Class:      "文本生成",
						Flavor:     "Anthropic",
						Size:       "2048",
						ParamsSize: 200.0,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRemoteModels(tt.models)
			if err != nil {
				t.Errorf("GetRemoteModels() error = %v, wantErr nil", err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetRemoteModels() = %v, want %v", result, tt.expected)
			}
		})
	}
}
