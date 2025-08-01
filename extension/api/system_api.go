package api

import (
	"net/http"

	"oadin/extension/api/dto"
	"oadin/extension/rpc"
	"oadin/extension/server"
	"oadin/extension/utils/bcode"
	"oadin/internal/logger"

	"github.com/gin-gonic/gin"
)

type SystemlApi struct {
	System server.System
}

func NewSystemlApi() *SystemlApi {
	return &SystemlApi{
		System: server.NewSystemImpl(),
	}
}

func (e *SystemlApi) InjectRoutes(api *gin.RouterGroup) {
	api.GET("/about", e.About)
	api.GET("/information", e.SystemSettings)
	api.PUT("/registry", e.ModifyRepositoryURL)
	api.PUT("/proxy", e.SetProxy)
	api.PUT("/proxy/switch", e.ProxySwitch)
}

// About 获取关于信息
func (e *SystemlApi) About(c *gin.Context) {
	res, err := rpc.About()
	if err != nil {
		logger.ApiLogger.Error("Failed to get about information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get about information"})
	}
	c.JSON(http.StatusOK, res)
}

// ModifyRepositoryURL 修改仓库地址
func (e *SystemlApi) ModifyRepositoryURL(c *gin.Context) {
	var body struct {
		Url string `json:"url"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		bcode.ReturnError(c, err)
		return
	}
	err := e.System.ModifyRegistry(c.Request.Context(), body.Url)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "repository has modified successfully"})
}

// SetProxy 设置代理地址
func (e *SystemlApi) SetProxy(c *gin.Context) {
	var req dto.ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ApiLogger.Error("Invalid request body", "error", err)
		bcode.ReturnError(c, err)
		return
	}
	err := e.System.SetProxy(c.Request.Context(), req)
	if err != nil {
		logger.ApiLogger.Error("Failed to set proxy", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理设置成功"})
}

// ProxySwitch 代理启停
func (e *SystemlApi) ProxySwitch(c *gin.Context) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		logger.ApiLogger.Error("Invalid request body", "error", err)
		bcode.ReturnError(c, err)
		return
	}
	err := e.System.SwitchProxy(c.Request.Context(), body.Enabled)
	if err != nil {
		logger.ApiLogger.Error("Failed to switch proxy", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "代理切换成功"})
}

// SystemSettings 获取系统设置
func (e *SystemlApi) SystemSettings(c *gin.Context) {
	res, err := e.System.GetSystemSettings(c.Request.Context())
	if err != nil {
		logger.ApiLogger.Error("Failed to submit feedback", "error", err)
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
