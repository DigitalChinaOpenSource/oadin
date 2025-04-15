package server

import (
	"aipc/byze/internal/api/dto"
	"aipc/byze/internal/datastore"
	"aipc/byze/internal/provider"
	"aipc/byze/internal/provider/template"
	"aipc/byze/internal/schedule"
	"aipc/byze/internal/types"
	"aipc/byze/internal/utils"
	"aipc/byze/internal/utils/bcode"
	"aipc/byze/internal/utils/client"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

type Model interface {
	CreateModel(ctx context.Context, request *dto.CreateModelRequest) (*dto.CreateModelResponse, error)
	DeleteModel(ctx context.Context, request *dto.DeleteModelRequest) (*dto.DeleteModelResponse, error)
	GetModels(ctx context.Context, request *dto.GetModelsRequest) (*dto.GetModelsResponse, error)
}

type ModelImpl struct {
	Ds datastore.Datastore
}

func NewModel() Model {
	return &ModelImpl{
		Ds: datastore.GetDefaultDatastore(),
	}
}

func (s *ModelImpl) CreateModel(ctx context.Context, request *dto.CreateModelRequest) (*dto.CreateModelResponse, error) {
	sp := new(types.ServiceProvider)
	if request.ProviderName != "" {
		sp.ProviderName = request.ProviderName
	} else {
		// get default service provider
		// todo Currently only chat and generate services support pulling models.
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate && request.ServiceName != types.ServiceEmbed {
			return nil, bcode.ErrServer
		}

		service := &types.Service{}
		service.Name = request.ServiceName

		err := s.Ds.Get(ctx, service)
		if err != nil {
			return nil, err
		}

		if request.ServiceSource == types.ServiceSourceLocal && service.LocalProvider != "" {
			sp.ProviderName = service.LocalProvider
		} else if request.ServiceSource == types.ServiceSourceRemote && service.RemoteProvider != "" {
			sp.ProviderName = service.RemoteProvider
		}
	}

	sp.ServiceName = request.ServiceName
	sp.ServiceSource = request.ServiceSource

	err := s.Ds.Get(ctx, sp)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrServiceRecordNotFound
	}

	m := new(types.Model)
	m.ProviderName = sp.ProviderName
	m.ModelName = request.ModelName

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
	stream := false
	pullReq := &types.PullModelRequest{
		Model:  request.ModelName,
		Stream: &stream,
	}
	go AsyncPullModel(ctx, sp, m, pullReq)

	return &dto.CreateModelResponse{
		Bcode: *bcode.ModelCode,
	}, nil
}

func (s *ModelImpl) DeleteModel(ctx context.Context, request *dto.DeleteModelRequest) (*dto.DeleteModelResponse, error) {
	sp := new(types.ServiceProvider)
	sp.ProviderName = request.ProviderName

	err := s.Ds.Get(ctx, sp)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo err debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrServiceRecordNotFound
	}

	m := new(types.Model)
	m.ProviderName = request.ProviderName
	m.ModelName = request.ModelName

	err = s.Ds.Get(ctx, m)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo err debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrModelRecordNotFound
	}

	// Call engin to delete model.
	modelEngine := provider.GetModelEngine(sp.Flavor)
	deleteReq := &types.DeleteRequest{
		Model: request.ModelName,
	}

	err = modelEngine.DeleteModel(ctx, deleteReq)
	if err != nil {
		// todo err debug log output
		return nil, bcode.ErrEngineDeleteModel
	}

	err = s.Ds.Delete(ctx, m)
	if err != nil {
		// todo err debug log output
		return nil, err
	}
	return &dto.DeleteModelResponse{
		Bcode: *bcode.ModelCode,
	}, nil
}

