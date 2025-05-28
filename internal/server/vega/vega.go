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
	host := "127.0.0.1:3000"

	// default scheme
	scheme := "http"

	return &VegaClient{
		Client: *client.NewClient(&url.URL{
			Scheme: scheme,
			Host:   host,
		}, http.DefaultClient),
	}
}

func QueryCloudModelJson(ctx context.Context, hybridPolicy string, page, pageSize int) ([]Model, error) {
	c := NewVegaClient()
	routerPath := "/api/llm/model/search/"

	req := QueryCloudModelJsonRequest{
		HybridPolicy: hybridPolicy,
		Page:         page,
		PageSize:     pageSize,
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
	return resp.Data.List, nil
}

func QueryCloudSupplierJson(ctx context.Context, page, pageSize int) ([]Supplier, error) {
	c := NewVegaClient()
	routerPath := "/api/llm/supplier/search/"

	req := QueryCloudSupplierJsonRequest{
		Page:     page,
		PageSize: pageSize,
	}
	resp := QueryCloudSupplierJsonRespond{}

	err := c.Client.Do(ctx, http.MethodPost, routerPath, req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code > 200 {
		fmt.Println(resp.Message)
		return nil, fmt.Errorf(resp.Message)
	}
	return resp.Data.List, nil
}

var typeMapClass = map[string]string{
	"chat":          "文本生成",
	"embed":         "文本向量化",
	"text_to_image": "文生图",
}

func GetRemoteModels(models []Model) (map[string][]dto.LocalSupportModelData, error) {
	RemoteServiceMap := make(map[string][]dto.LocalSupportModelData)
	for _, model := range models {
		parameterScale, err := strconv.Atoi(model.ParameterScale)
		if err != nil {
			parameterScale = 0
		}
		class, ok := typeMapClass[model.Type]
		if !ok {
			class = "未知"
		}
		flavor := model.Flavor
		LocalSupportModel := dto.LocalSupportModelData{
			OllamaId:    model.Id,
			Name:        model.Name,
			Avatar:      model.Avatar,
			Description: model.Description,
			Class:       class,
			Flavor:      flavor,
			Size:        model.FileSize,
			ParamsSize:  float32(parameterScale),
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
