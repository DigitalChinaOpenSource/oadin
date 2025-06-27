package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"oadin/internal/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/provider"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils"
	"oadin/internal/utils/bcode"
	"oadin/version"
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
	sp := &types.ServiceProvider{}
	m := &types.Model{}

	sp.ProviderName = request.ProviderName
	sp.ServiceSource = request.ServiceSource
	sp.Method = http.MethodPost
	if request.Method != "" {
		sp.Method = request.Method
	}
	sp.ServiceName = request.ServiceName
	sp.Desc = request.Desc
	sp.Flavor = request.ApiFlavor

	m.ProviderName = request.ProviderName
	providerServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, request.ServiceName)
	if request.ServiceSource == types.ServiceSourceRemote {
		sp.URL = request.Url
		if request.Url == "" {
			sp.URL = providerServiceInfo.RequestUrl
		}
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

		m.ModelName = providerServiceInfo.DefaultModel
	} else {
		recommendConfig := getRecommendConfig(request.ServiceName)
		// Check if ollama is installed locally and if it is available.
		// If it is available, proceed to the next step. Otherwise, prompt that ollama is not installed.
		engineProvider := provider.GetModelEngine(recommendConfig.ModelEngine)
		engineConfig := engineProvider.GetConfig()
		if request.ModelName != "" {
			recommendConfig.ModelName = request.ModelName
		}

		if _, err := os.Stat(engineConfig.ExecPath); os.IsNotExist(err) {
			err := engineProvider.InstallEngine()
			if err != nil {
				slog.Error("Install model engine failed :", err.Error())
				return nil, bcode.ErrAIGCServiceInstallEngine
			}
		}

		slog.Info("Setting env...")
		err := engineProvider.InitEnv()
		if err != nil {
			slog.Error("Setting env error: ", err.Error())
			return nil, bcode.ErrAIGCServiceInitEnv
		}
		err = engineProvider.HealthCheck()
		if err != nil {
			err = engineProvider.StartEngine()
			if err != nil {
				slog.Error("Start engine error: ", err.Error())
				return nil, bcode.ErrAIGCServiceStartEngine
			}

			slog.Info("Waiting ollama engine start 60s...")
			for i := 60; i > 0; i-- {
				time.Sleep(time.Second * 1)
				err = engineProvider.HealthCheck()
				if err == nil {
					break
				}
				slog.Info("Waiting ollama engine start ...", strconv.Itoa(i), "s")
			}
		}

		err = engineProvider.HealthCheck()
		if err != nil {
			slog.Error("Ollama engine failed start...")
			return nil, bcode.ErrAIGCServiceStartEngine
		}

		sp.URL = providerServiceInfo.RequestUrl
		sp.AuthType = types.AuthTypeNone
		sp.AuthKey = ""
		sp.ExtraJSONBody = ""
		sp.ExtraHeaders = ""
		sp.Properties = `{"max_input_tokens":2048,"supported_response_mode":["stream","sync"],"mode_is_changeable":true,"xpu":["GPU"]}`
		sp.Status = 0

		// Check whether deepseek-r1 has been pulled locally.
		// If it has been pulled, proceed to the next step. Otherwise, call the ollama API to pull it.
		if !request.SkipModelFlag {
			models, err := engineProvider.ListModels(ctx)
			if err != nil {
				slog.Error("Get model list error: ", err.Error())
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
			m.ModelName = strings.ToLower(recommendConfig.ModelName)

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
				slog.Info("[services] Add model: ", m.ModelName)
			}
			if m.Status == "failed" {
				m.Status = "downloading"
			}

			if !isPulled {
				stream := false
				pullReq := &types.PullModelRequest{
					Model:  recommendConfig.ModelName,
					Stream: &stream,
				}
				go AsyncPullModel(ctx, sp, m, pullReq)
			} else {
				m.Status = "downloaded"
				err = s.Ds.Put(ctx, m)
				if err != nil {
					return nil, bcode.ErrAddModel
				}
			}
		}

		if request.ServiceName == types.ServiceChat {
			generateProviderServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ApiFlavor, types.ServiceGenerate)
			generateSp := &types.ServiceProvider{}
			generateSp.ProviderName = strings.Replace(request.ProviderName, "chat", "generate", -1)
			generateSp.ServiceSource = request.ServiceSource
			generateSp.AuthType = request.AuthType
			generateSp.Status = sp.Status
			generateSp.Method = http.MethodPost
			if request.Method != "" {
				generateSp.Method = request.Method
			}
			generateSp.ServiceName = strings.Replace(request.ServiceName, "chat", "generate", -1)
			generateSp.Desc = strings.Replace(request.Desc, "chat", "generate", -1)
			generateSp.Flavor = request.ApiFlavor
			generateSp.URL = generateProviderServiceInfo.RequestUrl

			generateSpIsExist, err := s.Ds.IsExist(ctx, generateSp)
			if err != nil {
				generateSpIsExist = false
			}

			if !generateSpIsExist {
				err := s.Ds.Add(ctx, generateSp)
				if err != nil {
					slog.Error("Service Provider model already exist")
					return nil, bcode.ErrModelIsExist
				}
			}

			err = s.defaultProviderProcess(ctx, types.ServiceGenerate, generateSp.ServiceSource, generateSp.ProviderName)
			if err != nil {
				return nil, err
			}

		}
	}

	// Check whether the service provider already exists.
	spIsExist, err := s.Ds.IsExist(ctx, sp)
	if err != nil {
		spIsExist = false
	}

	// Check whether the service provider model already exists.
	mIsExist, err := s.Ds.IsExist(ctx, m)
	if spIsExist && mIsExist {
		slog.Error("Service Provider model already exist")
		return nil, bcode.ErrModelIsExist
	}

	// todo: pending transaction processing
	if !spIsExist {
		err = s.Ds.Add(ctx, sp)
		if err != nil {
			slog.Error("Add service provider error: %s", err.Error())
			return nil, bcode.ErrAIGCServiceAddProvider
		}

		// Temporary logic, to be removed later.
		// Add model service
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
			err = s.Ds.Add(ctx, modelService)
			if err != nil {
				slog.Error("Add model service error: %s", err.Error())
				return nil, bcode.ErrAddModelService
			}
		}

		err = s.defaultProviderProcess(ctx, "models", "local", fmt.Sprintf("%s_%s_%s", "local", "ollama", "models"))
		if err != nil {
			return nil, err
		}
	}

	// Default provider processing
	err = s.defaultProviderProcess(ctx, sp.ServiceName, sp.ServiceSource, sp.ProviderName)
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
		Version:          version.OadinVersion,
		Services:         dbServices.Services,
		ServiceProviders: dbServices.ServiceProviders,
	}, nil
}

