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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/constants"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/schedule"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
	"intel.com/aog/version"
)

const (
	// Download URLs
	OllamaEngineDownloadURL   = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ipex-llm-ollama-win.zip"
	OpenvinoEngineDownloadURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/windows/ovms_windows.zip"
)

type AIGCService interface {
	CreateAIGCService(ctx context.Context, request *dto.CreateAIGCServiceRequest) (*dto.CreateAIGCServiceResponse, error)
	UpdateAIGCService(ctx context.Context, request *dto.UpdateAIGCServiceRequest) (*dto.UpdateAIGCServiceResponse, error)
	GetAIGCService(ctx context.Context, request *dto.GetAIGCServiceRequest) (*dto.GetAIGCServiceResponse, error)
	GetAIGCServices(ctx context.Context, request *dto.GetAIGCServicesRequest) (*dto.GetAIGCServicesResponse, error)
	ExportService(ctx context.Context, request *dto.ExportServiceRequest) (*dto.ExportServiceResponse, error)
	ImportService(ctx context.Context, request *dto.ImportServiceRequest) (*dto.ImportServiceResponse, error)
}

type AIGCServiceImpl struct {
	Ds datastore.Datastore
}

func NewAIGCService() AIGCService {
	return &AIGCServiceImpl{
		Ds: datastore.GetDefaultDatastore(),
	}
}

