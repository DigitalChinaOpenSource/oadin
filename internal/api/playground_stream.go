package api

import (
	"encoding/json"
	"io"
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

	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 开始流式处理
	respChan, errChan := t.Playground.SendMessageStream(c.Request.Context(), &req)

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
					ID:            chunk.ID,
					SessionID:     req.SessionID,
					Content:       chunk.Content,
					IsComplete:    chunk.IsComplete,
					Thoughts:      chunk.Thoughts,
					Type:          chunk.Type,
					ToolCalls:     chunk.ToolCalls,     // 新增，支持Ollama工具调用
					TotalDuration: chunk.TotalDuration, // 新增总耗时
					ToolGroupID:   chunk.ToolGroupID,   // 新增工具组ID
				},
			} // 序列化为JSON
			data, err := json.Marshal(response)
			if err != nil {
				return false // 序列化错误，结束流
			}

			// 增加SSE格式的前缀
			c.Writer.Write([]byte("data: "))
			// 写入数据
			c.Writer.Write(data)
			c.Writer.Write([]byte("\n\n")) // 添加双换行符符合SSE格式

			// 立即刷新写入器，确保数据能及时发送到客户端
			c.Writer.Flush()

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
