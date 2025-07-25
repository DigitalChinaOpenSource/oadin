package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"oadin/internal/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/provider"
	"oadin/internal/provider/template"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"oadin/internal/utils"
	"oadin/internal/utils/bcode"
	"oadin/internal/utils/client"
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

	// Call engin to delete model.
	if m.Status == "downloaded" {
		modelEngine := provider.GetModelEngine(sp.Flavor)
		runningModelList, err := modelEngine.GetRunModels(ctx)
		if err != nil {
			return nil, bcode.ErrEngineDeleteModel
		}
		for _, runningModel := range runningModelList.Models {
			if runningModel.Name == m.ModelName {
				return nil, bcode.ErrModelIsRunning
			}
		}
		deleteReq := &types.DeleteRequest{
			Model:          request.ModelName,
			OllamaRegistry: m.OllamaRegistry,
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

	// Sort the response data by CreatedAt in descending order (newest first).
	if len(respData) > 1 {
		sort.Slice(respData, func(i, j int) bool {
			return respData[i].CreatedAt.After(respData[j].CreatedAt)
		})
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
	m := new(types.Model)
	m.ModelName = strings.ToLower(request.ModelName)
	m.ProviderName = sp.ProviderName
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
	providerEngine := provider.GetModelEngine(sp.Flavor)
	steam := true
	req := types.PullModelRequest{
		Model:  modelName,
		Stream: &steam,
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
					// 数据通道关闭，发送结束标记
					// fmt.Fprintf(w, "event: end\ndata: [DONE]\n\n")
					// fmt.Fprintf(w, "\n[DONE]\n\n")
					// flusher.Flush()
					if data == nil {
						client.ModelClientMap[strings.ToLower(request.ModelName)] = nil
						return
					}
				}

				// 解析Ollama响应
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

				// 获取响应文本
				// 使用SSE格式发送到前端
				// fmt.Fprintf(w, "data: %s\n\n", response)
				if resp.Completed > 0 || resp.Status == "success" {
					if resp.Status == "success" {
						// 增加处理逻辑, 如果是私仓下载模型, 通过cp复制一个无私仓前缀的模型镜像
						if req.Insecure {
							// 处理私仓模型下载完成后的逻辑
							var copyModelRequest types.CopyModelRequest
							copyModelRequest.Source = extractRegistryModelName(req.Model)
							copyModelRequest.Destination = ExtractModelName(req.Model)
							// 通过api复制模型
							// Copying model from private registry to local model store: {smartvision-registry.dcclouds.com/library/qwen3:0.6b qwen3:0.6b}
							fmt.Println("Copying model from private registry to local model store:", copyModelRequest)
							err := providerEngine.CopyModel(ctx, &copyModelRequest)
							if err != nil {
								newErrorCh <- err
								return
							}
							m.OllamaRegistry = ExtractRegistryName(copyModelRequest.Source)
						}

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
				client.ModelClientMap[strings.ToLower(request.ModelName)] = nil
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
	// 查service是否有模型，没有是否自动将第一个设为默认模型（待确认）
	err := ds.Put(newCtx, m)
	if err != nil {
		slog.Error("[Pull model] Update model error:", err.Error())
		return
	}
	// service model list
	// 如果有默认模型，仅校验默认模型即可
	// 如果没有，则需依次校验每个模型，直到有可用模型

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

func GetSupportModelList(ctx context.Context, request *dto.GetSupportModelRequest) (*dto.GetSupportModelResponse, error) {
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
		// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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
			// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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

			// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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
	if request.ServiceName != "" {
		queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
			Key:   "service_name",
			Query: request.ServiceName,
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
		// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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
			// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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
			if smartvisionErr != nil && jdsErr != nil {
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

			// 数据处理, 如果是我的模型数据，则进行数据过滤 -> 使用canSelect过滤
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

// 需要提取模型名
func extractRegistryModelName(fullName string) string {
	// 去除https:// 或 http:// 前缀
	fullName = strings.TrimPrefix(fullName, "https://")
	fullName = strings.TrimPrefix(fullName, "http://")
	return fullName
}

// 从smartvision-registry.dcclouds.com/library/qwen3:0.6b提取到smartvision-registry.dcclouds.com
func ExtractRegistryName(fullName string) string {
	// 找到第一个"/"的位置
	idx := strings.Index(fullName, "/")
	if idx == -1 {
		// 没有"/"，直接返回原始字符串
		return fullName
	}
	return fullName[:idx]
}

// ExtractModelName 提取模型名
func ExtractModelName(fullName string) string {
	// 找到最后一个"/"的位置
	idx := strings.LastIndex(fullName, "/")
	if idx == -1 {
		// 没有"/"，直接返回原始字符串
		return fullName
	}
	return fullName[idx+1:]
}