func (s *AIGCServiceImpl) CreateAIGCService(ctx context.Context, request *dto.CreateAIGCServiceRequest) (*dto.CreateAIGCServiceResponse, error) {
	if request.ServiceSource == "" {
		request.ServiceSource = types.ServiceSourceLocal
	}

	service := &types.Service{Name: request.ServiceName}
	err := s.Ds.Get(ctx, service)
	if err == nil {
		service.Status = 2
		_ = s.Ds.Put(ctx, service)
	}

	sp := &types.ServiceProvider{}
	m := &types.Model{}
	mIsExist := false
	providerServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, request.ServiceName)
	sp.ProviderName = request.ProviderName
	sp.ServiceSource = request.ServiceSource
	sp.Method = http.MethodPost
	if request.Method != "" {
		sp.Method = request.Method
	}
	sp.ServiceName = request.ServiceName
	sp.Desc = request.Desc
	sp.Flavor = request.ApiFlavor
	if request.Url == "" {
		sp.URL = providerServiceInfo.RequestUrl
	}

	if request.ApiFlavor == types.FlavorOllama && request.ServiceName == types.ServiceTextToImage {
		return nil, fmt.Errorf("Ollama not support  text-to-image service")
	}

	m.ProviderName = request.ProviderName
	if request.ServiceSource == types.ServiceSourceRemote {
		sp.URL = request.Url
		sp.AuthType = request.AuthType
		if request.AuthType != types.AuthTypeNone && request.AuthKey == "" {
			return nil, bcode.ErrProviderAuthInfoLost
		}
		sp.AuthKey = request.AuthKey
		sp.ExtraJSONBody = request.ExtraJsonBody
		sp.ExtraHeaders = request.ExtraHeaders
		if request.ExtraHeaders == "" {
			sp.ExtraHeaders = providerServiceInfo.ExtraHeaders
		}
		sp.Properties = request.Properties

		m.ServiceSource = types.ServiceSourceRemote
		m.ServiceName = request.ServiceName
		m.ModelName = providerServiceInfo.DefaultModel

		// 在创建后置为0
		err := s.Ds.Get(ctx, service)
		if err == nil {
			service.Status = 0
			_ = s.Ds.Put(ctx, service)
		}
	} else {
		recommendConfig := getRecommendConfig(request.ServiceName)
		// Check if ollama is installed locally and if it is available.
		// If it is available, proceed to the next step. Otherwise, prompt that ollama is not installed.
		engineProvider := provider.GetModelEngine(recommendConfig.ModelEngine)
		engineConfig := engineProvider.GetConfig()
		if request.ModelName != "" {
			recommendConfig.ModelName = request.ModelName
		}
		if request.ApiFlavor != recommendConfig.ModelEngine {
			request.ApiFlavor = recommendConfig.ModelEngine
			sp.Flavor = recommendConfig.ModelEngine
		}

		if request.ProviderName == "" {
			sp.ProviderName = fmt.Sprintf("%s_%s_%s", request.ServiceSource, request.ApiFlavor, request.ServiceName)
		}

		cmd := exec.Command(engineConfig.ExecFile, "-h")
		err := cmd.Run()
		if err != nil {
			logger.LogicLogger.Info("Check model engine " + recommendConfig.ModelEngine + "  not exist...")
			reCheckCmd := exec.Command(engineConfig.ExecPath+"/"+engineConfig.ExecFile, "-h")
			err = reCheckCmd.Run()
			_, isExistErr := os.Stat(engineConfig.ExecPath + "/" + engineConfig.ExecFile)
			if err != nil && isExistErr != nil {
				logger.LogicLogger.Info("Model engine " + recommendConfig.ModelEngine + " not exist, start download...")
				err := engineProvider.InstallEngine()
				if err != nil {
					logger.LogicLogger.Error("Install model "+recommendConfig.ModelEngine+" engine failed :", err.Error())
					return nil, bcode.ErrAIGCServiceInstallEngine
				}
				logger.LogicLogger.Info("Model engine " + recommendConfig.ModelEngine + " download completed...")
			}
		}

		logger.LogicLogger.Info("Setting env...")
		err = engineProvider.InitEnv()
		if err != nil {
			logger.LogicLogger.Error("Setting env error: ", err.Error())
			return nil, bcode.ErrAIGCServiceInitEnv
		}

		err = engineProvider.HealthCheck()
		if err != nil {
			err = engineProvider.StartEngine(types.EngineStartModeDaemon)
			if err != nil {
				logger.LogicLogger.Error("Start engine "+recommendConfig.ModelEngine+" error: ", err.Error())
				return nil, bcode.ErrAIGCServiceStartEngine
			}

			logger.LogicLogger.Info("Waiting " + recommendConfig.ModelEngine + " start 60s...")
			for i := 60; i > 0; i-- {
				time.Sleep(time.Second * 1)
				err = engineProvider.HealthCheck()
				if err == nil {
					break
				}
				logger.LogicLogger.Info("Waiting "+recommendConfig.ModelEngine+" start ...", strconv.Itoa(i), "s")
			}
		}

		err = engineProvider.HealthCheck()
		if err != nil {
			service := &types.Service{Name: request.ServiceName}
			err := s.Ds.Get(ctx, service)
			if err == nil {
				service.Status = -1
				_ = s.Ds.Put(ctx, service)
			}
			logger.LogicLogger.Error(recommendConfig.ModelEngine + "failed start...")
			return nil, bcode.ErrAIGCServiceStartEngine
		}

		// 在创建后置为0
		service := &types.Service{Name: request.ServiceName}
		err = s.Ds.Get(ctx, service)
		if err == nil {
			service.Status = 0
			_ = s.Ds.Put(ctx, service)
		}

		sp.URL = providerServiceInfo.RequestUrl
		sp.AuthType = types.AuthTypeNone
		sp.AuthKey = ""
		sp.ExtraJSONBody = ""
		sp.ExtraHeaders = ""
		// TODO This needs to be obtained dynamically later.
		sp.Properties = `{"max_input_tokens":2048,"supported_response_mode":["stream","sync"],"mode_is_changeable":true,"xpu":["GPU"]}`
		sp.Status = 0

		// Check whether deepseek-r1 has been pulled locally.
		// If it has been pulled, proceed to the next step. Otherwise, call the ollama API to pull it.
		if !request.SkipModelFlag {
			models, err := engineProvider.ListModels(ctx)
			if err != nil {
				logger.LogicLogger.Error("Get "+recommendConfig.ModelEngine+" model list error: ", err.Error())
				return nil, bcode.ErrGetEngineModelList
			}
			isPulled := false
			for _, model := range models.Models {
				if model.Name == recommendConfig.ModelName {
					isPulled = true
					break
				}
			}

			m.ProviderName = sp.ProviderName
			m.ModelName = recommendConfig.ModelName
			m.ServiceName = request.ServiceName
			m.ServiceSource = types.ServiceSourceLocal

			err = s.Ds.Get(ctx, m)
			if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
				logger.LogicLogger.Error("Get model from db error:", err)
				return nil, bcode.ErrServer
			} else if errors.Is(err, datastore.ErrEntityInvalid) {
				m.Status = "downloading"
				err = s.Ds.Add(ctx, m)
				if err != nil {
					logger.LogicLogger.Error("Add model to db error:", err)
					return nil, bcode.ErrAddModel
				}
			} else {
				mIsExist = true
			}

			if !isPulled {
				if m.Status == "failed" {
					m.Status = "downloading"
				}
				stream := false
				pullReq := &types.PullModelRequest{
					Model:     recommendConfig.ModelName,
					Stream:    &stream,
					ModelType: sp.ServiceName,
				}
				go AsyncPullModel(sp, m, pullReq)
			} else {
				m.Status = "downloaded"
				err = s.Ds.Put(ctx, m)
				if err != nil {
					return nil, bcode.ErrAddModel
				}
				// 本地已存在模型则将 Status 置为 1
				checkServer := ChooseCheckServer(*sp, m.ModelName)
				if checkServer != nil {
					service.Status = 1
					_ = s.Ds.Put(ctx, service)
				}
			}
		}

		// get current service  task_type
		currentServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, request.ServiceName)
		providerServices := schedule.GetProviderServices(request.ApiFlavor)

		for serviceName, serviceInfo := range providerServices {
			if serviceInfo.TaskType == currentServiceInfo.TaskType && serviceName != request.ServiceName {
				// create related service provider
				relatedSp := &types.ServiceProvider{}
				relatedSp.ProviderName = strings.Replace(request.ProviderName, request.ServiceName, serviceName, -1)
				relatedSp.ServiceSource = request.ServiceSource
				relatedSp.AuthType = request.AuthType
				relatedSp.Status = sp.Status
				relatedSp.Method = http.MethodPost
				relatedSp.ServiceName = serviceName
				relatedSp.Desc = strings.Replace(request.Desc, request.ServiceName, serviceName, -1)
				relatedSp.Flavor = request.ApiFlavor
				relatedSp.URL = serviceInfo.RequestUrl

				// check if related service provider already exists
				isExist, err := s.Ds.IsExist(ctx, relatedSp)
				if err != nil {
					isExist = false
				}

				if !isExist {
					err := s.Ds.Put(ctx, relatedSp)
					if err != nil {
						logger.LogicLogger.Error("Service Provider model already exist")
						return nil, bcode.ErrModelIsExist
					}
				}

				err = DefaultProviderProcess(ctx, serviceName, relatedSp.ServiceSource, relatedSp.ProviderName)
				if err != nil {
					logger.LogicLogger.Error("Failed to process related service",
						"service", serviceName,
						"error", err)
					return nil, err
				}
			}
		}
	}

	// Check whether the service provider already exists.
	spIsExist, err := s.Ds.IsExist(ctx, sp)
	if err != nil {
		spIsExist = false
	}

	// Check whether the service provider model already exists.
	if spIsExist && mIsExist {
		logger.LogicLogger.Error("Service Provider model already exist")
		return nil, bcode.ErrModelIsExist
	}

	// todo: pending transaction processing
	if !spIsExist {
		err = s.Ds.Put(ctx, sp)
		if err != nil {
			logger.LogicLogger.Error("Add service provider error: %s", err.Error())
			return nil, bcode.ErrAIGCServiceAddProvider
		}

		// Temporary logic, to be removed later.
		// Add model service
		if sp.Flavor == types.FlavorOllama {
			modelService := &types.ServiceProvider{
				ProviderName:  fmt.Sprintf("%s_%s_%s", "local", "ollama", "models"),
				ServiceName:   "models",
				ServiceSource: "local",
				Desc:          "",
				Method:        http.MethodGet,
				URL:           "http://127.0.0.1:16677/api/tags",
				AuthType:      "none",
				AuthKey:       "",
				Flavor:        "ollama",
				ExtraHeaders:  "{}",
				ExtraJSONBody: "{}",
				Properties:    "{}",
				Status:        1,
			}
			modelIsExist, err := s.Ds.IsExist(ctx, modelService)
			if !modelIsExist {
				err = s.Ds.Put(ctx, modelService)
				if err != nil {
					logger.LogicLogger.Error("Add model service error: %s", err.Error())
					return nil, bcode.ErrAddModelService
				}
			}

			err = DefaultProviderProcess(ctx, types.ServiceModels, types.ServiceSourceLocal,
				fmt.Sprintf("%s_%s_%s", types.ServiceSourceLocal, sp.Flavor, types.ServiceModels))
			if err != nil {
				return nil, err
			}
		}

	}

	// Default provider processing
	err = DefaultProviderProcess(ctx, sp.ServiceName, sp.ServiceSource, sp.ProviderName)
	if err != nil {
		return nil, err
	}

	return &dto.CreateAIGCServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) UpdateAIGCService(ctx context.Context, request *dto.UpdateAIGCServiceRequest) (*dto.UpdateAIGCServiceResponse, error) {
	service := types.Service{
		Name: request.ServiceName,
	}

	err := s.Ds.Get(ctx, &service)
	if err != nil {
		return nil, bcode.ErrServiceRecordNotFound
	}

	if request.RemoteProvider != "" {
		service.RemoteProvider = request.RemoteProvider
	}
	if request.LocalProvider != "" {
		service.LocalProvider = request.LocalProvider
	}
	service.HybridPolicy = request.HybridPolicy
	err = s.Ds.Put(ctx, &service)
	if err != nil {
		return nil, bcode.ErrServiceRecordNotFound
	}

	return &dto.UpdateAIGCServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) GetAIGCService(ctx context.Context, request *dto.GetAIGCServiceRequest) (*dto.GetAIGCServiceResponse, error) {
	return &dto.GetAIGCServiceResponse{}, nil
}

