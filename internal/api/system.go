package api

import (
	"byze/internal/api/dto"
	"byze/internal/rpc"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

// About 获取关于信息
func (t *ByzeCoreServer) About(c *gin.Context) {
	res, err := rpc.About()
	if err != nil {
		slog.Error("Failed to get about information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get about information"})
	}
	c.JSON(http.StatusOK, res)

}

func (t *ByzeCoreServer) GetChangeList(c *gin.Context) {

}

// ModifyRepositoryURL 修改仓库地址
func (t *ByzeCoreServer) ModifyRepositoryURL(c *gin.Context) {
	var body struct {
		Url string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement the logic to modify the repository URL
	c.JSON(http.StatusOK, gin.H{"message": "repository has modified successfully"})
}

// SetProxy 设置代理地址
func (t *ByzeCoreServer) SetProxy(c *gin.Context) {
	var req dto.ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	// TODO: Implement the logic to set the proxy URL

	c.JSON(http.StatusOK, gin.H{"message": "Proxy set successfully"})
}

func (t *ByzeCoreServer) Feedback(c *gin.Context) {
	var req dto.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement the logic to handle feedback
	c.JSON(http.StatusOK, gin.H{"message": "Feedback set successfully"})
}