func (s *AIGCServiceImpl) ImportService(ctx context.Context, request *dto.ImportServiceRequest) (*dto.ImportServiceResponse, error) {
	if request.Version != version.OadinVersion {
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
				slog.Info(fmt.Sprintf("Pull model %s start ...", m))
				stream := false
				pullReq := &types.PullModelRequest{
					Model:  m,
					Stream: &stream,
				}
				modelObj := new(types.Model)
				modelObj.ProviderName = tmpSp.ProviderName
				modelObj.ModelName = strings.ToLower(m)

				err = s.Ds.Get(ctx, modelObj)
				if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
					// todo debug log output
					return nil, bcode.ErrServer
				} else if errors.Is(err, datastore.ErrEntityInvalid) {
					modelObj.Status = "downloading"
					err = s.Ds.Add(ctx, modelObj)
					if err != nil {
						return nil, bcode.ErrAddModel
					}
				}
				if modelObj.Status == "failed" {
					modelObj.Status = "downloading"
				}
				go AsyncPullModel(ctx, tmpSp, modelObj, pullReq)
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
					err := s.Ds.Add(ctx, tmpModel)
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
			err := s.Ds.Add(ctx, tmpSp)
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
				err := s.Ds.Add(ctx, generateSp)
				if err != nil {
					slog.Error("Service Provider model already exist")
					return nil, bcode.ErrModelIsExist
				}
			}

			err = s.defaultProviderProcess(ctx, types.ServiceGenerate, generateSp.ServiceSource, generateSp.ProviderName)
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
			providerEngine := provider.GetModelEngine(localSp.Flavor)
			err = providerEngine.HealthCheck()
			if err == nil {
				serviceStatus = 1
			}
		}
		tmp.RemoteProvider = dsService.RemoteProvider
		if dsService.RemoteProvider != "" {
			remoteSp := &types.ServiceProvider{
				ProviderName: dsService.RemoteProvider,
			}
			err = s.Ds.Get(ctx, remoteSp)
			remoteModel := &types.Model{
				ProviderName: dsService.RemoteProvider,
			}
			err = s.Ds.Get(ctx, remoteModel)

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

func (s *AIGCServiceImpl) defaultProviderProcess(ctx context.Context, serviceName, serviceSource, providerName string) error {
	service := &types.Service{
		Name: serviceName,
	}
	err := s.Ds.Get(ctx, service)
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

	err = s.Ds.Put(ctx, service)
	if err != nil {
		slog.Error("Service default ", serviceSource, " provider set failed")
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
	switch service {
	case types.ServiceChat:
		return types.RecommendConfig{
			ModelEngine:       "ollama",
			ModelName:         "deepseek-r1:7b",
			EngineDownloadUrl: "http://120.232.136.73:31619/byzedev/ipex-llm-ollama-Installer-20250226.exe",
		}
	case types.ServiceEmbed:
		return types.RecommendConfig{
			ModelEngine:       "ollama",
			ModelName:         "bge-m3",
			EngineDownloadUrl: "http://120.232.136.73:31619/byzedev/ipex-llm-ollama-Installer-20250226.exe",
		}
	case types.ServiceModels:
		return types.RecommendConfig{}
	case types.ServiceGenerate:
		return types.RecommendConfig{
			ModelEngine:       "ollama",
			ModelName:         "deepseek-r1:7b",
			EngineDownloadUrl: "http://120.232.136.73:31619/byzedev/ipex-llm-ollama-Installer-20250226.exe",
		}
	default:
		return types.RecommendConfig{}
	}
}
