package api

import (
	"net/http"
	"runtime"

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
	api.GET("/:service/monitor", e.ServiceMonitor)
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

// ServiceMonitor 通用服务监控方法
func (e *SystemlApi) ServiceMonitor(c *gin.Context) {
	serviceName := c.Param("service")

	// 验证服务名称是否有效
	validServices := map[string]bool{
		"ollama":   true,
		"openvino": true,
	}

	if !validServices[serviceName] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid service name",
			"message": "无效的服务名称",
		})
		return
	}

	e.serviceMonitor(c, serviceName)
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

// serviceMonitor 通用的服务监控方法
func (e *SystemlApi) serviceMonitor(c *gin.Context, serviceName string) {
	platform := runtime.GOOS

	// 根据平台调用对应的处理方法
	switch platform {
	case "darwin":
		e.handleDarwinMonitor(c, serviceName)
	case "windows":
		e.handleWindowsMonitor(c, serviceName)
	case "linux":
		e.handleLinuxMonitor(c, serviceName)
	}
}

// handleDarwinMonitor 处理苹果系统监控
func (e *SystemlApi) handleDarwinMonitor(c *gin.Context, serviceName string) {
	// 获取苹果M系统芯片下，serviceName 应用的 cpu使用量，总量，占用比

	cpuUsed, cpuTotal, _ := e.getDarwinCPUUsage(serviceName)
	gpuUsed, _, _ := e.getDarwinGPUUsage(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"cpu_percentage": cpuUsed,     // 当前应用CPU使用率（%）- 可准确获取
		"cpu_cores":      cpuTotal,    // 系统CPU核心数（个）- 可准确获取
		"gpu_available":  gpuUsed > 0, // GPU是否可用（布尔值）- 可检测
	})
}

// getDarwinCPUUsage 获取Darwin系统下指定服务的CPU使用情况
func (e *SystemlApi) getDarwinCPUUsage(serviceName string) (float64, float64, float64) {
	return 0.0, 0.0, 0.0
}

// getDarwinGPUUsage 获取Darwin系统下指定服务的GPU使用情况
func (e *SystemlApi) getDarwinGPUUsage(serviceName string) (float64, float64, float64) {
	// 在macOS上，可以使用ioreg命令或者其他方式获取GPU使用情况
	// 这里先返回默认值，后续可以根据需要实现具体逻辑

	// TODO: 实现macOS下GPU使用率监控
	// 可以使用 ioreg -r -d 1 -w 0 -c IOPCIDevice | grep -i gpu
	// 或者使用 powermetrics 等工具

	return 0.0, 0.0, 0.0
}

// handleWindowsMonitor 处理Windows系统监控
func (e *SystemlApi) handleWindowsMonitor(c *gin.Context, serviceName string) {
	// Windows系统下可以通过以下方式获取进程信息：
	// 1. tasklist /fo csv | findstr "serviceName"
	// 2. wmic process where "name='serviceName.exe'" get ProcessId,PageFileUsage,WorkingSetSize
	// 3. PowerShell: Get-Process -Name serviceName

	cpuUsed, cpuTotal, _ := e.getWindowsCPUUsage(serviceName)
	gpuUsed, _, _ := e.getWindowsGPUUsage(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"cpu_percentage": cpuUsed,     // 当前应用CPU使用率（%）- 可通过tasklist或wmic获取
		"cpu_cores":      cpuTotal,    // 系统CPU核心数（个）- 可通过wmic获取
		"gpu_available":  gpuUsed > 0, // GPU是否可用（布尔值）- 可通过nvidia-smi或wmic检测
	})
}

// getWindowsCPUUsage 获取Windows系统下指定服务的CPU使用情况
func (e *SystemlApi) getWindowsCPUUsage(serviceName string) (float64, float64, float64) {
	return 0.0, 0.0, 0.0
}

// getWindowsGPUUsage 获取Windows系统下指定服务的GPU使用情况
func (e *SystemlApi) getWindowsGPUUsage(serviceName string) (float64, float64, float64) {
	return 0.0, 1.0, 0.0
}

// handleLinuxMonitor 处理Linux系统监控
func (e *SystemlApi) handleLinuxMonitor(c *gin.Context, serviceName string) {

	c.JSON(http.StatusOK, gin.H{
		"cpu_used":       0.0,   // 当前应用CPU使用率（%）
		"cpu_total":      100.0, // 系统CPU总容量（%）
		"cpu_percentage": 0.0,   // 当前应用CPU占用比（%）
		"gpu_used":       0.0,   // 当前应用GPU使用率（%）
		"gpu_total":      100.0, // 系统GPU总容量（%）
		"gpu_percentage": 0.0,   // 当前应用GPU占用比（%）
	})
}
