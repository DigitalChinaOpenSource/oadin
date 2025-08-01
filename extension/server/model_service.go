package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"oadin/internal/logger"
	"runtime"
	"sort"
	"strconv"
	"time"

	"oadin/extension/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/provider/template"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils"
	"oadin/internal/utils/bcode"
)

func GetSupportSmartVisionModels(ctx context.Context, request *dto.SmartVisionSupportModelRequest) (*dto.SmartVisionSupportModelResponse, error) {
	var res dto.SmartVisionSupportModelRes
	resData, err := GetSmartVisionModelData(ctx, request.EnvType)
	if err != nil {
		return &dto.SmartVisionSupportModelResponse{}, err
	}
	res.Data = resData
	return &dto.SmartVisionSupportModelResponse{
		Bcode: *bcode.ModelCode,
		Data:  res,
	}, nil
}

func GetSmartVisionModelData(ctx context.Context, envType string) ([]dto.SmartVisionModelData, error) {
	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: transport}
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

		if canSelect {
			model.CreatedAt = modelQuery.CreatedAt
		}

		resData = append(resData, model)
	}
	return resData, nil
}

func GetSupportModelListCombine(ctx context.Context, request *dto.GetSupportModelRequest) (*dto.GetSupportModelResponse, error) {
	jds := datastore.GetDefaultJsonDatastore()
	ds := datastore.GetDefaultDatastore()
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
	resultList := []dto.RecommendModelData{}
	queryOpList := []datastore.FuzzyQueryOption{}
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
	sm := &types.SupportModel{}
	sortOption := []datastore.SortOption{
		{Key: "name", Order: 1},
	}
	options := &datastore.ListOptions{FilterOptions: datastore.FilterOptions{Queries: queryOpList}, SortBy: sortOption}
	if request.ServiceSource == types.ServiceSourceLocal {
		if request.Flavor != "" && request.Flavor != types.FlavorOllama {
			return nil, errors.New(fmt.Sprintf("%s flavor is not local flavor", request.Flavor))
		}
		totalCount, err := jds.Count(ctx, sm, &datastore.FilterOptions{Queries: queryOpList})
		if err != nil {
			return nil, err
		}
		resData.Total = int(totalCount)
		if int(totalCount)%pageSize == 0 {
			resData.TotalPage = int(totalCount) / pageSize
		} else {
			resData.TotalPage = int(totalCount)/pageSize + 1
		}
		if resData.TotalPage == 0 {
			resData.TotalPage = 1
		}
		options.Page = page
		options.PageSize = pageSize
		supportModelList, err := jds.List(ctx, sm, options)
		if err != nil {
			return nil, err
		}

		recommendModel, _ := RecommendModels()
		for _, supportModel := range supportModelList {
			IsRecommend := false
			smInfo := supportModel.(*types.SupportModel)
			if smInfo.ApiFlavor == types.FlavorOllama {
				if recommendModel == nil {
					IsRecommend = false
				}
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

			providerName := fmt.Sprintf("%s_%s_%s", smInfo.ServiceSource, types.FlavorOllama, smInfo.ServiceName)
			modelQuery := new(types.Model)
			modelQuery.ModelName = smInfo.Name
			modelQuery.ProviderName = providerName
			canSelect := true
			err := ds.Get(context.Background(), modelQuery)
			if err != nil {
				canSelect = false
			}
			if modelQuery.Status != "downloaded" {
				canSelect = false
			}

			if canSelect {
				smInfo.CreatedAt = modelQuery.CreatedAt
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
				ApiFlavor:       types.FlavorOllama,
				Flavor:          smInfo.Flavor,
				AuthType:        providerServiceDefaultInfo.AuthType,
				AuthFields:      authFields,
				AuthApplyUrl:    providerServiceDefaultInfo.AuthApplyUrl,
				ServiceProvider: fmt.Sprintf("%s_%s_%s", smInfo.ServiceSource, types.FlavorOllama, smInfo.ServiceName),
				CanSelect:       canSelect,
				IsRecommended:   IsRecommend,
				Source:          smInfo.ServiceSource,
				InputLength:     smInfo.InputLength,
				OutputLength:    smInfo.OutputLength,
				Class:           smInfo.Class,
				Size:            smInfo.Size,
				OllamaId:        smInfo.OllamaId,
				Think:           smInfo.Think,
				ThinkSwitch:     smInfo.ThinkSwitch,
				Tools:           smInfo.Tools,
				Context:         smInfo.Context,
				CreatedAt:       smInfo.CreatedAt,
			}
			resultList = append(resultList, modelData)

		}
		// 数据处理, 如果是我的模型数据，则进行数据过滤-> 使用canSelect过滤
		if request.Mine {
			myModelFilter(&resultList)
			resData.Total = len(resultList)
			resData.TotalPage = len(resultList) / pageSize
			if resData.TotalPage == 0 {
				resData.TotalPage = 1
			}
		}
	} else {
		if request.Flavor == types.FlavorSmartVision {
			smartvisionModelData, err := GetSmartVisionModelData(ctx, request.EnvType)
			if err != nil {
				return nil, err
			}
			dataStart := (page - 1) * pageSize
			dataEnd := page * pageSize
			if dataEnd > len(smartvisionModelData) {
				dataEnd = len(smartvisionModelData) - 1
			}
			pageData := smartvisionModelData[dataStart:dataEnd]
			// dataList := []dto.RecommendModelData{}
			for _, d := range pageData {
				providerName := fmt.Sprintf("%s_%s_%s", request.ServiceSource, request.Flavor, "chat")
				modelQuery := new(types.Model)
				modelQuery.ModelName = d.Name
				modelQuery.ProviderName = providerName
				canSelect := true
				err := ds.Get(context.Background(), modelQuery)
				if err != nil {
					canSelect = false
				}
				if modelQuery.Status != "downloaded" {
					canSelect = false
				}
				var authFields []string
				for _, cInfo := range d.CredentialParams {
					authFields = append(authFields, cInfo.Name)
				}
				modelData := dto.RecommendModelData{
					Id:                  strconv.Itoa(d.ID),
					Name:                d.Name,
					Avatar:              d.Avatar,
					Desc:                d.Introduce,
					Service:             types.ServiceChat,
					ApiFlavor:           types.FlavorSmartVision,
					Flavor:              d.Provider,
					AuthType:            types.AuthTypeCredentials,
					AuthFields:          authFields,
					AuthApplyUrl:        "",
					ServiceProvider:     fmt.Sprintf("%s_%s_%s", request.ServiceSource, request.Flavor, types.ServiceChat),
					CanSelect:           canSelect,
					IsRecommended:       false,
					Source:              types.ServiceSourceRemote,
					Class:               d.Tags,
					SmartVisionProvider: d.Provider,
					SmartVisionModelKey: d.ModelKey,
					CreatedAt:           d.CreatedAt,
				}
				resultList = append(resultList, modelData)
			}
			// 数据处理, 如果是我的模型数据，则进行数据过滤-> 使用canSelect过滤
			if request.Mine {
				myModelFilter(&resultList)
			}
			resData.Total = len(smartvisionModelData)
			resData.TotalPage = len(smartvisionModelData) / pageSize
			if resData.TotalPage == 0 {
				resData.TotalPage = 1
			}
		} else if request.Flavor == "" {
			smartvisionModelData, smartvisionErr := GetSmartVisionModelData(ctx, request.EnvType)
			jdsDataList, jdsErr := jds.List(ctx, sm, options)
			if smartvisionErr != nil || jdsErr != nil {
				return nil, errors.New("Get data Failed, please retry")
			}
			for _, d := range smartvisionModelData {
				providerName := fmt.Sprintf("%s_%s_%s", request.ServiceSource, types.FlavorSmartVision, "chat")
				modelQuery := new(types.Model)
				modelQuery.ModelName = d.Name
				modelQuery.ProviderName = providerName
				canSelect := true
				err := ds.Get(context.Background(), modelQuery)
				if err != nil {
					canSelect = false
				}
				if modelQuery.Status != "downloaded" {
					canSelect = false
				}
				var authFields []string
				for _, cInfo := range d.CredentialParams {
					authFields = append(authFields, cInfo.Name)
				}
				modelData := dto.RecommendModelData{
					Id:                  strconv.Itoa(d.ID),
					Name:                d.Name,
					Avatar:              d.Avatar,
					Desc:                d.Introduce,
					Service:             types.ServiceChat,
					ApiFlavor:           types.FlavorSmartVision,
					Flavor:              types.FlavorSmartVision,
					AuthType:            types.AuthTypeCredentials,
					AuthFields:          authFields,
					AuthApplyUrl:        "",
					ServiceProvider:     fmt.Sprintf("%s_%s_%s", request.ServiceSource, types.FlavorSmartVision, types.ServiceChat),
					CanSelect:           canSelect,
					IsRecommended:       false,
					Source:              types.ServiceSourceRemote,
					Class:               d.Tags,
					SmartVisionProvider: d.Provider,
					SmartVisionModelKey: d.ModelKey,
					CreatedAt:           d.CreatedAt,
				}
				resultList = append(resultList, modelData)
			}
			for _, jdModel := range jdsDataList {
				jdModelInfo := jdModel.(*types.SupportModel)
				providerName := fmt.Sprintf("%s_%s_%s", request.ServiceSource, jdModelInfo.Flavor, jdModelInfo.ServiceName)
				modelQuery := new(types.Model)
				modelQuery.ModelName = jdModelInfo.Name
				modelQuery.ProviderName = providerName
				canSelect := true
				err := ds.Get(context.Background(), modelQuery)
				if err != nil {
					canSelect = false
				}
				if modelQuery.Status != "downloaded" {
					canSelect = false
				}
				providerServiceDefaultInfo := schedule.GetProviderServiceDefaultInfo(jdModelInfo.Flavor, jdModelInfo.ServiceName)
				authFields := []string{""}
				if providerServiceDefaultInfo.AuthType == types.AuthTypeToken {
					authFields = []string{"secret_id", "secret_key"}
				} else if providerServiceDefaultInfo.AuthType == types.AuthTypeApiKey {
					authFields = []string{"api_key"}
				}
				modelData := dto.RecommendModelData{
					Id:              jdModelInfo.Id,
					Name:            jdModelInfo.Name,
					Avatar:          jdModelInfo.Avatar,
					Desc:            jdModelInfo.Description,
					Service:         jdModelInfo.ServiceName,
					ApiFlavor:       jdModelInfo.ApiFlavor,
					Flavor:          jdModelInfo.Flavor,
					AuthType:        providerServiceDefaultInfo.AuthType,
					AuthFields:      authFields,
					AuthApplyUrl:    providerServiceDefaultInfo.AuthApplyUrl,
					ServiceProvider: fmt.Sprintf("%s_%s_%s", jdModelInfo.ServiceSource, jdModelInfo.Flavor, jdModelInfo.ServiceName),
					CanSelect:       canSelect,
					IsRecommended:   false,
					Source:          jdModelInfo.ServiceSource,
					InputLength:     jdModelInfo.InputLength,
					OutputLength:    jdModelInfo.OutputLength,
					Class:           jdModelInfo.Class,
					OllamaId:        jdModelInfo.OllamaId,
					Size:            jdModelInfo.Size,
					ParamsSize:      jdModelInfo.ParamSize,
					Think:           jdModelInfo.Think,
					ThinkSwitch:     jdModelInfo.ThinkSwitch,
					Tools:           jdModelInfo.Tools,
					CreatedAt:       jdModelInfo.CreatedAt,
				}
				resultList = append(resultList, modelData)
			}

			// 数据处理, 如果是我的模型数据，则进行数据过滤-> 使用canSelect过滤
			if request.Mine {
				myModelFilter(&resultList)
			}

			dataStart := (page - 1) * pageSize
			dataEnd := page * pageSize
			if dataEnd > len(resultList) {
				dataEnd = len(resultList)
			}

			totalCount := len(resultList)
			// 当前页数据切片
			resultList = resultList[dataStart:dataEnd]
			resData.Total = totalCount
			if totalCount%pageSize == 0 {
				resData.TotalPage = totalCount / pageSize
			} else {
				resData.TotalPage = totalCount/pageSize + 1
			}

			if resData.TotalPage == 0 {
				resData.TotalPage = 1
			}
		}
	}
	resData.Data = resultList

	// If in model marketplace, prioritize the recommended models in the sorting.
	if !request.Mine {
		sort.Slice(resultList, func(i, j int) bool {
			a := resultList[i]
			b := resultList[j]

			if a.IsRecommended && !b.IsRecommended {
				return true
			}
			if !a.IsRecommended && b.IsRecommended {
				return false
			}
			return false
		})
	}

	// If the models are download, sort them only by the download time in chronological order.
	if request.Mine {
		sort.Slice(resultList, func(i, j int) bool {
			return resultList[i].CreatedAt.After(resultList[j].CreatedAt)
		})
	}

	return &dto.GetSupportModelResponse{
		*bcode.ModelCode,
		resData,
	}, nil
}

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
			logger.LogicLogger.Error("[Model] windows version < 10")
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
