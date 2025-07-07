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
	// "encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/schedule"
	"intel.com/aog/internal/types"
	"intel.com/aog/version"

	// "intel.com/aog/internal/utils"
	"intel.com/aog/internal/utils/bcode"
)

func myModelFilter(modelList *[]dto.RecommendModelData) {
	var finalDataList []dto.RecommendModelData
	if modelList == nil || len(*modelList) == 0 {
		return
	}

	var tempList []dto.RecommendModelData = *modelList
	for i := len(tempList) - 1; i >= 0; i-- {
		if tempList[i].CanSelect {
			finalDataList = append(finalDataList, tempList[i])
		}
	}
	// 数据回填
	*modelList = finalDataList
}

func GetSupportModelListCombine(ctx context.Context, request *dto.GetSupportModelRequest) (*dto.GetSupportModelResponse, error) {
	jds := datastore.GetDefaultJsonDatastore()
	if jds == nil {
		return nil, errors.New("json datastore is nil, please check initialization")
	}
	ds := datastore.GetDefaultDatastore()

	if request.ServiceName == types.ServiceGenerate {
		request.ServiceName = types.ServiceChat
	}

	// 分页参数
	page := request.Page
	if page == 0 {
		page = 1
	}
	pageSize := request.PageSize
	if pageSize == 0 {
		pageSize = 10
	}
	var resData dto.GetSupportModelResponseData
	resData.PageSize = pageSize
	resData.Page = page

	// 查询及排序
	queryOpList := []datastore.FuzzyQueryOption{}
	if request.SearchName != "" {
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "name",
			Query: request.SearchName,
		})
	}
	if request.Flavor != "" {
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "flavor",
			Query: request.Flavor,
		})
	}
	if request.ServiceSource != "" {
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "service_source",
			Query: request.ServiceSource,
		})
	}
	if request.ServiceName != "" {
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "service_name",
			Query: request.ServiceName,
		})
	}
	sm := &types.SupportModel{}
	sortOption := []datastore.SortOption{
		{Key: "id", Order: 1},
	}
	options := &datastore.ListOptions{
		FilterOptions: datastore.FilterOptions{Queries: queryOpList},
		SortBy:        sortOption,
		// 不分页，查全部
	}

	var allModels []dto.RecommendModelData
	defaultIdx := -1

	if request.ServiceSource == types.ServiceSourceLocal {
		if request.Flavor != "" && request.Flavor != types.FlavorOllama {
			return nil, errors.New(fmt.Sprintf("%s flavor is not local flavor", request.Flavor))
		}
		// 查全部
		supportModelList, err := jds.List(ctx, sm, options)
		if err != nil {
			return nil, err
		}
		recommendModel, _ := RecommendModels()
		for _, supportModel := range supportModelList {
			IsRecommend := false
			smInfo := supportModel.(*types.SupportModel)
			if smInfo.ApiFlavor == types.FlavorOllama {
				if recommendModel != nil {
					rmServiceModelInfo := recommendModel[smInfo.ServiceName]
					if rmServiceModelInfo != nil {
						for _, rm := range rmServiceModelInfo {
							if rm.Name == smInfo.Name {
								IsRecommend = true
								break
							}
						}
					}
				}
			}

			providerName := fmt.Sprintf("%s_%s_%s", smInfo.ServiceSource, smInfo.ApiFlavor, smInfo.ServiceName)

			// todo
			if smInfo.ServiceName == types.ServiceTextToImage || smInfo.ServiceName == types.ServiceSpeechToText {
				providerName = fmt.Sprintf("%s_%s_%s", smInfo.ServiceSource, types.FlavorOpenvino, smInfo.ServiceName)
			}

			modelQuery := &types.Model{
				ModelName:    smInfo.Name,
				ProviderName: providerName,
			}
			canSelect := true
			err := ds.Get(context.Background(), modelQuery)
			if err != nil || modelQuery.Status != "downloaded" {
				canSelect = false
			}
			isDefault := false
			if modelQuery.IsDefault {
				isDefault = true
			}
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(smInfo.Flavor, smInfo.ServiceName)
			authFields := []string{""}
			if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
				authFields = []string{"secret_id", "secret_key"}
			} else if providerServiceDefaultInfo.AuthType == types.AuthTypeApiKey {
				authFields = []string{"api_key"}
			}
			modelData := dto.RecommendModelData{
				Id:              smInfo.Id,
				Name:            smInfo.Name,
				Avatar:          smInfo.Avatar,
				Desc:            smInfo.Description,
				Service:         smInfo.ServiceName,
				ApiFlavor:       smInfo.ApiFlavor,
				Flavor:          smInfo.Flavor,
				AuthType:        providerServiceDefaultInfo.AuthType,
				AuthFields:      authFields,
				AuthApplyUrl:    providerServiceDefaultInfo.AuthApplyUrl,
				ServiceProvider: providerName,
				CanSelect:       canSelect,
				IsRecommended:   IsRecommend,
				Source:          smInfo.ServiceSource,
				InputLength:     smInfo.MaxInput,
				OutputLength:    smInfo.MaxOutput,
				Class:           smInfo.Class,
				Size:            smInfo.Size,
				OllamaId:        smInfo.OllamaId,
				IsDefault:       fmt.Sprintf("%v", isDefault),
			}
			if isDefault {
				defaultIdx = len(allModels)
			}
			allModels = append(allModels, modelData)
		}
	} else {
		// 远程模型
		supportModelList, err := jds.List(ctx, sm, options)
		if err != nil {
			return nil, err
		}
		for _, supportModel := range supportModelList {
			smInfo := supportModel.(*types.SupportModel)
			providerName := fmt.Sprintf("%s_%s_%s", smInfo.ServiceSource, smInfo.Flavor, smInfo.ServiceName)
			modelQuery := &types.Model{
				ModelName:     smInfo.Name,
				ProviderName:  providerName,
				ServiceName:   smInfo.ServiceName,
				ServiceSource: smInfo.ServiceSource,
			}
			canSelect := true
			err := ds.Get(context.Background(), modelQuery)
			if err != nil || modelQuery.Status != "downloaded" {
				canSelect = false
			}
			isDefault := false
			if modelQuery.IsDefault {
				isDefault = true
			}
			providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(smInfo.Flavor, smInfo.ServiceName)
			authFields := []string{""}
			if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
				authFields = []string{"secret_id", "secret_key"}
			} else if providerServiceDefaultInfo.AuthType == types.AuthTypeApiKey {
				authFields = []string{"api_key"}
			}
			modelData := dto.RecommendModelData{
				Id:              smInfo.Id,
				Name:            smInfo.Name,
				Avatar:          smInfo.Avatar,
				Desc:            smInfo.Description,
				Service:         smInfo.ServiceName,
				ApiFlavor:       smInfo.ApiFlavor,
				Flavor:          smInfo.Flavor,
				AuthType:        providerServiceDefaultInfo.AuthType,
				AuthFields:      authFields,
				AuthApplyUrl:    providerServiceDefaultInfo.AuthApplyUrl,
				ServiceProvider: providerName,
				CanSelect:       canSelect,
				IsRecommended:   true,
				Source:          smInfo.ServiceSource,
				InputLength:     smInfo.MaxInput,
				OutputLength:    smInfo.MaxOutput,
				Class:           smInfo.Class,
				Size:            smInfo.Size,
				OllamaId:        smInfo.OllamaId,
				IsDefault:       fmt.Sprintf("%v", isDefault),
			}
			if isDefault {
				defaultIdx = len(allModels)
			}
			allModels = append(allModels, modelData)
		}
	}

	// 将IsDefault的模型放到第一个
	if defaultIdx > 0 {
		defaultModel := allModels[defaultIdx]
		allModels = append(allModels[:defaultIdx], allModels[defaultIdx+1:]...)
		allModels = append([]dto.RecommendModelData{defaultModel}, allModels...)
	}

	// mine 过滤
	if request.Mine {
		myModelFilter(&allModels)
	}

	// 分页
	total := len(allModels)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	resultList := allModels[start:end]

	// 计算总页数
	totalPage := total / pageSize
	if total%pageSize != 0 {
		totalPage++
	}
	if totalPage == 0 {
		totalPage = 1
	}

	resData.Total = total
	resData.TotalPage = totalPage
	resData.Data = resultList

	return &dto.GetSupportModelResponse{
		*bcode.ModelCode,
		resData,
	}, nil
}

