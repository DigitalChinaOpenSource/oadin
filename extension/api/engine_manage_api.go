package api

import (
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/internal/utils"
	"oadin/extension/utils/bcode"
	"oadin/internal/provider"
	"oadin/internal/types"
	dto2 "oadin/internal/api/dto"
	"oadin/internal/logger"
)

type EngineApi struct {
	EngineManageService server.EngineManageService
}

func NewEngineApi() *EngineApi {
	return &EngineApi{
		EngineManageService: server.NewEngineManageService(),
	}
}

func (e *EngineApi) InjectRoutes(api *gin.RouterGroup) {
	api.GET("/exist", e.exist)
	api.POST("/install", e.install)
	api.POST("/download/streamEngine", e.DownloadStreamEngine)
	api.GET("/download/checkMemoryConfig", e.CheckMemoryConfig)
	api.POST("/download/streamModel", e.DownloadStreamModel)
	api.POST("/download/checkDist", e.DownloadCheckDist)
	api.POST("/download/checkModel", e.DownloadCheckModel)
}

// exist 检查引擎是否存在
func (e *EngineApi) exist(c *gin.Context) {
	req := dto.EngineManageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidFailure(c, err.Error())
		return

	}
	// 检查引擎的安装状态
	exist := e.EngineManageService.Exist(c, req)

	dto.Success(c, exist)
}

// 执行引擎下载逻辑
func (e *EngineApi) install(c *gin.Context) {
	req := dto.EngineManageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}

	// 执行安装逻辑
	err := e.EngineManageService.Install(c, req)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	dto.Success(c, "引擎安装成功")
}

// 根据引擎名称下载引擎，流式返回下载进度
func (e *EngineApi) DownloadStreamEngine(c *gin.Context) {
	request := dto.EngineDownloadRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}
	if !strings.Contains("ollama,openvino,llamacpp", request.EngineName) {
		dto.ValidFailure(c, fmt.Sprintf("invalid engine name: %s", request.EngineName))
		return
	}

	if request.Stream {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
	}

	ctx := c.Request.Context()

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.NotFound(w, c.Request)
		return
	}

	modelEngine := provider.GetModelEngine(request.EngineName)

	res := dto.DownloadResponse{
		Status: "success",
	}

	engineConfig := modelEngine.GetConfig()
	execPath := filepath.Join(engineConfig.ExecPath, engineConfig.ExecFile)
	if _, err := os.Stat(execPath); err == nil {
		err = modelEngine.HealthCheck()
		if err != nil {
			err = modelEngine.InitEnv()
			if err != nil {
				res.Status = "error"
				logger.EngineLogger.Error("DownloadStreamEngine InitEnv error 1: ", err)
			} else {
				err = modelEngine.StartEngine(types.EngineStartModeDaemon)
				if err != nil {
					res.Status = "error"
					logger.EngineLogger.Error("DownloadStreamEngine StartEngine error 1: ", err)
				}
			}
		}

		if request.Stream {
			dataBytes, _ := json.Marshal(res)
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		} else {
			c.JSON(http.StatusOK, res)
		}
		return
	}

	dataCh := make(chan []byte, 100)
	errCh := make(chan error, 1)
	go modelEngine.InstallEngineStream(ctx, dataCh, errCh)

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				// 数据通道关闭，发送结束标记
				if _, err := os.Stat(execPath); err == nil {
					err = modelEngine.HealthCheck()
					if err != nil {
						err = modelEngine.InitEnv()
						if err != nil {
							res.Status = "error"
							logger.EngineLogger.Error("DownloadStreamEngine InitEnv error 2: ", err)
						} else {
							err = modelEngine.StartEngine(types.EngineStartModeDaemon)
							if err != nil {
								res.Status = "error"
								logger.EngineLogger.Error("DownloadStreamEngine StartEngine error 2: ", err)
							}
						}
					}
				} else {
					res.Status = "error"
					logger.EngineLogger.Error("DownloadStreamEngine exec file not found after install")
				}

				if request.Stream {
					dataBytes, _ := json.Marshal(res)
					fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
					flusher.Flush()
				} else {
					c.JSON(http.StatusOK, res)
				}
				return
			}

			if request.Stream && data != nil {
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
			}
		case err, _ := <-errCh:
			if err != nil {
				logger.EngineLogger.Error("DownloadStreamEngine", err)
				res.Status = "error"
				res.Data = err.Error()
				if request.Stream {
					dataBytes, _ := json.Marshal(res)
					fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
					flusher.Flush()
				} else {
					c.JSON(http.StatusInternalServerError, res)
				}
			}
		case <-ctx.Done():
			logger.EngineLogger.Error("DownloadStreamEngine context done")
			res.Status = "error"
			res.Data = "timeout"
			if request.Stream {
				dataBytes, _ := json.Marshal(res)
				fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
				flusher.Flush()
			} else {
				c.JSON(http.StatusInternalServerError, res)
			}
			return
		}
	}
}