func (s *ModelImpl) GetModels(ctx context.Context, request *dto.GetModelsRequest) (*dto.GetModelsResponse, error) {
	m := &types.Model{}
	if request.ModelName != "" {
		m.ModelName = request.ModelName
	}
	if request.ProviderName != "" {
		m.ProviderName = request.ProviderName
	}
	list, err := s.Ds.List(ctx, m, &datastore.ListOptions{
		Page:     0,
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}

	respData := make([]dto.Model, 0)
	for _, v := range list {
		tmp := new(dto.Model)
		dsModel := v.(*types.Model)

		tmp.ModelName = dsModel.ModelName
		tmp.ProviderName = dsModel.ProviderName
		tmp.Status = dsModel.Status
		tmp.CreatedAt = dsModel.CreatedAt
		tmp.UpdatedAt = dsModel.UpdatedAt

		respData = append(respData, *tmp)
	}

	return &dto.GetModelsResponse{
		Bcode: *bcode.ModelCode,
		Data:  respData,
	}, nil
}

func CreateModelStream(ctx context.Context, request dto.CreateModelRequest) (chan []byte, chan error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)
	defer close(dataChan)
	defer close(errChan)
	ds := datastore.GetDefaultDatastore()
	sp := new(types.ServiceProvider)
	if request.ProviderName != "" {
		sp.ProviderName = request.ProviderName
	} else {

		defer close(dataChan)
		defer close(errChan)
		// get default service provider
		// todo Currently only chat and generate services support pulling models.
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate && request.ServiceName != types.ServiceEmbed {
			errChan <- bcode.ErrServer
			return dataChan, errChan
		}

		service := &types.Service{}
		service.Name = request.ServiceName

		err := ds.Get(ctx, service)
		if err != nil {
			errChan <- err
			return dataChan, errChan
		}

		if request.ServiceSource == types.ServiceSourceLocal && service.LocalProvider != "" {
			sp.ProviderName = service.LocalProvider
		} else if request.ServiceSource == types.ServiceSourceRemote && service.RemoteProvider != "" {
			sp.ProviderName = service.RemoteProvider
		}
	}
	err := ds.Get(ctx, sp)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo debug log output
		errChan <- err
		return dataChan, errChan
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		errChan <- err
		return dataChan, errChan
	}
	modelName := request.ModelName
	providerEngine := provider.GetModelEngine(sp.Flavor)
	steam := true
	req := types.PullModelRequest{
		Model:  modelName,
		Stream: &steam,
	}
	dataChan, errChan = providerEngine.PullModelStream(ctx, &req)
	return dataChan, errChan

}

func ModelStreamCancel(ctx context.Context, req *dto.ModelStreamCancelRequest) (*dto.ModelStreamCancelResponse, error) {
	modelClientCancelArray := client.ModelClientMap[req.ModelName]
	if modelClientCancelArray != nil {
		for _, cancel := range modelClientCancelArray {
			cancel()
		}
	}
	return &dto.ModelStreamCancelResponse{
		Bcode: *bcode.ModelCode,
	}, nil
}

func AsyncPullModel(ctx context.Context, sp *types.ServiceProvider, m *types.Model, pullReq *types.PullModelRequest) {
	ds := datastore.GetDefaultDatastore()
	newCtx := context.Background()
	modelEngine := provider.GetModelEngine(sp.Flavor)
	_, err := modelEngine.PullModel(newCtx, pullReq, nil)
	if err != nil {
		slog.Error("Pull model error: ", err.Error())
		m.Status = "failed"
		err = ds.Put(newCtx, m)
		if err != nil {
			return
		}
		return
	}
	slog.Info("Pull model %s completed ..." + m.ModelName)

	m.Status = "downloaded"
	err = ds.Put(newCtx, m)
	if err != nil {
		slog.Error("[Pull model] Update model error:", err.Error())
		return
	}

	if sp.Status == 0 {
		checkServer := ChooseCheckServer(*sp, m.ModelName)
		if checkServer == nil {
			slog.Error("[Pull model] Update service provider error: service_name is not unavailable")
			return
		}
		checkServerStatus := checkServer.CheckServer()
		if !checkServerStatus {
			slog.Error("[Pull model] Update service provider error: server is unavailable")
			return
		}
		err = ds.Get(newCtx, sp)
		if err != nil {
			slog.Error("[Pull model] Update service provider error: ", err.Error())
			return
		}
		sp.Status = 1
		err = ds.Put(newCtx, sp)
		if err != nil {
			slog.Error("[Pull model] Update service provider error: ", err.Error())
			return
		}
		if sp.ServiceName == types.ServiceChat {
			generateSp := new(types.ServiceProvider)
			generateSp.ProviderName = strings.Replace(sp.ProviderName, "chat", "generate", -1)
			err = ds.Get(newCtx, generateSp)
			if err != nil && errors.Is(err, datastore.ErrEntityInvalid) {
				slog.Error("[Pull model] Update service provider error: service provider not found")
				return
			}
			generateCheckServer := ChooseCheckServer(*sp, m.ModelName)
			if generateCheckServer == nil {
				slog.Error("[Pull model] Update service provider error: service_name is not unavailable")
				return
			}
			generateCheckServerStatus := generateCheckServer.CheckServer()
			if !generateCheckServerStatus {
				slog.Error("[Pull model] Update service provider error: server is unavailable")
				return
			}
			generateSp.Status = 1
			err = ds.Put(newCtx, generateSp)
			if err != nil {
				slog.Error("[Pull model] Update service provider error: ", err.Error())
				return
			}
		}

	}
	if sp.ServiceName == types.ServiceChat {
		generateM := &types.Model{}

		generateM.ProviderName = strings.Replace(sp.ProviderName, "chat", "generate", -1)
		generateM.ModelName = m.ModelName
		generateM.Status = "downloaded"

		// Check whether the service provider model already exists.
		generateMIsExist, err := ds.IsExist(newCtx, generateM)
		if !generateMIsExist {
			err = ds.Add(newCtx, generateM)
			if err != nil {
				slog.Error("Add model error: %s", err.Error())
				return
			}
		}
	}
}

