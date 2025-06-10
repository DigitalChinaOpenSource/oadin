package api

import (
	"encoding/json"
	"io"
	"net/http"

	"byze/internal/api/dto"
	"byze/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

// 发送消息并流式返回响应
func (h *PlaygroundHandler) SendMessageStream(c *gin.Context) {
	var req dto.SendStreamMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 开始流式处理
	respChan, errChan := h.playground.SendMessageStream(c.Request.Context(), &req)

	// 写入响应流
	c.Stream(func(w io.Writer) bool {
		select {
		case chunk, ok := <-respChan:
			if !ok {
				return false // 流结束
			}

			response := dto.StreamMessageResponse{
				Bcode: bcode.SuccessCode,
				Data: dto.MessageChunk{
					ID:         chunk.ID,
					SessionID:  req.SessionID,
					Content:    chunk.Content,
					IsComplete: chunk.IsComplete,
					Thoughts:   chunk.Thoughts,
					Type:       chunk.Type,
				},
			}

			// 序列化为JSON
			data, err := json.Marshal(response)
			if err != nil {
				return false // 序列化错误，结束流
			}

			// 写入数据
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n")) // 添加换行符分隔事件

			return !chunk.IsComplete // 如果是最后一块，则结束流

		case err, ok := <-errChan:
			if !ok || err == nil {
				return false
			}

			// 处理错误
			errResp := gin.H{"error": err.Error()}
			data, _ := json.Marshal(errResp)
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n"))

			return false
		}
	})
}
