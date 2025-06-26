package api

import (
	"byze/internal/api/dto"
	"byze/internal/rpc"
	"byze/internal/utils/bcode"
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

// ModifyRepositoryURL 修改仓库地址
func (t *ByzeCoreServer) ModifyRepositoryURL(c *gin.Context) {
	var body struct {
		Url string `json:"url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bcode.ReturnError(c, err)
		return
	}
	err := t.System.ModifyRegistry(c.Request.Context(), body.Url)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "repository has modified successfully"})
}

// SetProxy 设置代理地址
func (t *ByzeCoreServer) SetProxy(c *gin.Context) {
	var req dto.ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Invalid request body", "error", err)
		bcode.ReturnError(c, err)
		return
	}
	err := t.System.SetProxy(c.Request.Context(), req)
	if err != nil {
		slog.Error("Failed to set proxy", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理设置成功"})
}

// ProxySwitch 代理启停
func (t *ByzeCoreServer) ProxySwitch(c *gin.Context) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		slog.Error("Invalid request body", "error", err)
		bcode.ReturnError(c, err)
		return
	}
	err := t.System.SwitchProxy(c.Request.Context(), body.Enabled)
	if err != nil {
		slog.Error("Failed to switch proxy", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理切换成功"})

}

// SystemSettings 获取系统设置
func (t *ByzeCoreServer) SystemSettings(c *gin.Context) {

	res, err := t.System.GetSystemSettings(c.Request.Context())
	if err != nil {
		slog.Error("Failed to submit feedback", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)

}
