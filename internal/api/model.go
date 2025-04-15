package api

import (
	"aipc/byze/internal/api/dto"
	"aipc/byze/internal/server"
	"aipc/byze/internal/types"
	"aipc/byze/internal/utils/bcode"
	"aipc/byze/internal/utils/client"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (t *ByzeCoreServer) CreateModel(c *gin.Context) {
	request := new(dto.CreateModelRequest)
	if err := c.ShouldBindJSON(request); err != nil {
		if !errors.Is(err, io.EOF) {
			bcode.ReturnError(c, bcode.ErrModelBadRequest)
			return
		}
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.Model.CreateModel(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *ByzeCoreServer) DeleteModel(c *gin.Context) {
	request := new(dto.DeleteModelRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.Model.DeleteModel(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *ByzeCoreServer) CreateModelStream(c *gin.Context) {
	request := new(dto.CreateModelRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.NotFound(w, c.Request)
		return
	}

	dataCh, errCh := server.CreateModelStream(ctx, *request)

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				// 数据通道关闭，发送结束标记
				//fmt.Fprintf(w, "data: [DONE]\n\n")
				fmt.Fprintf(w, "\n[DONE]\n\n")
				flusher.Flush()
				client.ModelClientMap[request.ModelName] = nil
				return
			}

			// 解析Ollama响应
			var resp types.ProgressResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				log.Printf("Error unmarshaling response: %v", err)
				continue
			}

			// 获取响应文本
			// 使用SSE格式发送到前端
			//fmt.Fprintf(w, "data: %s\n\n", response)
			if resp.Completed > 0 || resp.Status == "success" {
				fmt.Fprintf(w, string(data)+"\n")
				flusher.Flush()
			}

		case err, ok := <-errCh:
			if !ok {
				return
			}
			log.Printf("Error: %v", err)
			// 发送错误信息到前端
			fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
			flusher.Flush()
			client.ModelClientMap[request.ModelName] = nil
			return

		case <-ctx.Done():
			fmt.Fprintf(w, "event: timeout\ndata: request timeout\n\n")
			flusher.Flush()
			client.ModelClientMap[request.ModelName] = nil
			return
		}
	}

}

func (t *ByzeCoreServer) ModelStreamStop(c *gin.Context) {

}

func (t *ByzeCoreServer) GetModels(c *gin.Context) {
	request := new(dto.GetModelsRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.Model.GetModels(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *ByzeCoreServer) CancelModelStream(c *gin.Context) {
	request := new(dto.ModelStreamCancelRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}
	ctx := c.Request.Context()
	data, err := server.ModelStreamCancel(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (t *ByzeCoreServer) GetRecommendModels(c *gin.Context) {
	data, err := server.GetRecommendModel()
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (t *ByzeCoreServer) GetModelList(c *gin.Context) {
	request := new(dto.GetModelListRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}
	data, err := server.GetSupportModelList(request.ServiceSource, request.Flavor)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}
