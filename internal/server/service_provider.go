package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils/bcode"
)

type ServiceProvider interface {
	CreateServiceProvider(ctx context.Context, request *dto.CreateServiceProviderRequest) (*dto.CreateServiceProviderResponse, error)
	DeleteServiceProvider(ctx context.Context, request *dto.DeleteServiceProviderRequest) (*dto.DeleteServiceProviderResponse, error)
	UpdateServiceProvider(ctx context.Context, request *dto.UpdateServiceProviderRequest) (*dto.UpdateServiceProviderResponse, error)
	GetServiceProvider(ctx context.Context, request *dto.GetServiceProviderRequest) (*dto.GetServiceProviderResponse, error)
	GetServiceProviders(ctx context.Context, request *dto.GetServiceProvidersRequest) (*dto.GetServiceProvidersResponse, error)
}

type ServiceProviderImpl struct{}

func NewServiceProvider() ServiceProvider {
	return &ServiceProviderImpl{}
}

func (s *ServiceProviderImpl) CreateServiceProvider(ctx context.Context, request *dto.CreateServiceProviderRequest) (*dto.CreateServiceProviderResponse, error) {
	ds := datastore.GetDefaultDatastore()

	sp := &types.ServiceProvider{}
	sp.ProviderName = request.ProviderName

	isExist, err := ds.IsExist(ctx, sp)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, bcode.ErrAIGCServiceProviderIsExist
	}

	sp.ServiceName = request.ServiceName
	sp.ServiceSource = request.ServiceSource
	sp.Flavor = request.ApiFlavor
	sp.AuthType = request.AuthType
	sp.AuthKey = request.AuthKey
	sp.Desc = request.Desc
	sp.Method = request.Method
	sp.URL = request.Url
	sp.ExtraHeaders = request.ExtraHeaders
	sp.ExtraJSONBody = request.ExtraJsonBody
	sp.Properties = request.Properties
	sp.CreatedAt = time.Now()
	sp.UpdatedAt = time.Now()

	modelIsExist := make(map[string]bool)

	// todo 如果是本地，models不为空，检查model是否存在，不存在则直接拉取
	if request.ServiceSource == types.ServiceSourceLocal {
		engineProvider := provider.GetModelEngine(request.ApiFlavor)
		err := engineProvider.HealthCheck()
		if err != nil {
			return nil, err
		}

		modelList, err := engineProvider.ListModels(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range modelList.Models {
			for _, mName := range request.Models {
				if v.Name == mName {
					modelIsExist[mName] = true
				} else if _, ok := modelIsExist[mName]; !ok {
					modelIsExist[mName] = false
				}
			}
		}

		for _, mName := range request.Models {
			if !modelIsExist[mName] {
				slog.Info("The model " + mName + " does not exist, ready to start pulling the model.")
				stream := false
				pullReq := &types.PullModelRequest{
					Model:  mName,
					Stream: &stream,
				}

				slog.Info("Pull model  start ..." + mName)
				_, err := engineProvider.PullModel(ctx, pullReq, nil)
				if err != nil {
					slog.Error("Pull model error: ", err.Error())
					return nil, bcode.ErrEnginePullModel
				}
				slog.Info("Pull model %s completed ..." + mName)
			}
		}
	}

	for _, mName := range request.Models {
		server := ChooseCheckServer(*sp, mName)
		checkRes := server.CheckServer()
		if !checkRes {
			return nil, bcode.ErrProviderIsUnavailable
		}
	}

	err = ds.Add(ctx, sp)
	if err != nil {
		return nil, err
	}

	for _, mName := range request.Models {
		model := &types.Model{
			ModelName:    mName,
			ProviderName: request.ProviderName,
			Status:       "downloaded",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = ds.Add(ctx, model)
		if err != nil {
			return nil, err
		}
	}

	return &dto.CreateServiceProviderResponse{
		Bcode: *bcode.ServiceProviderCode,
	}, nil
}

func (s *ServiceProviderImpl) DeleteServiceProvider(ctx context.Context, request *dto.DeleteServiceProviderRequest) (*dto.DeleteServiceProviderResponse, error) {
	sp := new(types.ServiceProvider)
	if request.ProviderName == "" {
		return nil, bcode.ErrProviderInvalid
	}
	sp.ProviderName = request.ProviderName

	ds := datastore.GetDefaultDatastore()
	err := ds.Get(ctx, sp)
	if err != nil {
		return nil, err
	}

	// 删除provider下模型
	m := new(types.Model)
	m.ProviderName = request.ProviderName
	list, err := ds.List(ctx, m, &datastore.ListOptions{
		Page:     0,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}

	if sp.ServiceSource == types.ServiceSourceLocal {
		// 删除本地已下载模型 , 需检查本地模型是否被其他service provider共同引用，若有，则不删本地模型，只删记录
		engine := provider.GetModelEngine(sp.Flavor)
		for _, m := range list {
			dsModel := m.(*types.Model)
			tmpModel := &types.Model{
				ModelName: dsModel.ModelName,
			}
			count, err := ds.Count(ctx, tmpModel, &datastore.FilterOptions{})
			if err != nil || count > 1 {
				continue
			}

			delReq := &types.DeleteRequest{Model: dsModel.ModelName}

			err = engine.DeleteModel(ctx, delReq)
			if err != nil {
				return nil, err
			}
		}
	}

	err = ds.Delete(ctx, m)
	if err != nil {
		return nil, err
	}

	err = ds.Delete(ctx, sp)
	if err != nil {
		return nil, err
	}

	// 检查当前设定的local和remote service provider, 若是则置空
	service := &types.Service{Name: sp.ServiceName}
	err = ds.Get(ctx, service)
	if err != nil {
		return nil, err
	}
	if sp.ServiceSource == types.ServiceSourceRemote && sp.ProviderName == service.RemoteProvider {
		service.RemoteProvider = ""
		if service.LocalProvider == "" {
			service.Status = 0
		}
	} else if sp.ServiceSource == types.ServiceSourceLocal && sp.ProviderName == service.LocalProvider {
		service.LocalProvider = ""
		if service.RemoteProvider == "" {
			service.Status = 0
		}
	}

	err = ds.Put(ctx, service)
	if err != nil {
		return nil, err
	}

	return &dto.DeleteServiceProviderResponse{
		Bcode: *bcode.ServiceProviderCode,
	}, nil
}

func (s *ServiceProviderImpl) UpdateServiceProvider(ctx context.Context, request *dto.UpdateServiceProviderRequest) (*dto.UpdateServiceProviderResponse, error) {
	ds := datastore.GetDefaultDatastore()
	sp := &types.ServiceProvider{}
	sp.ProviderName = request.ProviderName

	err := ds.Get(ctx, sp)
	if err != nil {
		return nil, err
	}

	if request.ServiceName != "" {
		sp.ServiceName = request.ServiceName
	}
	if request.ServiceSource != "" {
		sp.ServiceSource = request.ServiceSource
	}
	if request.ApiFlavor != "" {
		sp.Flavor = request.ApiFlavor
	}
	if request.AuthType != "" {
		sp.AuthType = request.AuthType
	}
	if request.AuthKey != "" {
		sp.AuthKey = request.AuthKey
	}
	if request.Desc != "" {
		sp.Desc = request.Desc
	}
	if request.Method != "" {
		sp.Method = request.Method
	}
	if request.Url != "" {
		sp.URL = request.Url
	}
	if request.ExtraHeaders != "" {
		sp.ExtraHeaders = request.ExtraHeaders
	}
	if request.ExtraJsonBody != "" {
		sp.ExtraJSONBody = request.ExtraJsonBody
	}
	if request.Properties != "" {
		sp.Properties = request.Properties
	}
	sp.UpdatedAt = time.Now()

	model := types.Model{ProviderName: sp.ProviderName}
	err = ds.Get(ctx, &model)
	if err != nil {
		return nil, err
	}

	server := ChooseCheckServer(*sp, model.ModelName)
	checkRes := server.CheckServer()
	if !checkRes {
		return nil, bcode.ErrProviderIsUnavailable
	}

	err = ds.Put(ctx, sp)
	if err != nil {
		return nil, err
	}

	return &dto.UpdateServiceProviderResponse{
		Bcode: *bcode.ServiceProviderCode,
	}, nil
}

func (s *ServiceProviderImpl) GetServiceProvider(ctx context.Context, request *dto.GetServiceProviderRequest) (*dto.GetServiceProviderResponse, error) {
	// todo 实际实现逻辑
	return &dto.GetServiceProviderResponse{}, nil
}

func (s *ServiceProviderImpl) GetServiceProviders(ctx context.Context, request *dto.GetServiceProvidersRequest) (*dto.GetServiceProvidersResponse, error) {
	sp := new(types.ServiceProvider)
	sp.ServiceName = request.ServiceName
	sp.ProviderName = request.ProviderName
	sp.Flavor = request.ApiFlavor
	sp.ServiceSource = request.ServiceSource

	ds := datastore.GetDefaultDatastore()
	list, err := ds.List(ctx, sp, &datastore.ListOptions{Page: 0, PageSize: 100})
	if err != nil {
		return nil, err
	}
	var spNames []string
	for _, v := range list {
		dsProvider := v.(*types.ServiceProvider)
		spNames = append(spNames, dsProvider.ProviderName)
	}

	inOptions := make([]datastore.InQueryOption, 0)
	inOptions = append(inOptions, datastore.InQueryOption{
		Key:    "provider_name",
		Values: spNames,
	})
	m := new(types.Model)
	mList, err := ds.List(ctx, m, &datastore.ListOptions{
		FilterOptions: datastore.FilterOptions{
			In: inOptions,
		},
		Page:     0,
		PageSize: 10,
	})
	if err != nil {
		return nil, err
	}

	spModels := make(map[string][]string)
	for _, v := range mList {
		dsModel := v.(*types.Model)
		spModels[dsModel.ProviderName] = append(spModels[dsModel.ProviderName], dsModel.ModelName)
	}

	respData := make([]dto.ServiceProvider, 0)
	for _, v := range list {
		dsProvider := v.(*types.ServiceProvider)
		tmp := &dto.ServiceProvider{
			ProviderName:  dsProvider.ProviderName,
			ServiceName:   dsProvider.ServiceName,
			ServiceSource: dsProvider.ServiceSource,
			Desc:          dsProvider.Desc,
			AuthType:      dsProvider.AuthType,
			AuthKey:       dsProvider.AuthKey,
			Flavor:        dsProvider.Flavor,
			Properties:    dsProvider.Properties,
			Status:        dsProvider.Status,
			CreatedAt:     dsProvider.CreatedAt,
			UpdatedAt:     dsProvider.UpdatedAt,
		}

		if models, ok := spModels[dsProvider.ProviderName]; ok {
			tmp.Models = models
		}

		respData = append(respData, *tmp)
	}

	return &dto.GetServiceProvidersResponse{
		Bcode: *bcode.ServiceProviderCode,
		Data:  respData,
	}, nil
}

type ModelServiceManager interface {
	CheckServer() bool
}

type CheckModelsServer struct {
	ServiceProvider types.ServiceProvider
}
type CheckChatServer struct {
	ServiceProvider types.ServiceProvider
	ModelName       string
}
type CheckGenerateServer struct {
	ServiceProvider types.ServiceProvider
	ModelName       string
}

type CheckEmbeddingServer struct {
	ServiceProvider types.ServiceProvider
}

func (m *CheckModelsServer) CheckServer() bool {
	client := &http.Client{}
	req, err := http.NewRequest(m.ServiceProvider.Method, m.ServiceProvider.URL, nil)
	if err != nil {
		return false
	}
	if m.ServiceProvider.AuthType != "none" {
		req.Header.Set("Authorization", "Bearer "+m.ServiceProvider.AuthKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("[Schedule] Failed to request", "error", resp.StatusCode)
		return false
	}
	return true
}

func (c *CheckChatServer) CheckServer() bool {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport}
	type Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type RequestBody struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}

	requestBody := RequestBody{
		Model:  c.ModelName,
		Stream: false,
		Messages: []Message{
			{
				Role:    "user",
				Content: "你好！",
			},
		},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(c.ServiceProvider.Method, c.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		slog.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	if c.ServiceProvider.AuthType != "none" {
		req.Header.Set("Authorization", "Bearer "+c.ServiceProvider.AuthKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("[Schedule] Failed to request", "error", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("[Schedule] Failed to request", "error", resp.StatusCode)
		return false
	}
	return true
}

func (g *CheckGenerateServer) CheckServer() bool {
	return false
}

func (e *CheckEmbeddingServer) CheckServer() bool {
	return false
}

func ChooseCheckServer(sp types.ServiceProvider, modelName string) ModelServiceManager {
	var server ModelServiceManager
	switch sp.ServiceName {
	case types.ServiceModels:
		server = &CheckModelsServer{ServiceProvider: sp}
	case types.ServiceChat:
		server = &CheckChatServer{ServiceProvider: sp, ModelName: modelName}
	case types.ServiceGenerate:
		server = &CheckGenerateServer{ServiceProvider: sp, ModelName: modelName}
	case types.ServiceEmbed:
		server = &CheckEmbeddingServer{ServiceProvider: sp}
	default:
		slog.Error("[Schedule] Unknown service name", "error", sp.ServiceName)
		return nil
	}
	return server
}
