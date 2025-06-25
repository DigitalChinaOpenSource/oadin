package api

import (
	"net/http"

	"byze/internal/api/dto"

	"github.com/gin-gonic/gin"
)

// 切换会话深度思考模式
func (t *ByzeCoreServer) ToggleSessionThinking(c *gin.Context) {
	var req dto.ToggleThinkingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.Playground.ToggleThinking(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *ByzeCoreServer) HandleToolCalls(c *gin.Context) {
	resp := t.Playground.HandleToolCalls(c.Request.Context(), "c9625fa2-9b79-4439-93c7-ae64a2eb4303", "adf21327-cdfe-4d44-8887-96d19bb4f352")
	c.JSON(http.StatusOK, resp)
}
