//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/schedule"
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

type ServiceProviderImpl struct {
	Ds datastore.Datastore
}

func NewServiceProvider() ServiceProvider {
	return &ServiceProviderImpl{
		Ds: datastore.GetDefaultDatastore(),
	}
}

func (s *ServiceProviderImpl) CreateServiceProvider(ctx context.Context, request *dto.CreateServiceProviderRequest) (*dto.CreateServiceProviderResponse, error) {
	ds := datastore.GetDefaultDatastore()

	sp := &types.ServiceProvider{}
	sp.ProviderName = request.ProviderName

	// isExist, err := ds.IsExist(ctx, sp)
	// if err != nil {
	// 	return nil, err
	// }
	// if isExist {
	// 	return nil, bcode.ErrAIGCServiceProviderIsExist
	// }
	providerServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, request.ServiceName)

	sp.ServiceName = request.ServiceName
	sp.ServiceSource = request.ServiceSource
	sp.Flavor = request.ApiFlavor
	sp.AuthType = request.AuthType
	if request.AuthType != types.AuthTypeNone && request.AuthKey == "" {
		return nil, bcode.ErrProviderAuthInfoLost
	}
	sp.AuthKey = request.AuthKey
	sp.Desc = request.Desc
	sp.Method = request.Method
	sp.URL = request.Url
	sp.Status = 0
	if request.Url == "" {
		sp.URL = providerServiceInfo.RequestUrl
	}
	if request.Method == "" {
		sp.Method = "POST"
	}
	sp.ExtraHeaders = request.ExtraHeaders
	if request.ExtraHeaders == "" {
		sp.ExtraHeaders = providerServiceInfo.ExtraHeaders
	}
	if request.ExtraJsonBody == "" {
		sp.ExtraJSONBody = "{}"
	}
	if request.Properties == "" {
		sp.Properties = "{}"
	}
	sp.CreatedAt = time.Now()
	sp.UpdatedAt = time.Now()

	modelIsExist := make(map[string]bool)

	if request.ServiceSource == types.ServiceSourceLocal {
		engineProvider := provider.GetModelEngine(request.ApiFlavor)
		engineConfig := engineProvider.GetConfig()
		if strings.Contains(request.Url, engineConfig.Host) {
			parseUrl, err := url.Parse(request.Url)
			if err != nil {
				return nil, bcode.ErrProviderServiceUrlNotFormat
			}
			host := parseUrl.Host
			engineConfig.Host = host
		}
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
				m := new(types.Model)
				m.ModelName = mName
				m.ProviderName = request.ProviderName
				err = s.Ds.Get(ctx, m)
				if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
					// todo debug log output
					return nil, bcode.ErrServer
				} else if errors.Is(err, datastore.ErrEntityInvalid) {
					m.Status = "downloading"
					err = s.Ds.Put(ctx, m)
					if err != nil {
						return nil, bcode.ErrAddModel
					}
				}
				if m.Status == "failed" {
					m.Status = "downloading"
				}
				if err != nil {
				}
				go AsyncPullModel(sp, m, pullReq)
			}
		}
	} else if request.ServiceSource == types.ServiceSourceRemote {
		for _, mName := range request.Models {
			server := ChooseCheckServer(*sp, mName)
			if server == nil {
				// return nil, bcode.ErrProviderIsUnavailable
				continue
			}
			checkRes := server.CheckServer()
			if !checkRes {
				// return nil, bcode.ErrProviderIsUnavailable
				continue
			}

			model := &types.Model{
				ModelName:    mName,
				ProviderName: request.ProviderName,
				Status:       "downloaded",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			err := ds.Put(ctx, model)
			if err != nil {
				return nil, err
			}
		}
		sp.Status = 1
	}

	err := ds.Put(ctx, sp)
	if err != nil {
		return nil, err
	}
	if request.ServiceName == types.ServiceChat {
		generateSp := &types.ServiceProvider{}
		generateSp.ProviderName = request.ProviderName

		generateSpIsExist, err := ds.IsExist(ctx, generateSp)
		if err != nil {
			return nil, err
		}
		if !generateSpIsExist {
			generateProviderServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, strings.Replace(request.ServiceName, "chat", "generate", -1))

			generateSp.ServiceName = strings.Replace(request.ServiceName, "chat", "generate", -1)
			generateSp.ServiceSource = request.ServiceSource
			generateSp.Flavor = request.ApiFlavor
			generateSp.AuthType = request.AuthType
			generateSp.AuthType = request.AuthType
			if request.AuthType != types.AuthTypeNone && request.AuthKey == "" {
				return nil, bcode.ErrProviderAuthInfoLost
			}
			generateSp.AuthKey = request.AuthKey
			generateSp.Desc = request.Desc
			generateSp.Method = request.Method
			generateSp.URL = generateProviderServiceInfo.RequestUrl
			generateSp.ExtraHeaders = request.ExtraHeaders
			if request.ExtraHeaders == "" {
				generateSp.ExtraHeaders = providerServiceInfo.ExtraHeaders
			}
			generateSp.ExtraJSONBody = request.ExtraJsonBody
			generateSp.Properties = request.Properties
			generateSp.CreatedAt = time.Now()
			generateSp.UpdatedAt = time.Now()
		}

	}

	err = DefaultProviderProcess(ctx, sp.ServiceName, sp.ServiceSource, sp.ProviderName)
	if err != nil {
		return nil, err
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
		// Delete the locally downloaded model.
		// It is necessary to check whether the local model is jointly referenced by other service providers.
		// If so, do not delete the local model but only delete the record.
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
			if dsModel.Status == "downloaded" {
				delReq := &types.DeleteRequest{Model: dsModel.ModelName}

				err = engine.DeleteModel(ctx, delReq)
				if err != nil {
					return nil, err
				}
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

	// Check the currently set local and remote service providers. If so, set them to empty.
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

	for _, modelName := range request.Models {
		model := types.Model{ProviderName: sp.ProviderName, ModelName: modelName, ServiceSource: sp.ServiceSource, ServiceName: sp.ServiceName}
		if request.ServiceSource == types.ServiceSourceLocal && request.ProviderName == types.FlavorOllama {
			model = types.Model{ProviderName: sp.ProviderName, ModelName: modelName}
		}

		err = ds.Get(ctx, &model)
		if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
			return nil, err
		}
		server := ChooseCheckServer(*sp, model.ModelName)
		if server == nil {
			model.Status = "failed"
			err = ds.Put(ctx, sp)
			if err != nil {
				return nil, err
			}
			if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
				return nil, err
			} else if errors.Is(err, datastore.ErrEntityInvalid) {
				err = ds.Add(ctx, &model)
				if err != nil {
					return nil, err
				}
			}
			err = ds.Put(ctx, &model)
			if err != nil {
				return nil, err
			}
			return nil, bcode.ErrProviderIsUnavailable
		}
		checkRes := server.CheckServer()
		if !checkRes {
			model.Status = "failed"
			err = ds.Put(ctx, sp)
			if err != nil {
				return nil, err
			}
			if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
				return nil, err
			} else if errors.Is(err, datastore.ErrEntityInvalid) {
				err = ds.Add(ctx, &model)
				if err != nil {
					return nil, err
				}
			}
			err = ds.Put(ctx, &model)
			if err != nil {
				return nil, err
			}
			return nil, bcode.ErrProviderIsUnavailable
		}
		err = ds.Get(ctx, &model)
		model.Status = "downloaded"
		if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
			return nil, err
		} else if errors.Is(err, datastore.ErrEntityInvalid) {
			err = ds.Add(ctx, &model)
			if err != nil {
				return nil, err
			}
		}
		err = ds.Put(ctx, &model)
		if err != nil {
			return nil, err
		}
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
		serviceProviderStatus := 0
		if dsProvider.ServiceSource == types.ServiceSourceRemote {
			model := types.Model{
				ProviderName: dsProvider.ProviderName,
			}
			err = ds.Get(ctx, &model)
			checkServerObj := ChooseCheckServer(*dsProvider, model.ModelName)
			status := checkServerObj.CheckServer()
			if status {
				serviceProviderStatus = 1
			}
		} else {
			providerEngine := provider.GetModelEngine(dsProvider.Flavor)
			err = providerEngine.HealthCheck()
			if err == nil {
				serviceProviderStatus = 1
			}
		}

		tmp := &dto.ServiceProvider{
			ProviderName:  dsProvider.ProviderName,
			ServiceName:   dsProvider.ServiceName,
			ServiceSource: dsProvider.ServiceSource,
			Desc:          dsProvider.Desc,
			AuthType:      dsProvider.AuthType,
			AuthKey:       dsProvider.AuthKey,
			Flavor:        dsProvider.Flavor,
			Properties:    dsProvider.Properties,
			Status:        serviceProviderStatus,
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
	ModelName       string
}

