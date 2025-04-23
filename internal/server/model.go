package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"byze/internal/api/dto"
	"byze/internal/datastore"
	"byze/internal/provider"
	"byze/internal/provider/template"
	"byze/internal/schedule"
	"byze/internal/types"
	"byze/internal/utils"
	"byze/internal/utils/bcode"
	"byze/internal/utils/client"
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
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate && request.ServiceName != types.ServiceEmbed &&
			request.ServiceName != types.ServiceTextToImage {
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
	if sp.ServiceSource == types.ServiceSourceLocal {
		m.ModelName = strings.ToLower(request.ModelName)
	}
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
	if request.ServiceSource == types.ServiceSourceLocal {
		m.ModelName = strings.ToLower(request.ModelName)
	}

	err = s.Ds.Get(ctx, m)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		// todo err debug log output
		return nil, bcode.ErrServer
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		return nil, bcode.ErrModelRecordNotFound
	}

	//Call engin to delete model.
	if m.Status == "downloaded" {
		modelEngine := provider.GetModelEngine(sp.Flavor)
		deleteReq := &types.DeleteRequest{
			Model: request.ModelName,
		}

		err = modelEngine.DeleteModel(ctx, deleteReq)
		if err != nil {
			// todo err debug log output
			return nil, bcode.ErrEngineDeleteModel
		}
	}

	err = s.Ds.Delete(ctx, m)
	if err != nil {
		// todo err debug log output
		return nil, err
	}
	if request.ServiceName == types.ServiceChat {
		generateM := types.Model{
			ProviderName: strings.Replace(request.ProviderName, "chat", "generate", -1),
			ModelName:    m.ModelName,
		}
		err = s.Ds.Get(ctx, &generateM)
		if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
			return nil, err
		}
		err = s.Ds.Delete(ctx, &generateM)
		if err != nil {
			return nil, err
		}
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
	newDataChan := make(chan []byte)
	newErrChan := make(chan error)
	defer close(newDataChan)
	defer close(newErrChan)
	ds := datastore.GetDefaultDatastore()
	sp := new(types.ServiceProvider)
	if request.ProviderName != "" {
		sp.ProviderName = request.ProviderName
	} else {
		// get default service provider
		// todo Currently only chat and generate services support pulling models.
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate && request.ServiceName != types.ServiceEmbed {
			newErrChan <- bcode.ErrServer
			return newDataChan, newErrChan
		}

		service := &types.Service{}
		service.Name = request.ServiceName

		err := ds.Get(ctx, service)
		if err != nil {
			newErrChan <- err
			return newDataChan, newErrChan
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
		newErrChan <- err
		return newDataChan, newErrChan
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		newErrChan <- err
		return newDataChan, newErrChan
	}
	modelName := request.ModelName
	providerEngine := provider.GetModelEngine(sp.Flavor)
	steam := true
	req := types.PullModelRequest{
		Model:  modelName,
		Stream: &steam,
	}
	dataChan, errChan := providerEngine.PullModelStream(ctx, &req)
	return dataChan, errChan
}

