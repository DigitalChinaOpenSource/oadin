package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"oadin/internal/api/dto"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

// 发送消息并流式返回响应
func (t *OadinCoreServer) SendMessageStream(c *gin.Context) {
	var req dto.SendStreamMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.McpIds) > 0 {
		req.Tools = make([]types.Tool, 0)
		for _, id := range req.McpIds {
			tools, err := t.MCP.ClientGetTools(c, id)
			if err != nil {
				continue
			}

			newTools := make([]types.Tool, 0, len(tools))
			for _, tool := range tools {
				newTools = append(newTools, types.Tool{Type: "function", Function: types.TypeFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema}})
			}
			req.Tools = append(req.Tools, newTools...)
		}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ctx := c.Request.Context()

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		fmt.Printf("[API] Flusher不支持\n")
		http.NotFound(w, c.Request)
		return
	}

	// 开始流式处理
	respChan, errChan := t.Playground.SendMessageStream(ctx, &req)

	for {
		select {
		case chunk, ok := <-respChan:
			if !ok {
				return // 流结束
			}

			if chunk.Type == "error" {
				response := dto.StreamMessageResponse{
					Bcode: bcode.ErrServer,
					Data: dto.MessageChunk{
						ID:        chunk.ID,
						SessionID: req.SessionID,
						Content:   chunk.Content,
						Type:      "error",
					},
				}
				data, err := json.Marshal(response)
				if err == nil {
					n, err := fmt.Fprintf(w, "data: %s\n\n", data)
					fmt.Printf("[API] 错误消息写入结果: 字节数=%d, 错误=%v\n", n, err)
					flusher.Flush()
				}
				return
			}

			response := dto.StreamMessageResponse{
				Bcode: bcode.SuccessCode,
				Data: dto.MessageChunk{
					ID:            chunk.ID,
					SessionID:     req.SessionID,
					Content:       chunk.Content,
					IsComplete:    chunk.IsComplete,
					Thoughts:      chunk.Thoughts,
					Type:          chunk.Type,
					ToolCalls:     chunk.ToolCalls,
					TotalDuration: chunk.TotalDuration,
					ToolGroupID:   chunk.ToolGroupID,
				},
			}
			data, err := json.Marshal(response)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if chunk.IsComplete {
				return
			}

		case err, ok := <-errChan:
			if !ok || err == nil {
				return
			}
			flusher.Flush()
			return
		}
	}
}

func (t *OadinCoreServer) GenSessionTitle(c *gin.Context) {
	var req struct {
		SessionID string `json:"sessionId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	err := t.Playground.UpdateSessionTitle(c.Request.Context(), req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "404", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": "success"})
}