type CheckTextToImageServer struct {
	ServiceProvider types.ServiceProvider
	ModelName       string
}

func (m *CheckModelsServer) CheckServer() bool {
	req, err := http.NewRequest(m.ServiceProvider.Method, m.ServiceProvider.URL, nil)
	if err != nil {
		return false
	}
	status := CheckServerRequest(req, m.ServiceProvider, "")
	return status
}

func (c *CheckChatServer) CheckServer() bool {
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
		logger.LogicLogger.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(c.ServiceProvider.Method, c.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}

	status := CheckServerRequest(req, c.ServiceProvider, string(jsonData))
	return status
}

func (g *CheckGenerateServer) CheckServer() bool {
	return false
}

func (e *CheckEmbeddingServer) CheckServer() bool {
	type RequestBody struct {
		Model          string   `json:"model"`
		Input          []string `json:"input"`
		Inputs         []string `json:"inputs"`
		Dimensions     int      `json:"dimensions"`
		EncodingFormat string   `json:"encoding_format"`
	}
	requestBody := RequestBody{
		Model:          e.ModelName,
		Input:          []string{"test text"},
		Inputs:         []string{"test text"},
		Dimensions:     1024,
		EncodingFormat: "float",
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(e.ServiceProvider.Method, e.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}

	status := CheckServerRequest(req, e.ServiceProvider, string(jsonData))
	return status
}

func (e *CheckTextToImageServer) CheckServer() bool {
	prompt := "画一只小狗"
	var jsonData []byte
	var err error
	switch e.ServiceProvider.Flavor {
	case types.FlavorTencent:
		type RequestBody struct {
			Model      string `json:"model"`
			Prompt     string `json:"Prompt"`
			RspImgType string `json:"RspImgType"`
		}
		requestBody := RequestBody{
			Model:      e.ModelName,
			Prompt:     prompt,
			RspImgType: "url",
		}
		jsonData, err = json.Marshal(requestBody)
	case types.FlavorAliYun:
		type InputData struct {
			Prompt string `json:"prompt"`
		}
		type RequestBody struct {
			Model string    `json:"model"`
			Input InputData `json:"input"`
		}
		inputData := InputData{
			Prompt: prompt,
		}
		requestBody := RequestBody{
			Model: e.ModelName,
			Input: inputData,
		}
		jsonData, err = json.Marshal(requestBody)
	case types.FlavorBaidu:
		type RequestBody struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
		}
		requestBody := RequestBody{
			Model:  e.ModelName,
			Prompt: prompt,
		}
		jsonData, err = json.Marshal(requestBody)
	default:
		type RequestBody struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
		}
		requestBody := RequestBody{
			Model:  e.ModelName,
			Prompt: prompt,
		}
		jsonData, err = json.Marshal(requestBody)
	}
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(e.ServiceProvider.Method, e.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}

	status := CheckServerRequest(req, e.ServiceProvider, string(jsonData))
	return status
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
		server = &CheckEmbeddingServer{ServiceProvider: sp, ModelName: modelName}
	case types.ServiceTextToImage:
		server = &CheckTextToImageServer{ServiceProvider: sp, ModelName: modelName}
	default:
		logger.LogicLogger.Error("[Schedule] Unknown service name", "error", sp.ServiceName)
		return nil
	}
	return server
}

func CheckServerRequest(req *http.Request, serviceProvider types.ServiceProvider, reqBodyString string) bool {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	if serviceProvider.ExtraHeaders != "{}" {
		var extraHeader map[string]interface{}
		err := json.Unmarshal([]byte(serviceProvider.ExtraHeaders), &extraHeader)
		if err != nil {
			logger.LogicLogger.Error("Error parsing JSON:", err.Error())
			return false
		}
		for k, v := range extraHeader {
			req.Header.Set(k, v.(string))
		}

	}
	client := &http.Client{Transport: transport}
	req.Header.Set("Content-Type", "application/json")
	if serviceProvider.AuthType != "none" {
		authParams := &schedule.AuthenticatorParams{
			Request:      req,
			ProviderInfo: &serviceProvider,
			RequestBody:  reqBodyString,
		}
		authenticator := schedule.ChooseProviderAuthenticator(authParams)
		if authenticator == nil {
			return false
		}
		err := authenticator.Authenticate()
		if err != nil {
			return false
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.LogicLogger.Error("[Schedule] Failed to request", "error", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.LogicLogger.Error("[Schedule] Failed to request", "error", resp.StatusCode)
		return false
	}
	return true
}
