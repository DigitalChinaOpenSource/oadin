package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"byze/internal/api/dto"
	"byze/internal/datastore"
	"byze/internal/provider"
	"byze/internal/schedule"
	"byze/internal/types"
	"byze/internal/utils/bcode"
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

	isExist, err := ds.IsExist(ctx, sp)
	if err != nil {
		return nil, err
	}
	if isExist {
		return nil, bcode.ErrAIGCServiceProviderIsExist
	}
	providerServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, request.ServiceName)

	sp.ServiceName = request.ServiceName
	sp.ServiceSource = request.ServiceSource
	sp.Flavor = request.ApiFlavor
	sp.AuthType = request.AuthType
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
				m.ModelName = strings.ToLower(mName)
				m.ProviderName = request.ProviderName
				err = s.Ds.Get(ctx, m)
				if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
					// todo debug log output
					return nil, bcode.ErrServer
				} else if errors.Is(err, datastore.ErrEntityInvalid) {
					m.Status = "downloading"
					err = s.Ds.Add(ctx, m)
					if err != nil {
						return nil, bcode.ErrAddModel
					}
				}
				if m.Status == "failed" {
					m.Status = "downloading"
				}
				if err != nil {
				}
				go AsyncPullModel(ctx, sp, m, pullReq)
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

			err = ds.Add(ctx, model)
			if err != nil {
				return nil, err
			}
		}
		sp.Status = 1
	}

	err = ds.Add(ctx, sp)
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
				ModelName: strings.ToLower(dsModel.ModelName),
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
		if request.ApiFlavor == "smartvision" || sp.Flavor == "smartvision" {
			if sp.AuthKey != "" {
				var dbAuthInfoMap map[string]interface{}
				var requestInfoMap map[string]interface{}
				err = json.Unmarshal([]byte(request.AuthKey), &requestInfoMap)
				if err != nil {
					return nil, err
				}
				err = json.Unmarshal([]byte(sp.AuthKey), &dbAuthInfoMap)
				if err != nil {
					return nil, err
				}
				for k, v := range requestInfoMap {
					dbAuthInfoMap[k] = v
				}
				jsonBytes, err := json.Marshal(dbAuthInfoMap)
				if err != nil {
					fmt.Println("JSON 编码错误:", err)
					return nil, err
				}
				sp.AuthKey = string(jsonBytes)
			} else {
				sp.AuthKey = request.AuthKey
			}

		} else {
			sp.AuthKey = request.AuthKey
		}

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
		model := types.Model{ProviderName: sp.ProviderName, ModelName: modelName}
		if request.ServiceSource == types.ServiceSourceLocal {
			model = types.Model{ProviderName: sp.ProviderName, ModelName: strings.ToLower(modelName)}
		}

		err = ds.Get(ctx, &model)
		if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
			return nil, err
		}
		server := ChooseCheckServer(*sp, model.ModelName)
		if server == nil {
			return nil, bcode.ErrProviderIsUnavailable
		}
		checkRes := server.CheckServer()
		if !checkRes {
			return nil, bcode.ErrProviderIsUnavailable
		}
		model.Status = "downloaded"
		err = ds.Get(ctx, &model)
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
	providerName := request.ProviderName
	ds := datastore.GetDefaultDatastore()
	sp := &types.ServiceProvider{
		ProviderName: providerName,
	}
	err := ds.Get(ctx, sp)
	if err != nil {
		return nil, err
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 10
	}
	var supportModelList []dto.ProviderSupportModelData
	res := &dto.GetServiceProviderResponseData{}
	res.ServiceProvider = sp
	if sp.Flavor == types.FlavorSmartVision {
		smartvisionModelData, err := GetSmartVisionModelData(ctx, request.EnvType)
		if err != nil {
			return nil, err
		}
		res.TotalCount = len(smartvisionModelData)
		res.TotalPage = len(smartvisionModelData) / request.PageSize
		if res.TotalPage == 0 {
			res.TotalPage = 1
		}
		res.PageSize = request.PageSize
		res.Page = request.Page
		dataStart := (request.Page - 1) * request.PageSize
		dataEnd := request.Page * request.PageSize
		if dataEnd > len(smartvisionModelData) {
			dataEnd = len(smartvisionModelData) - 1
		}
		modelData := smartvisionModelData[dataStart:dataEnd]
		for _, model := range modelData {
			modelQuery := new(types.Model)
			modelQuery.ModelName = model.Name
			modelQuery.ProviderName = providerName
			isDownloaded := false
			err := ds.Get(context.Background(), modelQuery)
			if err != nil {
				isDownloaded = false
			}
			if modelQuery.Status != "downloaded" {
				isDownloaded = true
			}
			resModel := dto.ProviderSupportModelData{
				Name:         model.Name,
				Class:        model.Tags,
				Flavor:       model.Provider,
				ApiFlavor:    sp.Flavor,
				IsDownloaded: isDownloaded,
			}
			supportModelList = append(supportModelList, resModel)
		}
	} else {
		jds := datastore.GetDefaultJsonDatastore()
		sm := &types.SupportModel{}
		queryOpList := []datastore.FuzzyQueryOption{}
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "service_source",
			Query: sp.ServiceSource,
		})
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "api_flavor",
			Query: sp.Flavor,
		})
		options := &datastore.ListOptions{FilterOptions: datastore.FilterOptions{Queries: queryOpList}}
		totalCount, err := jds.Count(ctx, sm, &datastore.FilterOptions{Queries: queryOpList})
		if err != nil {
			return nil, err
		}
		res.TotalCount = int(totalCount)
		res.TotalPage = int(totalCount) / request.PageSize
		if res.TotalPage == 0 {
			res.TotalPage = 1
		}
		options.Page = request.Page
		options.PageSize = request.PageSize
		supportModel, err := jds.List(ctx, sm, options)
		if err != nil {
			return nil, err
		}
		for _, model := range supportModel {
			modelInfo := model.(*types.SupportModel)
			modelQuery := new(types.Model)
			modelQuery.ModelName = modelInfo.Name
			modelQuery.ProviderName = providerName
			isDownloaded := false
			err := ds.Get(context.Background(), modelQuery)
			if err != nil {
				isDownloaded = false
			}
			if modelQuery.Status == "downloaded" {
				isDownloaded = true
			}
			resModel := dto.ProviderSupportModelData{
				Name:         modelInfo.Name,
				Class:        modelInfo.Class,
				Flavor:       modelInfo.Flavor,
				ApiFlavor:    modelInfo.ApiFlavor,
				InputLength:  modelInfo.InputLength,
				OutputLength: modelInfo.OutputLength,
				ParamsSize:   modelInfo.ParamSize,
				IsDownloaded: isDownloaded,
			}
			supportModelList = append(supportModelList, resModel)
		}
	}
	res.SupportModelList = supportModelList
	return &dto.GetServiceProviderResponse{
		*bcode.ModelCode,
		*res,
	}, nil
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
			fmt.Println(1)
			//model := types.Model{
			//	ProviderName: dsProvider.ProviderName,
			//}
			//err = ds.Get(ctx, &model)
			//checkServerObj := ChooseCheckServer(*dsProvider, model.ModelName)
			//status := checkServerObj.CheckServer()
			//if status {
			//	serviceProviderStatus = 1
			//}
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
	content := types.HTTPContent{}
	status := CheckServerRequest(req, m.ServiceProvider, content)
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
		slog.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(c.ServiceProvider.Method, c.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		slog.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}
	content := types.HTTPContent{
		Body:   jsonData,
		Header: req.Header,
	}
	status := CheckServerRequest(req, c.ServiceProvider, content)
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
		slog.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(e.ServiceProvider.Method, e.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		slog.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}
	content := types.HTTPContent{
		Body:   jsonData,
		Header: req.Header,
	}

	status := CheckServerRequest(req, e.ServiceProvider, content)
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
		slog.Error("[Schedule] Failed to marshal request body", "error", err)
		return false
	}
	req, err := http.NewRequest(e.ServiceProvider.Method, e.ServiceProvider.URL, bytes.NewReader(jsonData))
	if err != nil {
		slog.Error("[Schedule] Failed to prepare request", "error", err)
		return false
	}
	content := types.HTTPContent{
		Body:   jsonData,
		Header: req.Header,
	}
	status := CheckServerRequest(req, e.ServiceProvider, content)
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
		slog.Error("[Schedule] Unknown service name", "error", sp.ServiceName)
		return nil
	}
	return server
}

func CheckServerRequest(req *http.Request, serviceProvider types.ServiceProvider, content types.HTTPContent) bool {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	if serviceProvider.ExtraHeaders != "{}" {
		var extraHeader map[string]interface{}
		err := json.Unmarshal([]byte(serviceProvider.ExtraHeaders), &extraHeader)
		if err != nil {
			slog.Error("Error parsing JSON:", err.Error())
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
			Content:      content,
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
		slog.Error("[Schedule] Failed to request", "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("[Schedule] Failed to request", "error", resp.StatusCode)
		return false
	}
	if serviceProvider.Flavor == types.FlavorSmartVision {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("[Schedule] Failed to read response body", "error", err)
			return false
		}
		var respData map[string]interface{}
		err = json.Unmarshal(body, &respData)
		if err != nil {
			return false
		}
		statusCode, ok := respData["status_code"].(float64)
		if !ok {
			return false
		}
		if statusCode != 200 {
			return false
		}
	}
	return true
}
