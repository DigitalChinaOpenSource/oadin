package vega

import (
	"byze/config"
	"byze/internal/api/dto"
	"byze/internal/provider/template"
	"byze/internal/utils/client"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type VegaClient struct {
	client.Client
}

func NewVegaClient() *VegaClient {
	// default host
	host := config.ConfigRootInstance.Vega.Url

	// default scheme
	scheme := "http"

	return &VegaClient{
		Client: *client.NewClient(&url.URL{
			Scheme: scheme,
			Host:   host,
		}, http.DefaultClient),
	}
}

func QueryCloudModelJson(ctx context.Context, deployMode string) ([]Model, error) {
	c := NewVegaClient()
	routerPath := "/api/llm/model/search/"

	req := QueryCloudModelJsonRequest{
		DeployMode: deployMode,
	}
	resp := QueryCloudModelJsonRespond{}

	err := c.Client.Do(ctx, http.MethodPost, routerPath, req, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Code > 200 {
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
		return nil, fmt.Errorf(resp.Message)
	}
	return resp.Data.List, nil
}

func GetModels(ctx context.Context, deployMode string) (map[string][]dto.LocalSupportModelData, error) {
	models, err := QueryCloudModelJson(ctx, deployMode)
	if err != nil {
		return getLocalModelList(deployMode)
	}
	res, err := getRemoteModels(models)
	if err != nil {
		return getLocalModelList(deployMode)
	}
	return res, nil
}

var typeMapClass = map[string]string{
	"chat":          "文本生成",
	"embed":         "文本向量化",
	"text_to_image": "文生图",
}

var localFlavorMap = map[string]string{
	"deepseek":    "Deepseek",
	"aliyun":      "阿里巴巴",
	"lingyiwanwu": "Yi",
	"zhipuai":     "智谱",
	"nomic":       "Nomic",
	"baai":        "BAAI",
	"snowflake":   "Snowflake",
	"ibm":         "IBM",
	"openvino":    "openvino",
	"mixedbread":  "Mixedbread",
}

func getRemoteModels(models []Model) (map[string][]dto.LocalSupportModelData, error) {
	RemoteServiceMap := make(map[string][]dto.LocalSupportModelData)
	for _, model := range models {
		class, ok := typeMapClass[model.Type]
		if !ok {
			class = "未知"
		}
		flavor := model.Flavor
		if model.OllamaId != "" {
			flavor, ok = localFlavorMap[model.Flavor]
			if !ok {
				flavor = "未知"
			}
		}

		LocalSupportModel := dto.LocalSupportModelData{
			OllamaId:    model.OllamaId,
			Name:        model.Name,
			Avatar:      model.Avatar,
			Description: model.Description,
			Class:       class,
			Flavor:      flavor,
			Size:        model.FileSize,
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

func getLocalModelList(deployMode string) (map[string][]dto.LocalSupportModelData, error) {
	// 从本地读取文件
	if deployMode != "local" && deployMode != "remote" {
		return nil, fmt.Errorf("deployMode must be local or server")
	}
	fileName := deployMode + "_model.json"
	modelMap := make(map[string][]dto.LocalSupportModelData)
	fileContent, err := template.FlavorTemplateFs.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Read file failed: %v\n", err)
		return nil, err
	}
	// parse struct
	err = json.Unmarshal(fileContent, &modelMap)
	if err != nil {
		fmt.Printf("Parse JSON failed: %v\n", err)
		return nil, err
	}
	return modelMap, nil
}