func (e *EngineApi) CheckMemoryConfig(c *gin.Context) {
	// 检查引擎配置
	memoryInfo, err := utils.GetMemoryInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, memoryInfo)
}

func (e *EngineApi) DownloadStreamModel(c *gin.Context) {
	request := dto.ModelDownloadRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}
	if !strings.Contains("ollama,openvino,llamacpp", request.EngineName) {
		dto.ValidFailure(c, fmt.Sprintf("invalid engine name: %s", request.EngineName))
		return
	}

	if request.Stream {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
	}

	ctx := c.Request.Context()

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.NotFound(w, c.Request)
		return
	}

	logger.EngineLogger.Info("DownloadStreamModel request: ", request)
	modelEngine := provider.GetModelEngine(request.EngineName)

	res := dto.DownloadResponse{
		Status: "success",
	}

	modelList, _ := modelEngine.ListModels(c)
	modelFileExist := false
	for _, model := range modelList.Models {
		if model.Name == request.ModelName || model.Model == request.ModelName {
			modelFileExist = true
			break
		}
	}
	if modelFileExist {
		logger.EngineLogger.Info("Model already downloaded: ", request.ModelName)
		err := e.EngineManageService.CheckLocalModelExist(ctx, request)
		if err != nil {
			err = e.EngineManageService.InsertLocalModel(ctx, request)
			if err != nil {
				logger.EngineLogger.Error("InsertLocalModel error: ", err)
				res.Status = "error"
				res.Data = err.Error()
			}
		}

		if request.Stream {
			dataBytes, _ := json.Marshal(res)
			fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
			flusher.Flush()
		} else {
			c.JSON(http.StatusOK, res)
		}
		return
	}

	req := types.PullModelRequest{
		Model:    request.ModelName,
		ModelType: request.ModelType,
	}
	// dataCh, errCh := t.Model.CreateModelStream(ctx, request)
	dataCh, errCh := modelEngine.PullModelStream(ctx, &req)

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				// 数据通道关闭，发送结束标记
				if data == nil {
					// 更新service表和model表
					newReq := &dto2.CreateAIGCServiceRequest{
						ServiceName: request.ModelType,
						ServiceSource: "local",
						ApiFlavor: request.EngineName,
						ModelName: request.ModelName,
					}
					err := e.EngineManageService.CreateAIGCServiceSync(ctx, newReq)
					if err != nil && err.Error() != "provider model already exist" {
						logger.EngineLogger.Error("CreateAIGCServiceSync error: ", err)
						res.Status = "error"
						res.Data = err.Error()
						if request.Stream {
							dataBytes, _ := json.Marshal(res)
							fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
							flusher.Flush()
						} else {
							c.JSON(http.StatusInternalServerError, res)
						}
						return
					}

					if request.Stream {
						dataBytes, _ := json.Marshal(res)
						fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
						flusher.Flush()
					} else {
						c.JSON(http.StatusOK, res)
					}
					return
				}
			}

			if request.Stream && data != nil {
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
			}
		case err, _ := <-errCh:
			if err != nil {
				logger.EngineLogger.Error("DownloadStreamModel err: ", err)
				res.Status = "error"
				res.Data = err.Error()
				if request.Stream {
					dataBytes, _ := json.Marshal(res)
					fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
					flusher.Flush()
				} else {
					c.JSON(http.StatusInternalServerError, res)
				}
				return
			}

		case <-ctx.Done():
			res.Status = "error"
			res.Data = "timeout"
			logger.EngineLogger.Error("DownloadStreamModel timeout")
			if request.Stream {
				dataBytes, _ := json.Marshal(res)
				fmt.Fprintf(w, "data: %s\n\n", string(dataBytes))
				flusher.Flush()
			} else {
				c.JSON(http.StatusInternalServerError, res)
			}
			return
		}
	}
}

