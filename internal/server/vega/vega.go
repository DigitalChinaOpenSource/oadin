package vega

import (
	"byze/internal/api/dto"
	"byze/internal/utils/client"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type VegaClient struct {
	client.Client
}

func NewVegaClient() *VegaClient {
	// default host
	host := "127.0.0.1:16677"

	// default scheme
	scheme := "http"

	return &VegaClient{
		Client: *client.NewClient(&url.URL{
			Scheme: scheme,
			Host:   host,
		}, http.DefaultClient),
	}
}

func QueryCloudModelJson(ctx context.Context, hybridPolicy string) ([]Model, error) {
	c := NewVegaClient()
	routerPath := fmt.Sprintf("/vega/%s/service", "0.1")

	req := QueryCloudModelJsonRequest{
		HybridPolicy: hybridPolicy,
	}
	resp := QueryCloudModelJsonRespond{}

	err := c.Client.Do(ctx, http.MethodPost, routerPath, req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code > 200 {
		fmt.Println(resp.Message)
		return nil, fmt.Errorf(resp.Message)
	}
	return resp.Data, nil
}

type LocalSupportModelData struct {
	OllamaId    string  `json:"id"`
	Name        string  `json:"name"`
	Avatar      string  `json:"avatar"`
	Description string  `json:"description"`
	Class       string  `json:"class"`
	Flavor      string  `json:"provider"`
	Size        string  `json:"size"`
	ParamsSize  float32 `json:"params_size"`
}

var typeMapClass = map[string]string{
	"chat":          "文本生成",
	"embed":         "文本向量化",
	"text_to_image": "文生图",
}

func GetRemoteModels(models []Model) (map[string][]dto.LocalSupportModelData, error) {
	RemoteServiceMap := make(map[string][]dto.LocalSupportModelData)
	for _, model := range models {
		size := strconv.Itoa(model.FileSize)
		class, ok := typeMapClass[model.Type]
		if !ok {
			class = "未知"
		}
		flavor := model.SupplierName
		LocalSupportModel := dto.LocalSupportModelData{
			OllamaId:    model.Id,
			Name:        model.Name,
			Avatar:      model.Avatar,
			Description: model.Description,
			Class:       class,
			Flavor:      flavor,
			Size:        size,
			ParamsSize:  model.ParameterScale,
		}
		if _, ok := RemoteServiceMap[model.Type]; !ok {
			RemoteServiceMap[model.Type] = []dto.LocalSupportModelData{
				LocalSupportModel,
			}
		} else {
			RemoteServiceMap[model.Type] = append(RemoteServiceMap[model.Type], LocalSupportModel)
		}

	}
	return RemoteServiceMap, nil
}