func SetDefaultModel(ctx context.Context, req *dto.SetDefaultModelRequest) error {
	ds := datastore.GetDefaultDatastore()
	m := &types.Model{
		ServiceName:   req.ServiceName,
		ServiceSource: req.ServiceSource,
	}
	list, err := ds.List(ctx, m, &datastore.ListOptions{Page: 0, PageSize: 1000})
	if err != nil {
		return err
	}
	var found bool
	for _, v := range list {
		model := v.(*types.Model)
		if model.ModelName == req.ModelName {
			if model.Status != "downloaded" {
				return errors.New("model must be downloaded to set as default")
			}
			if model.IsDefault {
				model.IsDefault = false
			} else {
				model.IsDefault = true
			}
			found = true
		} else {
			model.IsDefault = false
		}
		if err := ds.Put(ctx, model); err != nil {
			return err
		}
	}
	if !found {
		return errors.New("model not found")
	}
	return nil
}

func GetDashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	ds := datastore.GetDefaultDatastore()
	jds := datastore.GetDefaultJsonDatastore()

	// 获取所有模型
	modelList, err := ds.List(ctx, &types.Model{}, &datastore.ListOptions{
		Page: 0, PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}
	var models []dto.Model
	for _, v := range modelList {
		m := v.(*types.Model)
		if strings.ToLower(m.Status) != "downloaded" {
			continue // 只保留已下载的模型
		}
		// 查找对应的 SupportModel 获取 Avatar
		sm := &types.SupportModel{
			Name:          m.ModelName,
			ServiceName:   m.ServiceName,
			ServiceSource: m.ServiceSource,
		}
		smList, err := jds.List(ctx, sm, nil)
		avatar := ""
		if err == nil && len(smList) > 0 {
			for _, s := range smList {
				supportModel := s.(*types.SupportModel)
				if supportModel.Name == m.ModelName {
					avatar = supportModel.Avatar
					break
				}
			}
		}
		models = append(models, dto.Model{
			ModelName:     m.ModelName,
			ProviderName:  m.ProviderName,
			Status:        m.Status,
			ServiceName:   m.ServiceName,
			ServiceSource: m.ServiceSource,
			IsDefault:     m.IsDefault,
			CreatedAt:     m.CreatedAt,
			UpdatedAt:     m.UpdatedAt,
			Avatar:        avatar,
		})
	}

	sort.SliceStable(models, func(i, j int) bool {
		return models[i].UpdatedAt.After(models[j].UpdatedAt)
	})

	sort.SliceStable(models, func(i, j int) bool {
		return models[i].IsDefault && !models[j].IsDefault
	})
	// 只查 can_install = true 的服务
	serviceList, err := ds.List(ctx, &types.Service{}, &datastore.ListOptions{
		FilterOptions: datastore.FilterOptions{
			Queries: []datastore.FuzzyQueryOption{
				{Key: "can_install", Query: "1"},
			},
		},
		Page: 0, PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	var services []dto.Service
	for _, v := range serviceList {
		s := v.(*types.Service)
		services = append(services, dto.Service{
			ServiceName:    s.Name,
			LocalProvider:  s.LocalProvider,
			RemoteProvider: s.RemoteProvider,
			HybridPolicy:   s.HybridPolicy,
			Status:         s.Status,
			Avatar:         s.Avatar,
			CreatedAt:      s.CreatedAt,
			UpdatedAt:      s.UpdatedAt,
		})
	}

	return &dto.DashboardResponse{
		Models:   models,
		Services: services,
	}, nil
}

func GetProductInfo(ctx context.Context) (*dto.GetProductInfoResponse, error) {
	return &dto.GetProductInfoResponse{
		Icon:        version.AOGIcon,
		ProductName: version.AOGName,
		Description: version.AOGDescription,
		Version:     version.AOGVersion,
	}, nil
}

func GetModelkey(ctx context.Context, req *dto.GetModelkeyRequest) (*dto.GetModelkeyResponse, error) {
	ds := datastore.GetDefaultDatastore()

	// 构造查询条件
	sp := &types.ServiceProvider{
		ProviderName: req.ProviderName,
	}
	err := ds.Get(ctx, sp)
	if err != nil {
		return nil, err
	}
	if sp.AuthKey == "" {
		return &dto.GetModelkeyResponse{
			ModelKey: "",
		}, nil
	}
	return &dto.GetModelkeyResponse{
		ModelKey: sp.AuthKey,
	}, nil
}
