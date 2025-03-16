package server

import (
	"context"
	"fmt"
	"intel.com/aog/internal/schedule"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
	"intel.com/aog/version"
)

type AIGCService interface {
	CreateAIGCService(ctx context.Context, request *dto.CreateAIGCServiceRequest) (*dto.CreateAIGCServiceResponse, error)
	UpdateAIGCService(ctx context.Context, request *dto.UpdateAIGCServiceRequest) (*dto.UpdateAIGCServiceResponse, error)
	GetAIGCService(ctx context.Context, request *dto.GetAIGCServiceRequest) (*dto.GetAIGCServiceResponse, error)
	GetAIGCServices(ctx context.Context, request *dto.GetAIGCServicesRequest) (*dto.GetAIGCServicesResponse, error)
	ExportService(ctx context.Context, request *dto.ExportServiceRequest) (*dto.ExportServiceResponse, error)
	ImportService(ctx context.Context, request *dto.ImportServiceRequest) (*dto.ImportServiceResponse, error)
}

type AIGCServiceImpl struct{}

func NewAIGCService() AIGCService {
	return &AIGCServiceImpl{}
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

	if request.ServiceSource == types.ServiceSourceRemote {
		sp.URL = request.Url
		sp.AuthType = request.AuthType
		sp.AuthKey = request.AuthKey
		sp.ExtraJSONBody = request.ExtraJsonBody
		sp.ExtraHeaders = request.ExtraHeaders
		sp.Properties = request.Properties

		providerServiceInfo := schedule.GetProviderServiceDefaultInfo(request.ProviderName, request.ServiceName)
		m.ModelName = providerServiceInfo.DefaultModel
	} else {
		recommendConfig := getRecommendConfig(request.ServiceName)
		// 2. 检查本地是否安装了 ollama 以及 ollama 是否可用, 可用则进行下一步，否则 提示ollama 未安装
		engineProvider := provider.GetModelEngine(recommendConfig.ModelEngine)
		engineConfig := engineProvider.GetConfig()
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

		// 3. 检查本地是否已经拉取过deepseek-r1，已拉取进行下一步，否则调用ollama api 拉取
		models, err := engineProvider.ListModels(ctx)
		if err != nil {
			slog.Error("Get model list error: ", err.Error())
			return nil, bcode.ErrGetEngineModelList
		}
		isPulled := false
		for _, model := range models.Models {
			if model.Name == engineConfig.RecommendModel {
				isPulled = true
				break
			}
		}

		if !isPulled {
			stream := false
			pullReq := &types.PullModelRequest{
				Model:  engineConfig.RecommendModel,
				Stream: &stream,
			}

			slog.Info("Pull model  start ..." + engineConfig.RecommendModel)
			_, err := engineProvider.PullModel(ctx, pullReq, nil)
			if err != nil {
				slog.Error("Pull model error: ", err.Error())
				return nil, bcode.ErrEnginePullModel
			}
			slog.Info("Pull model %s completed ..." + engineConfig.RecommendModel)
		}

		sp.URL = fmt.Sprintf("%s://%s/%s", engineConfig.Scheme, engineConfig.Host, "api/chat")
		sp.AuthType = types.AuthTypeNone
		sp.AuthKey = ""
		sp.ExtraJSONBody = ""
		sp.ExtraHeaders = ""
		sp.Properties = `{"max_input_tokens":2048,"supported_response_mode":["stream","sync"],"mode_is_changeable":true,"xpu":["GPU"]}`
		sp.Status = 1

		m.ModelName = recommendConfig.ModelName
		m.Status = "downloaded"
	}

	ds := datastore.GetDefaultDatastore()
	// 检查service provider是否已存在
	spIsExist, err := ds.IsExist(ctx, sp)
	if err != nil {
		spIsExist = false
	}

	// 检查 service provider model 是否已存在
	mIsExist, err := ds.IsExist(ctx, m)
	if spIsExist && mIsExist {
		slog.Error("Service Provider model already exist")
		return nil, bcode.ErrModelIsExist
	}

	// todo 待进行事务处理
	if !spIsExist {
		err = ds.Add(ctx, sp)
		if err != nil {
			slog.Error("Add service provider error: %s", err.Error())
			return nil, bcode.ErrAIGCServiceAddProvider
		}

		// 临时逻辑，后期去除
		// 添加模型服务
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
		err = ds.Add(ctx, modelService)
		if err != nil {
			slog.Error("Add model service error: %s", err.Error())
			return nil, bcode.ErrAddModelService
		}
		err = defaultProviderProcess(ctx, "models", "local", fmt.Sprintf("%s_%s_%s", "local", "ollama", "models"))
		if err != nil {
			return nil, err
		}
	}

	err = ds.Add(ctx, m)
	if err != nil {
		slog.Error("Add model error: %s", err.Error())
		return nil, bcode.ErrAddModel
	}

	// Default provider processing
	err = defaultProviderProcess(ctx, sp.ServiceName, sp.ServiceSource, sp.ProviderName)
	if err != nil {
		return nil, err
	}

	return &dto.CreateAIGCServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) UpdateAIGCService(ctx context.Context, request *dto.UpdateAIGCServiceRequest) (*dto.UpdateAIGCServiceResponse, error) {
	ds := datastore.GetDefaultDatastore()
	service := types.Service{
		Name: request.ServiceName,
	}

	err := ds.Get(ctx, &service)
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
	err = ds.Put(ctx, &service)
	if err != nil {
		return nil, bcode.ErrServiceRecordNotFound
	}

	return &dto.UpdateAIGCServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) GetAIGCService(ctx context.Context, request *dto.GetAIGCServiceRequest) (*dto.GetAIGCServiceResponse, error) {
	// todo 实际实现逻辑
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
	ds := datastore.GetDefaultDatastore()

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

		if len(p.Models) == 0 && p.ServiceName != types.ServiceModels {
			return nil, bcode.ErrProviderModelEmpty
		}

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

		engineProvider := provider.GetModelEngine(tmpSp.Flavor)
		for _, m := range p.Models {
			if p.ServiceSource == types.ServiceSourceLocal && !utils.Contains(dbServices.ServiceProviders[providerName].Models, m) {
				slog.Info(fmt.Sprintf("Pull model %s start ...", m))
				stream := false
				pullReq := &types.PullModelRequest{
					Model:  m,
					Stream: &stream,
				}

				_, err := engineProvider.PullModel(ctx, pullReq, nil)
				if err != nil {
					slog.Error(fmt.Sprintf("Pull model error: %s", err.Error()))
					return nil, bcode.ErrEnginePullModel
				}

				slog.Info(fmt.Sprintf("Pull model %s completed ...", m))
			}

			server := ChooseCheckServer(*tmpSp, m)
			checkRes := server.CheckServer()
			if !checkRes {
				return nil, bcode.ErrProviderIsUnavailable
			}

			tmpModel := &types.Model{}
			tmpModel.ModelName = m
			tmpModel.ProviderName = tmpSp.ProviderName
			tmpModel.Status = "downloaded"
			tmpModel.UpdatedAt = time.Now()

			isExist, err := ds.IsExist(ctx, tmpModel)
			if err != nil || !isExist {
				err := ds.Add(ctx, tmpModel)
				if err != nil {
					return nil, bcode.ErrAddModel
				}
			}

		}

		if _, ok := dbServices.ServiceProviders[providerName]; ok {
			err := ds.Put(ctx, tmpSp)
			if err != nil {
				return nil, bcode.ErrProviderUpdateFailed
			}
		} else {
			err := ds.Add(ctx, tmpSp)
			if err != nil {
				return nil, err
			}
		}

		// 检查DBServices中LocalProvider和RemoteProvider是否存在，不存在则添加
		dbService.Name = p.ServiceName
		if p.ServiceSource == types.ServiceSourceLocal && dbServices.Services[p.ServiceName].ServiceProviders.Local == "" {
			dbService.LocalProvider = tmpSp.ProviderName
		}
		if p.ServiceSource == types.ServiceSourceRemote && dbServices.Services[p.ServiceName].ServiceProviders.Remote == "" {
			dbService.RemoteProvider = tmpSp.ProviderName
		}
		dbService.HybridPolicy = dbServices.Services[p.ServiceName].HybridPolicy

		err := ds.Put(ctx, dbService)
		if err != nil {
			return nil, bcode.ErrServiceUpdateFailed
		}
	}

	return &dto.ImportServiceResponse{
		Bcode: *bcode.AIGCServiceCode,
	}, nil
}