func (e *EngineApi) DownloadCheckDist(c *gin.Context) {
	request := dto.DownloadCheckDistRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}
	if !strings.Contains("ollama,openvino,llamacpp", request.EngineName) {
		dto.ValidFailure(c, fmt.Sprintf("invalid engine name: %s", request.EngineName))
		return
	}

	modelEngine := provider.GetModelEngine(request.EngineName)
	res := dto.DownloadResponse{
		Status: "success",
	}


	err := modelEngine.HealthCheck()
	if err != nil {
		res.Status = "error"
		res.Data = "engine error"
		c.JSON(http.StatusOK, res)
		return
	}

	modelList, err := modelEngine.ListModels(c)
	if err != nil {
		res.Status = "error"
		res.Data = "list error"
		c.JSON(http.StatusOK, res)
		return
	}

	if modelList == nil || len(modelList.Models) == 0 {
		res.Status = "error"
		res.Data = "no model error"
		c.JSON(http.StatusOK, res)
		return
	}

	// 检查引擎配置
	var models []string
	memoryInfo, err := utils.GetMemoryInfo()
	if err != nil {
		res.Status = "error"
		res.Data = "size error"
		c.JSON(http.StatusOK, res)
		return
	}

	if memoryInfo.Size > 32 {
		models = []string{"qwen3:14b", "bge-m3:567m"}
	} else if memoryInfo.Size > 16 && memoryInfo.Size <= 32 {
		models = []string{"qwen3:8b", "quentinz/bge-large-zh-v1.5:f16"}
	} else {
		models = []string{"qwen3:1.7b", "quentinz/bge-large-zh-v1.5:f16"}
	}

	// 判断modelList.Models是否包含models的模型 如果缺少models的模型，则报错
	for _, requiredModel := range models {
		found := false
		for _, existingModel := range modelList.Models {
			if existingModel.Name == requiredModel || existingModel.Model == requiredModel {
				found = true
				break
			}
		}
		
		// 如果找不到必需的模型，设置错误状态
		if !found {
			res.Status = "error"
			res.Data = fmt.Sprintf("missing required model: %s", requiredModel)
			c.JSON(http.StatusOK, res)
			return
		}

		modelType := "chat"
		if requiredModel == "bge-m3:567m" || requiredModel == "quentinz/bge-large-zh-v1.5:f16" {
			modelType = "embed"
		}

		req := dto.ModelDownloadRequest{
			EngineName: request.EngineName,
			ModelName:  requiredModel,
			ModelType:  modelType,
		}

		err := e.EngineManageService.CheckLocalModelExist(c, req);
		if err != nil {
			res.Status = "error"
			res.Data = fmt.Sprintf("table model not found: %s", err.Error())
			c.JSON(http.StatusOK, res)
			return
		}
	}

	// 如果所有必需模型都存在，继续执行
	c.JSON(http.StatusOK, res)
}

// 支持了云端模型后，需要增加具体模型的判断，可以先下载embed模型，再下载chat模型
func (e *EngineApi) DownloadCheckModel(c *gin.Context) {
	request := dto.DownloadCheckModelRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}
	if !strings.Contains("ollama,openvino,llamacpp", request.EngineName) {
		dto.ValidFailure(c, fmt.Sprintf("invalid engine name: %s", request.EngineName))
		return
	}

	modelEngine := provider.GetModelEngine(request.EngineName)
	res := dto.DownloadResponse{
		Status: "success",
	}


	err := modelEngine.HealthCheck()
	if err != nil {
		res.Status = "error"
		res.Data = "engine error"
		c.JSON(http.StatusOK, res)
		return
	}

	modelList, err := modelEngine.ListModels(c)
	if err != nil {
		res.Status = "error"
		res.Data = "list error"
		c.JSON(http.StatusOK, res)
		return
	}

	if modelList == nil || len(modelList.Models) == 0 {
		res.Status = "error"
		res.Data = "no model error"
		c.JSON(http.StatusOK, res)
		return
	}

	memoryInfo, err := utils.GetMemoryInfo()
	if err != nil {
		res.Status = "error"
		res.Data = "size error"
		c.JSON(http.StatusOK, res)
		return
	}

	var requiredModel string
	if request.ModelType == "embed" {
		if memoryInfo.Size > 32 {
			requiredModel = "bge-m3:567m"
		} else if memoryInfo.Size > 16 && memoryInfo.Size <= 32 {
			requiredModel = "quentinz/bge-large-zh-v1.5:f16"
		} else {
			requiredModel = "quentinz/bge-large-zh-v1.5:f16"
		}
	}

	if request.ModelType == "chat" {
		if memoryInfo.Size > 32 {
			requiredModel = "qwen3:14b"
		} else if memoryInfo.Size > 16 && memoryInfo.Size <= 32 {
			requiredModel = "qwen3:8b"
		} else {
			requiredModel = "qwen3:1.7b"
		}
	}



	// 判断modelList.Models是否包含models的模型 如果缺少models的模型，则报错
	found := false
	for _, existingModel := range modelList.Models {
		if existingModel.Name == requiredModel || existingModel.Model == requiredModel {
			found = true
			break
		}
	}
	
	// 如果找不到必需的模型，设置错误状态
	if !found {
		res.Status = "error"
		res.Data = fmt.Sprintf("missing required model: %s", requiredModel)
		c.JSON(http.StatusOK, res)
		return
	}

	modelType := "chat"
	if requiredModel == "bge-m3:567m" || requiredModel == "quentinz/bge-large-zh-v1.5:f16" {
		modelType = "embed"
	}

	req := dto.ModelDownloadRequest{
		EngineName: request.EngineName,
		ModelName:  requiredModel,
		ModelType:  modelType,
	}

	err = e.EngineManageService.CheckLocalModelExist(c, req);
	if err != nil {
		res.Status = "error"
		res.Data = fmt.Sprintf("table model not found: %s", err.Error())
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)	
}

