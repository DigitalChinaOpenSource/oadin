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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"strings"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/client"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/provider"
	"intel.com/aog/internal/provider/engine"
	"intel.com/aog/internal/provider/template"
	"intel.com/aog/internal/schedule"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
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
	m.ServiceName = request.ServiceName
	m.ServiceSource = request.ServiceSource

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
		Model:     request.ModelName,
		Stream:    &stream,
		ModelType: sp.ServiceName,
	}
	go AsyncPullModel(sp, m, pullReq)

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
		tmp.ServiceName = dsModel.ServiceName
		tmp.ServiceSource = dsModel.ServiceSource
		tmp.IsDefault = dsModel.IsDefault

		respData = append(respData, *tmp)
	}

	return &dto.GetModelsResponse{
		Bcode: *bcode.ModelCode,
		Data:  respData,
	}, nil
}

func CreateModelStream(ctx context.Context, request dto.CreateModelRequest) (chan []byte, chan error) {
	newDataChan := make(chan []byte, 100)
	newErrChan := make(chan error, 1)
	defer close(newDataChan)
	defer close(newErrChan)
	ds := datastore.GetDefaultDatastore()
	sp := new(types.ServiceProvider)
	if request.ProviderName != "" {
		sp.ProviderName = request.ProviderName
	} else {
		// get default service provider
		// todo Currently only chat and generate services support pulling models.
		if request.ServiceName != types.ServiceChat && request.ServiceName != types.ServiceGenerate && request.ServiceName != types.ServiceEmbed && request.ServiceName != types.ServiceTextToImage && request.ServiceName != types.ServiceSpeechToText {
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

	if request.Size != "" {
		// 判断剩余空间
		providerEngine := provider.GetModelEngine(sp.Flavor)
		var modelSavePath string
		switch eng := providerEngine.(type) {
		case *engine.OpenvinoProvider:
			modelSavePath = fmt.Sprintf("%s/models", eng.EngineConfig.EnginePath)
		case *engine.OllamaProvider:
			modelSavePath = eng.EngineConfig.DownloadPath
		}
		modelSizeGB := utils.ParseSizeToGB(request.Size)
		diskInfo, err := utils.SystemDiskSize(modelSavePath)
		if err != nil {
			newErrChan <- fmt.Errorf("Cannot Get Disk Size: %w", err)
			return newDataChan, newErrChan
		}
		if float64(diskInfo.FreeSize) < modelSizeGB {
			newErrChan <- fmt.Errorf("No enough disk space, need at least %.2f GB", modelSizeGB)
			return newDataChan, newErrChan
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
	m := new(types.Model)
	m.ModelName = request.ModelName
	m.ProviderName = sp.ProviderName
	m.ServiceName = request.ServiceName
	m.ServiceSource = request.ServiceSource

	err = ds.Get(ctx, m)
	if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
		newErrChan <- err
	} else if errors.Is(err, datastore.ErrEntityInvalid) {
		m.Status = "downloading"
		err = ds.Add(ctx, m)
		if err != nil {
			newErrChan <- err
		}
	}
	modelName := request.ModelName
	// todo
	sp.Flavor = strings.Split(sp.ProviderName, "_")[1]
	providerEngine := provider.GetModelEngine(sp.Flavor)
	steam := true
	req := types.PullModelRequest{
		Model:     modelName,
		Stream:    &steam,
		ModelType: request.ServiceName,
	}
	dataChan, errChan := providerEngine.PullModelStream(ctx, &req)

	newDataCh := make(chan []byte, 100)
	newErrorCh := make(chan error, 1)
	go func() {
		defer close(newDataCh)
		defer close(newErrorCh)
		for {
			select {
			case data, ok := <-dataChan:
				if !ok {
					if data == nil {
						client.ModelClientMap[request.ModelName] = nil
						return
					}
				}

				var errResp map[string]interface{}
				if err := json.Unmarshal(data, &errResp); err != nil {
					continue
				}
				if _, ok := errResp["error"]; ok {
					m.Status = "failed"
					err = ds.Put(ctx, m)
					if err != nil {
						newErrorCh <- err
					}
					newErrorCh <- errors.New(string(data))
					return
				}
				var resp types.ProgressResponse
				if err := json.Unmarshal(data, &resp); err != nil {
					log.Printf("Error unmarshaling response: %v", err)

					continue
				}

				if resp.Completed > 0 || resp.Status == "success" {
					if resp.Status == "success" {
						m.Status = "downloaded"
						err = ds.Put(ctx, m)
						if err != nil {
							newErrorCh <- err
							return
						}
						if request.ServiceName == "chat" {
							generateM := new(types.Model)
							generateM.ModelName = m.ModelName
							generateM.ProviderName = strings.Replace(m.ProviderName, "chat", "generate", -1)
							generateM.Status = m.Status
							err = ds.Get(ctx, generateM)
							if err != nil && !errors.Is(err, datastore.ErrEntityInvalid) {
								newErrorCh <- err
								return
							} else if errors.Is(err, datastore.ErrEntityInvalid) {
								err = ds.Add(ctx, generateM)
								if err != nil {
									newErrorCh <- err
								}
								return
							}
							err = ds.Put(ctx, generateM)
							if err != nil {
								newErrorCh <- err
								return
							}
						}
					}
					newDataCh <- data
				}

			case err, ok := <-errChan:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
				client.ModelClientMap[request.ModelName] = nil
				if err != nil && strings.Contains(err.Error(), "context cancel") {
					if strings.Contains(err.Error(), "context cancel") {
						newErrorCh <- err
						return
					} else {
						m.Status = "failed"
						err = ds.Put(ctx, m)
						if err != nil {
							newErrorCh <- err
						}
						return
					}
				}
			case <-ctx.Done():
				newErrorCh <- ctx.Err()
			}
		}
	}()
	return newDataCh, newErrorCh
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

func AsyncPullModel(sp *types.ServiceProvider, m *types.Model, pullReq *types.PullModelRequest) {
	ctx := context.Background()
	ds := datastore.GetDefaultDatastore()
	modelEngine := provider.GetModelEngine(sp.Flavor)
	_, err := modelEngine.PullModel(ctx, pullReq, nil)
	if err != nil {
		logger.LogicLogger.Error("[Pull model] Pull model error: ", err.Error())
		m.Status = "failed"
		err = ds.Put(ctx, m)
		if err != nil {
			return
		}
		return
	}
	logger.LogicLogger.Info("Pull model %s completed ..." + m.ModelName)

	m.Status = "downloaded"
	err = ds.Put(ctx, m)
	if err != nil {
		logger.LogicLogger.Error("[Pull model] Update model error:", err.Error())
		return
	}

	// 查询该 service 下所有模型
	service := &types.Service{Name: sp.ServiceName}
	err = ds.Get(ctx, service)
	if err != nil {
		logger.LogicLogger.Error("[Pull model] Get service error:", err.Error())
		return
	}
	modelList, err := ds.List(ctx, &types.Model{ServiceName: sp.ServiceName}, &datastore.ListOptions{})
	if err != nil {
		logger.LogicLogger.Error("[Pull model] List models error:", err.Error())
		return
	}

	// 查找是否有默认模型
	var hasDefault bool
	var defaultModel *types.Model
	for _, v := range modelList {
		model := v.(*types.Model)
		if model.IsDefault {
			hasDefault = true
			defaultModel = model
			break
		}
	}

	service.Status = 0
	if hasDefault {
		// 只校验默认模型状态
		checkServer := ChooseCheckServer(*sp, defaultModel.ModelName)
		if checkServer != nil && checkServer.CheckServer() {
			service.Status = 1
		}
	} else {
		// 没有默认模型则依次校验
		for _, v := range modelList {
			model := v.(*types.Model)
			checkServer := ChooseCheckServer(*sp, model.ModelName)
			if checkServer != nil && checkServer.CheckServer() {
				service.Status = 1
				break
			}
		}
	}

	err = ds.Put(ctx, service)
	if err != nil {
		logger.LogicLogger.Error("[Pull model] Update service status error:", err.Error())
	}

	if sp.Status == 0 {
		checkServer := ChooseCheckServer(*sp, m.ModelName)
		if checkServer == nil {
			logger.LogicLogger.Error("[Pull model] Update service provider error: service_name is not unavailable")
			return
		}
		checkServerStatus := checkServer.CheckServer()
		if !checkServerStatus {
			logger.LogicLogger.Error("[Pull model] Update service provider error: server is unavailable")
			return
		}
		err = ds.Get(ctx, sp)
		if err != nil {
			logger.LogicLogger.Error("[Pull model] Update service provider error: ", err.Error())
			return
		}
		sp.Status = 1
		err = ds.Put(ctx, sp)
		if err != nil {
			logger.LogicLogger.Error("[Pull model] Update service provider error: ", err.Error())
			return
		}
	}

	currentServiceInfo := schedule.GetProviderServiceDefaultInfo(sp.Flavor, sp.ServiceName)
	providerServices := schedule.GetProviderServices(sp.Flavor)

	for serviceName, serviceInfo := range providerServices {
		if serviceInfo.TaskType == currentServiceInfo.TaskType && serviceName != sp.ServiceName {
			relatedM := &types.Model{}
			relatedM.ModelName = m.ModelName
			relatedM.ProviderName = strings.Replace(sp.ProviderName, sp.ServiceName, serviceName, -1)
			relatedM.Status = "downloaded"

			relatedMIsExist, err := ds.IsExist(ctx, relatedM)
			if err != nil {
				relatedMIsExist = false
			}
			if !relatedMIsExist {
				err = ds.Add(ctx, relatedM)
				if err != nil {
					logger.LogicLogger.Error("Add model error: %s", err.Error())
					return
				}
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

func RecommendModels() (map[string][]dto.RecommendModelData, error) {
	recommendModelDataMap := make(map[string][]dto.RecommendModelData)
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
				recommendModelDataMap[serviceModelInfo.Service] = memoryModel.Models
				return recommendModelDataMap, nil
			}
		}

	} else {
		// Non-Windows systems determine based only on memory size.
		for _, memoryModel := range serviceModelInfo.MemoryModelsMapList {
			if memoryModel.MemorySize < memoryInfo.Size {
				recommendModelDataMap[serviceModelInfo.Service] = memoryModel.Models
				return recommendModelDataMap, nil
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
		//var recommendModelParamsSize float32
		recommendModel, err := RecommendModels()
		if err != nil {
			return &dto.RecommendModelResponse{Data: nil}, err
		}
		flavor = "ollama"
		// service := "chat"
		var resModelNameList []string

		for modelService, modelInfo := range recommendModel {
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, modelService)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			for _, model := range modelInfo {
				localModelInfo := localOllamaModelMap[model.Name]
				modelQuery := new(types.Model)
				modelQuery.ModelName = model.Name
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
				resModelNameList = append(resModelNameList, model.Name)
			}

		}
		for modelService, modelInfo := range localOllamaServiceMap {
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, modelService)
			if providerServiceDefaultInfo.Endpoints == nil {
				continue
			}
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			for _, localModel := range modelInfo {
				if !utils.Contains(resModelNameList, localModel.Name) {
					modelQuery := new(types.Model)
					modelQuery.ModelName = localModel.Name
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
					model.Size = localModel.Size
					model.OllamaId = localModel.OllamaId
					serviceModelList[modelService] = append(serviceModelList[modelService], *model)
					resModelNameList = append(resModelNameList, model.Name)
				}
			}

		}

	} else {
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
		for _, service := range types.SupportService {
			if service == types.ServiceModels || service == types.ServiceGenerate {
				continue
			}
			remoteModelInfoList := RemoteServiceMap[service]
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(flavor, service)
			parts := strings.SplitN(providerServiceDefaultInfo.Endpoints[0], " ", 2)
			authFields := []string{"api_key"}
			if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
				authFields = []string{"secret_id", "secret_key"}
			}
			for _, model := range remoteModelInfoList {
				if model.Flavor != flavor {
					continue
				}
				modelQuery := new(types.Model)
				modelQuery.ModelName = model.Name
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
					Name:            model.Name,
					Avatar:          model.Avatar,
					Desc:            model.Description,
					Service:         service,
					Flavor:          flavor,
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
