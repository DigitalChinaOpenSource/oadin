package api

import (
	"fmt"
	"strings"
	"net/http"
	"encoding/json"
	"context"

	"github.com/gin-gonic/gin"
	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/internal/utils"
	"oadin/extension/utils/bcode"
	"oadin/internal/provider"
	"oadin/internal/types"
	"oadin/version"
	"oadin/config"
	dto2 "oadin/internal/api/dto"
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
	modelEngine := provider.GetModelEngine(request.EngineName)

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

	dataCh := make(chan []byte, 100)
	errCh := make(chan error, 1)
	go modelEngine.InstallEngineStream(ctx, dataCh, errCh)

	res := dto.DownloadResponse{
		Status: "success",
	}

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				// 数据通道关闭，发送结束标记
				err := modelEngine.HealthCheck()
				if err != nil {
					err = modelEngine.InitEnv()
					if err != nil {
						res.Status = "error"
					}
					err = modelEngine.StartEngine(types.EngineStartModeDaemon)
					if err != nil {
						res.Status = "error"
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

			if request.Stream && data != nil {
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				flusher.Flush()
			}
		case err, _ := <-errCh:
			if err != nil {
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

	modelEngine := provider.GetModelEngine(request.EngineName)

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

	req := types.PullModelRequest{
		Model:    request.ModelName,
		ModelType: request.ModelType,
	}
	// dataCh, errCh := t.Model.CreateModelStream(ctx, request)
	dataCh, errCh := modelEngine.PullModelStream(ctx, &req)

	res := dto.DownloadResponse{
		Status: "success",
	}

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				// 更新service表和model表
				newC := config.NewOADINClient()
				routerPath := fmt.Sprintf("/oadin/%s/service/install", version.OADINSpecVersion)

				newReq := &dto2.CreateAIGCServiceRequest{
					ServiceName: request.ModelType,
					ServiceSource: "local",
					ApiFlavor: request.EngineName,
					ModelName: request.ModelName,
				}
				newResp := bcode.Bcode{}
				err := newC.Client.Do(context.Background(), http.MethodPost, routerPath, newReq, &newResp)
				if err != nil {
					fmt.Printf("\rService install failed: %s", err.Error())
				}

				// 数据通道关闭，发送结束标记
				if data == nil {
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