func (s *AIGCServiceImpl) GetAIGCServices(ctx context.Context, request *dto.GetAIGCServicesRequest) (*dto.GetAIGCServicesResponse, error) {
	ds := datastore.GetDefaultDatastore()
	service := &types.Service{}
	if request.ServiceName != "" {
		service.Name = request.ServiceName
	}

	list, err := ds.List(ctx, service, &datastore.ListOptions{PageSize: 10, Page: 0})
	if err != nil {
		return nil, err
	}

	respData := make([]dto.Service, 0)

	for _, v := range list {
		tmp := dto.Service{}

		dsService := v.(*types.Service)
		tmp.ServiceName = dsService.Name
		tmp.LocalProvider = dsService.LocalProvider
		tmp.RemoteProvider = dsService.RemoteProvider
		tmp.HybridPolicy = dsService.HybridPolicy
		tmp.Status = dsService.Status
		tmp.UpdatedAt = dsService.UpdatedAt
		tmp.CreatedAt = dsService.CreatedAt

		respData = append(respData, tmp)
	}

	return &dto.GetAIGCServicesResponse{
		Bcode: *bcode.AIGCServiceCode,
		Data:  respData,
	}, nil
}

func defaultProviderProcess(ctx context.Context, serviceName, serviceSource, providerName string) error {
	ds := datastore.GetDefaultDatastore()

	s := &types.Service{
		Name: serviceName,
	}
	err := ds.Get(ctx, s)
	if err != nil {
		return err
	}

	if serviceSource == types.ServiceSourceLocal && s.LocalProvider != "" {
		return nil
	} else if serviceSource == types.ServiceSourceRemote && s.RemoteProvider != "" {
		return nil
	}

	if serviceSource == types.ServiceSourceLocal {
		s.LocalProvider = providerName
	} else if serviceSource == types.ServiceSourceRemote {
		s.RemoteProvider = providerName
	}

	err = ds.Put(ctx, s)
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
			EngineDownloadUrl: "http://120.232.136.73:31619/aogdev/ipex-llm-ollama-Installer-20250122.exe",
		}
	case types.ServiceEmbed:
		return types.RecommendConfig{
			ModelEngine:       "ollama",
			ModelName:         "bge-m3",
			EngineDownloadUrl: "http://120.232.136.73:31619/aogdev/ipex-llm-ollama-Installer-20250122.exe",
		}
	case types.ServiceModels:
		return types.RecommendConfig{}
	case types.ServiceGenerate:
		return types.RecommendConfig{
			ModelEngine:       "ollama",
			ModelName:         "deepseek-r1:7b",
			EngineDownloadUrl: "http://120.232.136.73:31619/aogdev/ipex-llm-ollama-Installer-20250122.exe",
		}
	default:
		return types.RecommendConfig{}
	}
}
