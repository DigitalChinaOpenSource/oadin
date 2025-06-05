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

func QueryCloudSupplierJson(ctx context.Context, page, pageSize int) (*QueryCloudSupplierJsonRespond, error) {
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
	return &resp, nil
}

func GetSuppliers(ctx context.Context, resp *QueryCloudSupplierJsonRespond) (*dto.ImportServiceRequest, error) {
	requestModel := &dto.ImportServiceRequest{
		Services:         make(map[string]dto.ServiceEntry),
		ServiceProviders: make(map[string]dto.ServiceProviderEntry),
		Version:          resp.Data.Version,
	}
	services := resp.Data.Service
	ServiceProviders := resp.Data.ServiceProvider

	for _, service := range services {
		requestModel.Services[service.Name] = dto.ServiceEntry{
			HybridPolicy: service.HybridPolicy,
			ServiceProviders: dto.ServiceProviderInfo{
				Remote: service.RemoteProvider,
				Local:  service.LocalProvider,
			},
		}
	}
	for _, serviceProvider := range ServiceProviders {
		requestModel.ServiceProviders[serviceProvider.ProviderName] = dto.ServiceProviderEntry{
			ServiceName:   serviceProvider.ServiceName,
			ServiceSource: serviceProvider.ServiceSource,
			Desc:          serviceProvider.Desc,
			APIFlavor:     serviceProvider.ApiFlavor,
			Method:        serviceProvider.Method,
			AuthType:      serviceProvider.AuthType,
			AuthKey:       serviceProvider.AuthKey,
			Models:        serviceProvider.Models,
		}
	}
	return requestModel, nil
}

// RemoteServiceMap, err := vega.GetModels(ctx, "remote")
func GetModels(ctx context.Context, deployMode string) (map[string][]dto.LocalSupportModelData, error) {
	//models, err := QueryCloudModelJson(ctx, deployMode)
	//if err != nil {
	//	return getLocalModelList(deployMode)
	//}
	//res, err := getRemoteModels(models)
	//if err != nil {
	//	return getLocalModelList(deployMode)
	//}
	//return res, nil
	return getLocalModelList(deployMode)
}

/*
	localOllamaServiceMap, err := vega.GetModels(ctx, "local")
	if err != nil {
		fmt.Printf("GetModels failed: %v\n", err)
		return nil, err
	}

	替换
	localOllamaServiceMap := make(map[string][]dto.LocalSupportModelData)
	fileContent, err := template.FlavorTemplateFs.ReadFile("local_model.json")
	if err != nil {
		fmt.Printf("Read file failed: %v\n", err)
		return nil, err
	}
	// parse struct
	err = json.Unmarshal(fileContent, &localOllamaServiceMap)
	if err != nil {
		fmt.Printf("Parse JSON failed: %v\n", err)
		return nil, err
	}

	RemoteServiceMap, err := vega.GetModels(ctx, "remote")
	if err != nil {
		fmt.Printf("GetModels failed: %v\n", err)
		return nil, err
	}
	替换
		RemoteServiceMap := make(map[string][]dto.LocalSupportModelData)
		fileContent, err := template.FlavorTemplateFs.ReadFile("remote_model.json")
		if err != nil {
			fmt.Printf("Read file failed: %v\n", err)
			return nil, err
		}
		// parse struct
		err = json.Unmarshal(fileContent, &RemoteServiceMap)
		if err != nil {
			fmt.Printf("Parse JSON failed: %v\n", err)
			return nil, err
		}
*/

var typeMapClass = map[string][]string{
	"chat":          []string{"文本生成"},
	"embed":         []string{"文本向量化"},
	"text_to_image": []string{"文生图"},
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
		class, _ := typeMapClass[model.Type]
		flavor, ok := localFlavorMap[model.Flavor]
		if model.OllamaId != "" {
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
