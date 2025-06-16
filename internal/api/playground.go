package api

import (
	"net/http"

	"byze/internal/api/dto"
	"byze/internal/types"
	"github.com/gin-gonic/gin"
)

// 创建会话
func (t *ByzeCoreServer) CreateSession(c *gin.Context) {
	var req dto.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.Playground.CreateSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 获取会话列表
func (t *ByzeCoreServer) GetSessions(c *gin.Context) {
	resp, err := t.Playground.GetSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 发送消息
func (t *ByzeCoreServer) SendMessage(c *gin.Context) {
	var req dto.SendMessageRequest
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
	resp, err := t.Playground.SendMessage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 获取消息
func (t *ByzeCoreServer) GetMessages(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	req := &dto.GetMessagesRequest{
		SessionId: sessionID,
	}

	resp, err := t.Playground.GetMessages(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 删除会话
func (t *ByzeCoreServer) DeleteSession(c *gin.Context) {
	var req dto.DeleteSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.Playground.DeleteSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// 切换会话模型
func (t *ByzeCoreServer) ChangeSessionModel(c *gin.Context) {
	var req dto.ChangeSessionModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := t.Playground.ChangeSessionModel(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