type RecommendServicesInfo struct {
	Service             string             `json:"service"`
	MemoryModelsMapList []MemoryModelsInfo `json:"memory_size_models_map_list"`
}

type MemoryModelsInfo struct {
	MemorySize int                      `json:"memory_size"`
	MemoryType []string                 `json:"memory_type"`
	Models     []dto.RecommendModelData `json:"models"`
}

var RecommendModelData = make(map[string][]dto.RecommendModelData)

func RecommendModels() (map[string][]dto.RecommendModelData, error) {
	memoryInfo, err := utils.GetMemoryInfo()
	if err != nil {
		return nil, err
	}
	fileContent, err := template.FlavorTemplateFs.ReadFile("recommend_models.json")
	if err != nil {
		fmt.Printf("Read file failed: %v\n", err)
		return nil, err
	}
	// parse struct
	var serviceModelInfo RecommendServicesInfo
	err = json.Unmarshal(fileContent, &serviceModelInfo)
	if err != nil {
		fmt.Printf("Parse JSON failed: %v\n", err)
		return nil, err
	}
	//Windows system needs to include memory module model detection.
	if runtime.GOOS == "windows" {
		windowsVersion := utils.GetWindowsSystemVersion()
		if windowsVersion < 10 {
			slog.Error("[Model] windows version < 10")
			return nil, bcode.ErrNoRecommendModel
		}
		memoryTypeStatus := false
		for _, memoryModel := range serviceModelInfo.MemoryModelsMapList {
			for _, mt := range memoryModel.MemoryType {
				if memoryInfo.MemoryType == mt {
					memoryTypeStatus = true
					break
				}
			}
			if (memoryModel.MemorySize < memoryInfo.Size) && memoryTypeStatus {
				RecommendModelData[serviceModelInfo.Service] = memoryModel.Models
				return RecommendModelData, nil
			}
		}

	} else {
		// Non-Windows systems determine based only on memory size.
		for _, memoryModel := range serviceModelInfo.MemoryModelsMapList {
			if memoryModel.MemorySize < memoryInfo.Size {
				RecommendModelData[serviceModelInfo.Service] = memoryModel.Models
				return RecommendModelData, nil
			}
		}
	}

	return nil, err
}

func GetRecommendModel() (dto.RecommendModelResponse, error) {
	recommendModel, err := RecommendModels()
	if err != nil {
		return dto.RecommendModelResponse{Data: nil}, err
	}
	return dto.RecommendModelResponse{Bcode: *bcode.ModelCode, Data: recommendModel}, nil
}

func GetSupportModelList(source string, flavor string) (dto.RecommendModelResponse, error) {
	var serviceModelList = make(map[string][]dto.RecommendModelData)
	if source == types.ServiceSourceLocal {
		recommendModel, err := RecommendModels()
		if err != nil {
			return dto.RecommendModelResponse{Data: nil}, err
		}
		flavor = "ollama"
		service := "chat"
		providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, service)
		parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
		for _, model := range recommendModel[service] {
			model.Service = service
			model.Flavor = flavor
			model.Method = parts[0]
			model.Desc = fmt.Sprintf("%s %s %s", source, flavor, parts[1])
			model.Url = providerServiceDefaultInfo.RequestUrl
			model.AuthType = providerServiceDefaultInfo.AuthType
			model.IsRecommended = true
			model.ServiceProvider = fmt.Sprintf("%s %s %s", source, flavor, service)
			serviceModelList[service] = append(serviceModelList[service], model)
		}

	} else {
		for _, service := range types.SupportService {
			if service == types.ServiceModels || service == types.ServiceGenerate {
				continue
			}
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, service)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			authFields := []string{"Api key"}
			if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
				authFields = []string{"SecretId", "SecretKey"}
			}
			for _, model := range providerServiceDefaultInfo.SupportModels {
				modelData := dto.RecommendModelData{
					Name:            model,
					Service:         service,
					Flavor:          flavor,
					Desc:            fmt.Sprintf("%s %s %s", source, flavor, parts[1]),
					Method:          parts[0],
					Url:             providerServiceDefaultInfo.RequestUrl,
					AuthType:        providerServiceDefaultInfo.AuthType,
					AuthFields:      authFields,
					AuthApplyUrl:    providerServiceDefaultInfo.AuthApplyUrl,
					ServiceProvider: fmt.Sprintf("%s %s %s", source, flavor, service),
				}
				serviceModelList[service] = append(serviceModelList[service], modelData)
			}
		}
	}
	return dto.RecommendModelResponse{
		Bcode: *bcode.ModelCode,
		Data:  serviceModelList,
	}, nil
}