func ModelStreamCancel(ctx context.Context, req *dto.ModelStreamCancelRequest) (*dto.ModelStreamCancelResponse, error) {
	modelClientCancelArray := client.ModelClientMap[strings.ToLower(req.ModelName)]
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
	if sp.ServiceSource == types.ServiceSourceLocal {

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
	}

	slog.Info("Pull model %s completed ..." + m.ModelName)

	m.Status = "downloaded"
	err := ds.Put(newCtx, m)
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
	// Windows system needs to include memory module model detection.
	if runtime.GOOS == "windows" {
		windowsVersion := utils.GetSystemVersion()
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

func GetSupportModelList(ctx context.Context, request dto.GetModelListRequest) (*dto.RecommendModelResponse, error) {
	ds := datastore.GetDefaultDatastore()
	flavor := request.Flavor
	source := request.ServiceSource
	serviceModelList := make(map[string][]dto.RecommendModelData)
	if request.ServiceSource == types.ServiceSourceLocal {
		localOllamaModelMap := make(map[string]dto.LocalSupportModelData)
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
		for _, serviceInfo := range localOllamaServiceMap {
			for _, model := range serviceInfo {
				localOllamaModelMap[model.Name] = model
			}
		}
		//
		var recommendModelParamsSize float32
		recommendModel, err := RecommendModels()
		if err != nil {
			return &dto.RecommendModelResponse{Data: nil}, err
		}
		flavor = "ollama"
		//service := "chat"
		var resModelNameList []string

		for modelService, modelInfo := range recommendModel {
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, modelService)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			for _, model := range modelInfo {
				localModelInfo := localOllamaModelMap[model.Name]
				modelQuery := new(types.Model)
				modelQuery.ModelName = strings.ToLower(model.Name)
				modelQuery.ProviderName = fmt.Sprintf("%s_%s_%s", source, flavor, modelService)
				canSelect := true
				err := ds.Get(context.Background(), modelQuery)
				if err != nil {
					canSelect = false
				}
				if modelQuery.Status != "downloaded" {
					canSelect = false
				}
				model.Service = modelService
				model.Flavor = flavor
				model.Method = parts[0]
				model.Desc = localModelInfo.Description
				model.Url = providerServiceDefaultInfo.RequestUrl
				model.AuthType = providerServiceDefaultInfo.AuthType
				model.IsRecommended = true
				model.CanSelect = canSelect
				model.ServiceProvider = fmt.Sprintf("%s_%s_%s", source, flavor, modelService)
				model.Avatar = localModelInfo.Avatar
				model.Class = localModelInfo.Class
				model.OllamaId = localModelInfo.OllamaId
				serviceModelList[modelService] = append(serviceModelList[modelService], model)
				recommendModelParamsSize = model.ParamsSize
				resModelNameList = append(resModelNameList, model.Name)
			}

		}
		for modelService, modelInfo := range localOllamaServiceMap {
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, modelService)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			for _, localModel := range modelInfo {
				if !utils.Contains(resModelNameList, localModel.Name) {
					modelQuery := new(types.Model)
					modelQuery.ModelName = strings.ToLower(localModel.Name)
					modelQuery.ProviderName = fmt.Sprintf("%s_%s_%s", source, flavor, modelService)
					canSelect := true
					err := ds.Get(context.Background(), modelQuery)
					if err != nil {
						canSelect = false
					}
					if modelQuery.Status != "downloaded" {
						canSelect = false
					}
					model := new(dto.RecommendModelData)
					model.Name = localModel.Name
					model.Service = modelService
					model.Flavor = flavor
					model.Method = parts[0]
					model.Desc = localModel.Description
					model.Url = providerServiceDefaultInfo.RequestUrl
					model.AuthType = providerServiceDefaultInfo.AuthType
					model.IsRecommended = false
					model.CanSelect = canSelect
					model.ServiceProvider = fmt.Sprintf("%s_%s_%s", source, flavor, modelService)
					model.Avatar = localModel.Avatar
					model.Class = localModel.Class
					model.OllamaId = localModel.OllamaId
					serviceModelList[modelService] = append(serviceModelList[modelService], *model)
					resModelNameList = append(resModelNameList, model.Name)
				}
			}

		}

	} else {
		for _, service := range types.SupportService {
			if service == types.ServiceModels || service == types.ServiceGenerate {
				continue
			}
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, service)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			authFields := []string{"api_key"}
			if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
				authFields = []string{"secret_id", "secret_key"}
			}
			for _, model := range providerServiceDefaultInfo.SupportModels {
				modelQuery := new(types.Model)
				modelQuery.ModelName = model
				modelQuery.ProviderName = fmt.Sprintf("%s_%s_%s", source, flavor, service)
				canSelect := true
				err := ds.Get(context.Background(), modelQuery)
				if err != nil {
					canSelect = false
				}
				if modelQuery.Status != "downloaded" {
					canSelect = false
				}
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
					ServiceProvider: fmt.Sprintf("%s_%s_%s", source, flavor, service),
					CanSelect:       canSelect,
				}
				serviceModelList[service] = append(serviceModelList[service], modelData)
			}
		}
	}
	return &dto.RecommendModelResponse{
		Bcode: *bcode.ModelCode,
		Data:  serviceModelList,
	}, nil
}

func GetSupportSmartVisionModels(ctx context.Context, request *dto.SmartVisionSupportModelRequest) (*dto.SmartVisionSupportModelResponse, error) {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport}
	envType := request.EnvType
	smartVisionUrlMap := utils.GetSmartVisionUrl()
	smartVisionUrl := smartVisionUrlMap[envType]
	modelUrl := smartVisionUrl.Url + "/admin-api/api/model/list"
	req, err := http.NewRequest("GET", modelUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+smartVisionUrl.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, bcode.ErrServer
	}
	var res dto.SmartVisionSupportModelRes
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, bcode.ErrServer
	}

	ds := datastore.GetDefaultDatastore()
	var resData []dto.SmartVisionModelData
	for _, model := range res.Data {
		modelQuery := new(types.Model)
		modelQuery.ModelName = model.Name

		canSelect := true
		err := ds.Get(ctx, modelQuery)
		if err != nil {
			canSelect = false
		}

		if modelQuery.Status != "downloaded" {
			canSelect = false
		}
		model.CanSelect = canSelect
		resData = append(resData, model)
	}
	res.Data = resData
	return &dto.SmartVisionSupportModelResponse{
		Bcode: *bcode.ModelCode,
		Data:  res,
	}, nil
}
