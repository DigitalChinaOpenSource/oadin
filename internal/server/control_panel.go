package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
)

func GetModelFilePath(ctx context.Context) (*dto.GetModelFilePathResponse, error) {
	var defaultPath string
	userDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	defaultPath = filepath.Join(userDir, ".ollama")
	res := &dto.GetModelFilePathData{}
	value := os.Getenv("BYZE_OLLAMA_MODELS")
	if value != "" {
		res.Path = value
	} else {
		res.Path = defaultPath
	}

	return &dto.GetModelFilePathResponse{
		*bcode.ControlPanelCode,
		res,
	}, nil
}

func GetFilePathSize(ctx context.Context, req *dto.GetPathDiskSizeInfoRequest) (*dto.GetPathDiskSizeInfoResponse, error) {
	res := &dto.GetPathDiskSizeInfoResponse{}
	sizeInfo, err := utils.SystemDiskSize(req.Path)
	if err != nil {
		return nil, err
	}
	res.Data = sizeInfo
	res.Bcode = *bcode.ControlPanelCode
	return res, nil
}

func ModifyModelFilePath(ctx context.Context, req *dto.ModifyModelFilePathRequest) (*dto.ModifyModelFilePathResponse, error) {
	if req.TargetPath == req.SourcePath {
		return &dto.ModifyModelFilePathResponse{}, errors.New("target path is the same as the source path")
	}
	isTargetDirEmpty := utils.IsDirEmpty(req.TargetPath)
	if !isTargetDirEmpty {
		return &dto.ModifyModelFilePathResponse{}, errors.New("target path is not empty")
	}

	isSourceDirEmpty := utils.IsDirEmpty(req.SourcePath)
	if !isSourceDirEmpty {
		status, err := utils.SamePartitionStatus(req.SourcePath, req.TargetPath)
		if err != nil {
			return nil, err
		}
		if status {
			err = utils.CopyDir(req.SourcePath, req.TargetPath)
			if err != nil {
				return nil, err
			}
		} else {
			sourcePathSize, err := utils.GetFilePathTotalSize(req.SourcePath)
			if err != nil {
				return nil, err
			}
			targetDiskSizeInfo, err := utils.SystemDiskSize(req.TargetPath)
			if err != nil {
				return nil, err
			}
			if int(sourcePathSize) > targetDiskSizeInfo.FreeSize {
				return nil, errors.New("Target file path size is not enough")
			}
			err = utils.CopyDir(req.SourcePath, req.TargetPath)
			if err != nil {
				return nil, err
			}
		}
	}

	envInfo := &utils.EnvVariables{
		Name:  "BYZE_OLLAMA_MODELS",
		Value: req.TargetPath,
	}
	err := utils.ModifySystemUserVariables(envInfo)
	if err != nil {
		err = os.RemoveAll(req.TargetPath)
		if err != nil {
			return nil, err
		}
		return nil, err
	}
	os.Setenv("BYZE_OLLAMA_MODELS", req.TargetPath)

	// 在删除源路径前先停止引擎，避免进程仍在使用文件导致错误
	engine := provider.GetModelEngine("ollama")
	_ = engine.StopEngine()

	time.Sleep(1 * time.Second)

	err = os.RemoveAll(req.SourcePath)
	if err != nil {
		return nil, err
	}

	err = engine.InitEnv()
	if err != nil {
		return nil, err
	}

	err = engine.StartEngine()
	if err != nil {
		return nil, err
	}

	res := &dto.ModifyModelFilePathResponse{}
	res.Bcode = *bcode.ControlPanelCode
	res.Data = struct{}{}
	return res, nil
}

func GetModelList(ctx context.Context, request dto.GetModelListRequest) (*dto.RecommendModelResponse, error) {
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
				// model.Class = localModelInfo.Class
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
					// model.Class = localModel.Class
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
				if flavor != "" && model.Flavor != flavor {
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