func (s *AIGCServiceImpl) ExportService(ctx context.Context, request *dto.ExportServiceRequest) (*dto.ExportServiceResponse, error) {
	dbService := &types.Service{}
	if request.ServiceName != "" {
		dbService.Name = request.ServiceName
	}
	dbProvider := &types.ServiceProvider{}
	if request.ProviderName != "" {
		dbProvider.ProviderName = request.ProviderName
	}
	model := &types.Model{}
	if request.ModelName != "" {
		model.ModelName = request.ModelName
	}
	dbServices, err := getAllServices(dbService, dbProvider, model)
	if err != nil {
		return nil, err
	}

	return &dto.ExportServiceResponse{
		Version:          version.AOGVersion,
		Services:         dbServices.Services,
		ServiceProviders: dbServices.ServiceProviders,
	}, nil
}

func (s *AIGCServiceImpl) ImportService(ctx context.Context, request *dto.ImportServiceRequest) (*dto.ImportServiceResponse, error) {
	if request.Version != version.AOGVersion {
		return nil, bcode.ErrAIGCServiceVersionNotMatch
	}

	dbService := &types.Service{}
	dbProvider := &types.ServiceProvider{}
	model := &types.Model{}

	dbServices, err := getAllServices(dbService, dbProvider, model)
	if err != nil {
		return nil, err
	}

	for serviceName, service := range request.Services {
		if !utils.Contains(types.SupportService, serviceName) {
			return nil, bcode.ErrUnSupportAIGCService
		}

		if !utils.Contains(types.SupportHybridPolicy, service.HybridPolicy) {
			return nil, bcode.ErrUnSupportHybridPolicy
		}

		if service.ServiceProviders.Local == "" && service.ServiceProviders.Remote == "" {
			return nil, bcode.ErrAIGCServiceBadRequest
		}

		if service.HybridPolicy != types.HybridPolicyDefault && service.HybridPolicy != "" {
			tmpService := dbServices.Services[serviceName]
			tmpService.HybridPolicy = service.HybridPolicy
			dbServices.Services[serviceName] = tmpService
		}
	}

	for providerName, p := range request.ServiceProviders {
		if !utils.Contains(types.SupportFlavor, p.APIFlavor) {
			return nil, bcode.ErrUnSupportFlavor
		}
		if !utils.Contains(types.SupportAuthType, p.AuthType) {
			return nil, bcode.ErrUnSupportAuthType
		}

		if !utils.Contains(types.SupportService, p.ServiceName) {
			return nil, bcode.ErrUnSupportAIGCService
		}

		//if len(p.Models) == 0 && p.ServiceName != types.ServiceModels {
		//	return nil, bcode.ErrProviderModelEmpty
		//}
		providerDefaultInfo := schedule.GetProviderServiceDefaultInfo(p.APIFlavor, p.ServiceName)
		tmpSp := &types.ServiceProvider{}
		tmpSp.ProviderName = providerName
		tmpSp.AuthKey = p.AuthKey
		tmpSp.AuthType = p.AuthType
		tmpSp.Desc = p.Desc
		tmpSp.Flavor = p.APIFlavor
		tmpSp.Method = p.Method
		tmpSp.ServiceName = p.ServiceName
		tmpSp.ServiceSource = p.ServiceSource
		tmpSp.URL = p.URL
		if p.URL == "" {
			tmpSp.URL = providerDefaultInfo.RequestUrl
		}
		tmpSp.Status = 0
		tmpSp.ExtraHeaders = providerDefaultInfo.ExtraHeaders
		tmpSp.ExtraJSONBody = "{}"
		tmpSp.Properties = "{}"
		if p.ServiceName == types.ServiceChat || p.ServiceName == types.ServiceGenerate {
			tmpSp.Properties = `{"max_input_tokens":2048,"supported_response_mode":["stream","sync"],"mode_is_changeable":true,"xpu":["GPU"]}`
		}

		// engineProvider := provider.GetModelEngine(tmpSp.Flavor)
		for _, m := range p.Models {
			if p.ServiceSource == types.ServiceSourceLocal && !utils.Contains(dbServices.ServiceProviders[providerName].Models, m) {
				logger.LogicLogger.Info(fmt.Sprintf("Pull model %s start ...", m))
				stream := false
				pullReq := &types.PullModelRequest{
					Model:     m,
					Stream:    &stream,
					ModelType: p.ServiceName,
				}
				modelObj := new(types.Model)
				modelObj.ProviderName = tmpSp.ProviderName
				modelObj.ModelName = m

				err = s.Ds.Get(ctx, modelObj)
				if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
					// todo debug log output
					return nil, bcode.ErrServer
				} else if errors.Is(err, datastore.ErrEntityInvalid) {
					modelObj.Status = "downloading"
					err = s.Ds.Put(ctx, modelObj)
					if err != nil {
						return nil, bcode.ErrAddModel
					}
				}
				if modelObj.Status == "failed" {
					modelObj.Status = "downloading"
				}
				go AsyncPullModel(tmpSp, modelObj, pullReq)
				//_, err := engineProvider.PullModel(ctx, pullReq, nil)
				//if err != nil {
				//	slog.Error(fmt.Sprintf("Pull model error: %s", err.Error()))
				//	return nil, bcode.ErrEnginePullModel
				//}
				//
				//slog.Info(fmt.Sprintf("Pull model %s completed ...", m))
			} else if p.ServiceSource == types.ServiceSourceRemote && !utils.Contains(dbServices.ServiceProviders[providerName].Models, m) {
				server := ChooseCheckServer(*tmpSp, m)
				checkRes := server.CheckServer()
				if !checkRes {
					return nil, bcode.ErrProviderIsUnavailable
				}
				tmpSp.Status = 1
				tmpModel := &types.Model{}
				tmpModel.ModelName = m
				tmpModel.ProviderName = tmpSp.ProviderName
				tmpModel.Status = "downloaded"
				tmpModel.UpdatedAt = time.Now()

				isExist, err := s.Ds.IsExist(ctx, tmpModel)
				if err != nil || !isExist {
					err := s.Ds.Put(ctx, tmpModel)
					if err != nil {
						return nil, bcode.ErrAddModel
					}
				}
			}
		}

		if _, ok := dbServices.ServiceProviders[providerName]; ok {
			checkSp := &types.ServiceProvider{
				ProviderName: providerName,
			}
			err = s.Ds.Get(ctx, checkSp)
			if err != nil {
				return nil, bcode.ErrProviderUpdateFailed
			}
			tmpSp.ID = checkSp.ID
			tmpSp.Status = checkSp.Status
			if tmpSp.AuthType == "none" {
				tmpSp.AuthType = checkSp.AuthType
				tmpSp.AuthKey = checkSp.AuthKey
			}

			err := s.Ds.Put(ctx, tmpSp)
			if err != nil {
				return nil, bcode.ErrProviderUpdateFailed
			}

		} else {
			err := s.Ds.Put(ctx, tmpSp)
			if err != nil {
				return nil, err
			}
		}

		if p.ServiceSource == types.ServiceSourceLocal && p.ServiceName == types.ServiceChat {
			generateProviderServiceInfo := schedule.GetProviderServiceDefaultInfo(tmpSp.Flavor, types.ServiceGenerate)
			generateSp := &types.ServiceProvider{}
			generateSp.ProviderName = strings.Replace(tmpSp.ProviderName, "chat", "generate", -1)
			generateSp.ServiceSource = tmpSp.ServiceSource
			generateSp.AuthType = tmpSp.AuthType
			generateSp.Status = tmpSp.Status
			generateSp.Method = http.MethodPost
			generateSp.ServiceName = strings.Replace(tmpSp.ServiceName, "chat", "generate", -1)
			generateSp.Desc = strings.Replace(tmpSp.Desc, "chat", "generate", -1)
			generateSp.Flavor = tmpSp.Flavor
			generateSp.URL = generateProviderServiceInfo.RequestUrl
			generateSp.Properties = tmpSp.Properties
			generateSp.ExtraJSONBody = tmpSp.ExtraJSONBody
			generateSp.ExtraHeaders = tmpSp.ExtraHeaders

			generateSpIsExist, err := s.Ds.IsExist(ctx, generateSp)
			if err != nil {
				generateSpIsExist = false
			}

			if !generateSpIsExist {
				err := s.Ds.Put(ctx, generateSp)
				if err != nil {
					logger.LogicLogger.Error("Service Provider model already exist")
					return nil, bcode.ErrModelIsExist
				}
			}

			err = DefaultProviderProcess(ctx, types.ServiceGenerate, generateSp.ServiceSource, generateSp.ProviderName)
			if err != nil {
				return nil, err
			}
		}

		// Check whether LocalProvider and RemoteProvider exist in DBServices. If they do not exist, add them.
		dbService.Name = p.ServiceName
		err := s.Ds.Get(ctx, dbService)
		if err != nil {
			return nil, err
		}
		if p.ServiceSource == types.ServiceSourceLocal && dbServices.Services[p.ServiceName].ServiceProviders.Local == "" {
			dbService.LocalProvider = tmpSp.ProviderName
		}
		if p.ServiceSource == types.ServiceSourceRemote && dbServices.Services[p.ServiceName].ServiceProviders.Remote == "" {
			dbService.RemoteProvider = tmpSp.ProviderName
		}
		dbService.HybridPolicy = dbServices.Services[p.ServiceName].HybridPolicy

		err = s.Ds.Put(ctx, dbService)
		if err != nil {
			return nil, bcode.ErrServiceUpdateFailed
		}
	}

	return &dto.ImportServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) GetAIGCServices(ctx context.Context, request *dto.GetAIGCServicesRequest) (*dto.GetAIGCServicesResponse, error) {
	service := &types.Service{}
	if request.ServiceName != "" {
		service.Name = request.ServiceName
	}

	list, err := s.Ds.List(ctx, service, &datastore.ListOptions{PageSize: 10, Page: 0})
	if err != nil {
		return nil, err
	}

	respData := make([]dto.Service, 0)

	for _, v := range list {
		tmp := dto.Service{}

		dsService := v.(*types.Service)
		tmp.ServiceName = dsService.Name
		tmp.LocalProvider = dsService.LocalProvider
		serviceStatus := 0
		if dsService.LocalProvider != "" {
			localSp := &types.ServiceProvider{
				ProviderName: dsService.LocalProvider,
			}
			err = s.Ds.Get(ctx, localSp)
			if err == nil {
				providerEngine := provider.GetModelEngine(localSp.Flavor)
				err = providerEngine.HealthCheck()
				if err == nil {
					serviceStatus = 1
				}
			}

		}
		tmp.RemoteProvider = dsService.RemoteProvider
		if dsService.RemoteProvider != "" {
			remoteSp := &types.ServiceProvider{
				ProviderName: dsService.RemoteProvider,
			}
			err = s.Ds.Get(ctx, remoteSp)
			if err != nil {
				serviceStatus = 0
				continue
			}
			remoteModel := &types.Model{
				ProviderName: dsService.RemoteProvider,
			}
			err = s.Ds.Get(ctx, remoteModel)
			if err != nil {
				serviceStatus = 0
				continue
			}
			if dsService.Name == types.ServiceTextToImage {
				continue
			}
			checkServerObj := ChooseCheckServer(*remoteSp, remoteModel.ModelName)
			status := checkServerObj.CheckServer()
			if status {
				serviceStatus = 1
			}
		}
		tmp.HybridPolicy = dsService.HybridPolicy
		// tmp.Status = dsService.Status
		tmp.Status = serviceStatus
		tmp.UpdatedAt = dsService.UpdatedAt
		tmp.CreatedAt = dsService.CreatedAt

		respData = append(respData, tmp)
	}

	return &dto.GetAIGCServicesResponse{
		Bcode: *bcode.AIGCServiceCode,
		Data:  respData,
	}, nil
}

