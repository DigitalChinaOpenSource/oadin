package api

import (
	"fmt"
	"strings"
	"net/http"

	"github.com/gin-gonic/gin"
	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/extension/utils/bcode"
	"oadin/internal/provider"
	"oadin/internal/types"
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
	api.POST("/Download/streamEngine", e.DownloadStreamEngine)
	api.POST("/Download/streamModel", e.DownloadStreamModel)
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

	// 有就是不同批次返回，没有就是同一批出来的
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

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

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				select {
				case err, _ := <-errCh:
					if err != nil {
						fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"%v\"}\n\n", err)
						flusher.Flush()
						return
					}
				}
				// 数据通道关闭，发送结束标记
				if data == nil {
					fmt.Fprintf(w, "data: {\"status\": \"success\"}\n\n")
					return
				}
			}

			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

		case err, _ := <-errCh:
			if err != nil {
				// 发送错误信息到前端
				fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"%v\"}\n\n", err)
				flusher.Flush()
				return
			}

		case <-ctx.Done():
			fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"timeout\"}\n\n")
			flusher.Flush()
			return
		}
	}
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
	if !strings.Contains("ollama,openvino,llamacpp", request.EngineName) {
		dto.ValidFailure(c, fmt.Sprintf("invalid engine name: %s", request.EngineName))
		return
	}
	modelEngine := provider.GetModelEngine(request.EngineName)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

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

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				select {
				case err, _ := <-errCh:
					if err != nil {
						fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"%v\"}\n\n", err)
						flusher.Flush()
						return
					}
				}
				// 数据通道关闭，发送结束标记
				if data == nil {
					fmt.Fprintf(w, "data: {\"status\": \"success\"}\n\n")
					return
				}
			}

			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

		case err, _ := <-errCh:
			if err != nil {
				// 发送错误信息到前端
				fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"%v\"}\n\n", err)
				flusher.Flush()
				return
			}

		case <-ctx.Done():
			fmt.Fprintf(w, "data: {\"status\": \"error\", \"data\":\"timeout\"}\n\n")
			flusher.Flush()
			return
		}
	}
}
