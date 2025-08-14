package api

import (
	"net/http"

	"oadin/extension/api/dto"
	"oadin/extension/rpc"
	"oadin/extension/server"
	"oadin/extension/utils/bcode"
	"oadin/extension/utils/hardware"
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

	// 硬件监控相关路由
	api.GET("/cpu", e.GetCPUInfo)
	api.GET("/memory", e.GetMemoryInfo)
	api.GET("/gpu", e.GetGPUInfo)
	api.GET("/hardware", e.GetHardwareInfo)
	api.GET("/hardware/full", e.GetFullHardwareInfo)

	// ollama & openvino 监控
	api.GET("/ollama/monitor", e.OllamaMonitor)
	api.GET("/openvino/monitor", e.OpenVinoMonitor)
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

// GetMemoryInfo 获取系统内存信息
func (e *SystemlApi) GetMemoryInfo(c *gin.Context) {
	memInfo, err := hardware.GetMemoryInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get memory information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取内存信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    memInfo,
	})
}

// GetGPUInfo 获取系统GPU信息
func (e *SystemlApi) GetGPUInfo(c *gin.Context) {
	gpuInfo, err := hardware.GetGPUInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get GPU information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取GPU信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gpuInfo,
		"count":   len(gpuInfo),
	})
}

// GetHardwareInfo 获取完整的系统硬件信息
func (e *SystemlApi) GetHardwareInfo(c *gin.Context) {
	// 获取内存信息
	memInfo, err := hardware.GetMemoryInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get memory information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取内存信息失败",
			"message": err.Error(),
		})
		return
	}

	// 获取GPU信息
	gpuInfo, err := hardware.GetGPUInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get GPU information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取GPU信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"memory": memInfo,
			"gpus":   gpuInfo,
		},
	})
}

// GetCPUInfo 获取系统CPU信息
func (e *SystemlApi) GetCPUInfo(c *gin.Context) {
	cpuInfo, err := hardware.GetCPUInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get CPU information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取CPU信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    cpuInfo,
	})
}

// OllamaMonitor 获取Ollama监控信息
func (e *SystemlApi) OllamaMonitor(c *gin.Context) {
	// 获取Ollama监控信息
	monitorInfo, err := hardware.GetOllamaMonitor()
	if err != nil {
		logger.ApiLogger.Error("Failed to get Ollama monitor information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取Ollama监控信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitorInfo,
	})
}

// OpenVinoMonitor 获取OpenVino监控信息
func (e *SystemlApi) OpenVinoMonitor(c *gin.Context) {
	// 获取OpenVino监控信息
	monitorInfo, err := hardware.GetOpenVinoMonitor()
	if err != nil {
		logger.ApiLogger.Error("Failed to get OpenVino monitor information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取OpenVino监控信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    monitorInfo,
	})
}

// GetFullHardwareInfo 获取完整的系统硬件信息（包含CPU）
func (e *SystemlApi) GetFullHardwareInfo(c *gin.Context) {
	hardwareInfo, err := hardware.GetSystemHardwareInfo()
	if err != nil {
		logger.ApiLogger.Error("Failed to get full hardware information", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取完整硬件信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    hardwareInfo,
	})
}