func DefaultProviderProcess(ctx context.Context, serviceName, serviceSource, providerName string) error {
	ds := datastore.GetDefaultDatastore()
	service := &types.Service{
		Name: serviceName,
	}
	err := ds.Get(ctx, service)
	if err != nil {
		return err
	}

	if serviceSource == types.ServiceSourceLocal && service.LocalProvider != "" {
		return nil
	} else if serviceSource == types.ServiceSourceRemote && service.RemoteProvider != "" {
		return nil
	}

	if serviceSource == types.ServiceSourceLocal {
		service.LocalProvider = providerName
	} else if serviceSource == types.ServiceSourceRemote {
		service.RemoteProvider = providerName
	}

	err = ds.Put(ctx, service)
	if err != nil {
		logger.LogicLogger.Error("Service default ", serviceSource, " provider set failed")
		return err
	}

	return nil
}

func getAllServices(service *types.Service, provider *types.ServiceProvider, model *types.Model) (*dto.ImportServiceRequest, error) {
	ds := datastore.GetDefaultDatastore()

	serviceList, err := ds.List(context.Background(), service, &datastore.ListOptions{Page: 0, PageSize: 10})
	if err != nil {
		return nil, bcode.ErrAIGCServiceBadRequest
	}

	providerList, err := ds.List(context.Background(), provider, &datastore.ListOptions{Page: 0, PageSize: 100})
	if err != nil {
		return nil, bcode.ErrProviderInvalid
	}

	modelList, err := ds.List(context.Background(), model, &datastore.ListOptions{Page: 0, PageSize: 100})
	if err != nil {
		return nil, bcode.ErrModelRecordNotFound
	}

	dbServices := new(dto.ImportServiceRequest)
	dbServices.Services = make(map[string]dto.ServiceEntry)
	dbServices.ServiceProviders = make(map[string]dto.ServiceProviderEntry)
	for _, v := range serviceList {
		tmp := v.(*types.Service)
		tmpService := dbServices.Services[tmp.Name]
		tmpService.HybridPolicy = tmp.HybridPolicy
		tmpService.ServiceProviders.Local = tmp.LocalProvider
		tmpService.ServiceProviders.Remote = tmp.RemoteProvider
		dbServices.Services[tmp.Name] = tmpService
	}

	for _, v := range providerList {
		tmp := v.(*types.ServiceProvider)
		tmpProvider := dbServices.ServiceProviders[tmp.ProviderName]
		tmpProvider.AuthKey = tmp.AuthKey
		tmpProvider.AuthType = tmp.AuthType
		tmpProvider.Desc = tmp.Desc
		tmpProvider.APIFlavor = tmp.Flavor
		tmpProvider.Method = tmp.Method
		tmpProvider.ServiceName = tmp.ServiceName
		tmpProvider.ServiceSource = tmp.ServiceSource
		tmpProvider.URL = tmp.URL
		tmpProvider.Models = []string{}
		for _, m := range modelList {
			if m.(*types.Model).ProviderName == tmp.ProviderName {
				tmpProvider.Models = append(tmpProvider.Models, m.(*types.Model).ModelName)
			}
		}
		dbServices.ServiceProviders[tmp.ProviderName] = tmpProvider
	}

	return dbServices, nil
}

func getRecommendConfig(service string) types.RecommendConfig {
	recommendModelMap, _ := RecommendModels()
	recommendModelList := recommendModelMap[service]
	switch service {
	case types.ServiceChat:
		modelName := constants.DefaultChatModelName
		if len(recommendModelList) > 0 {
			modelName = recommendModelList[0].Name
		}
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         modelName,
			EngineDownloadUrl: OllamaEngineDownloadURL,
		}
	case types.ServiceEmbed:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         constants.DefaultEmbedModelName,
			EngineDownloadUrl: OllamaEngineDownloadURL,
		}
	case types.ServiceModels:
		return types.RecommendConfig{}
	case types.ServiceGenerate:
		modelName := constants.DefaultChatModelName
		if len(recommendModelList) > 0 {
			modelName = recommendModelList[0].Name
		}
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         modelName,
			EngineDownloadUrl: OllamaEngineDownloadURL,
		}
	case types.ServiceTextToImage:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultTextToImageModel,
			EngineDownloadUrl: OpenvinoEngineDownloadURL,
		}
	case types.ServiceSpeechToText:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultSpeechToTextModel,
			EngineDownloadUrl: OpenvinoEngineDownloadURL,
		}
	case types.ServiceSpeechToTextWS:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultSpeechToTextModel,
			EngineDownloadUrl: OpenvinoEngineDownloadURL,
		}
	default:
		return types.RecommendConfig{}
	}
}
