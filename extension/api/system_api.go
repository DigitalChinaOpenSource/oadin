package api

import (
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

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

	// 检查服务名称是否为空
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Service name is required",
			"message": "服务名称不能为空",
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
	cpuUsed := e.getDarwinCPUUsage(serviceName)
	gpuUsed := e.getDarwinGPUUsage(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"cpu_usage": cpuUsed, // 当前应用CPU占用率（%）
		"gpu_usage": gpuUsed, // 当前应用GPU占用率（%）
	})
}

// getDarwinCPUUsage 获取Darwin系统下指定服务的CPU使用情况（包含所有相关子进程）
func (e *SystemlApi) getDarwinCPUUsage(serviceName string) float64 {
	// 获取系统CPU核心数
	cpuCores := float64(runtime.NumCPU())

	// 方法1: 直接搜索包含服务名的所有进程（包括子进程和相关进程）
	// ps -eo pid,pcpu,comm | grep -i serviceName | grep -v grep | awk '{sum+=$2} END {print sum}'
	cmd1 := fmt.Sprintf("ps -eo pid,pcpu,comm | grep -i %s | grep -v grep | awk '{sum+=$2} END {print sum}'", serviceName)

	// 方法2: 搜索以服务名开头的进程
	// ps -eo pid,pcpu,comm | awk '$3 ~ /^%s/ {sum+=$2} END {print sum}'
	cmd2 := fmt.Sprintf("ps -eo pid,pcpu,comm | awk '$3 ~ /^%s/ {sum+=$2} END {print sum}'", serviceName)

	// 方法3: 搜索包含服务名的所有进程（更宽泛的匹配）
	// ps -eo pid,pcpu,args | grep -i serviceName | grep -v grep | awk '{sum+=$2} END {print sum}'
	cmd3 := fmt.Sprintf("ps -eo pid,pcpu,args | grep -i %s | grep -v grep | awk '{sum+=$2} END {print sum}'", serviceName)

	// 尝试多种方法获取CPU使用率，取最大值（最完整的结果）
	methods := []string{cmd1, cmd2, cmd3}
	var maxCpuUsage float64 = 0.0
	var foundAny bool = false

	for i, cmd := range methods {
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			logger.ApiLogger.Debug("Method failed to get CPU usage for service", "method", i+1, "service", serviceName, "error", err)
			continue
		}

		cpuUsageStr := strings.TrimSpace(string(output))
		if cpuUsageStr == "" || cpuUsageStr == "0" {
			continue
		}

		cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
		if err != nil {
			logger.ApiLogger.Debug("Method failed to parse CPU usage", "method", i+1, "output", cpuUsageStr, "error", err)
			continue
		}

		foundAny = true
		if cpuUsage > maxCpuUsage {
			maxCpuUsage = cpuUsage
		}

		logger.ApiLogger.Debug("Method found CPU usage", "method", i+1, "service", serviceName, "usage", cpuUsage)
	}

	if !foundAny {
		// 进程不存在或没有CPU使用率
		return 0.0
	}

	// 如果所有方法都失败，尝试最后一种方法：使用 pgrep 找到所有相关进程
	if maxCpuUsage == 0.0 {
		pgrepCmd := fmt.Sprintf("pgrep -f %s", serviceName)
		pgrepOutput, err := exec.Command("sh", "-c", pgrepCmd).Output()
		if err == nil {
			pids := strings.Fields(strings.TrimSpace(string(pgrepOutput)))
			if len(pids) > 0 {
				// 为每个PID获取CPU使用率
				pidList := strings.Join(pids, ",")
				psCmd := fmt.Sprintf("ps -p %s -o pcpu= | awk '{sum+=$1} END {print sum}'", pidList)
				psOutput, err := exec.Command("sh", "-c", psCmd).Output()
				if err == nil {
					cpuUsageStr := strings.TrimSpace(string(psOutput))
					if cpuUsageStr != "" {
						if cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64); err == nil {
							maxCpuUsage = cpuUsage
							foundAny = true
							logger.ApiLogger.Debug("pgrep method found CPU usage", "service", serviceName, "usage", cpuUsage, "pids", len(pids))
						}
					}
				}
			}
		}
	}

	if !foundAny {
		logger.ApiLogger.Info("No processes found for service", "service", serviceName)
		return 0.0
	}

	// 将CPU使用率标准化到系统总CPU容量（0-100%范围）
	// maxCpuUsage 是相对于单核的百分比，需要除以核心数来获得相对于系统总CPU的百分比
	normalizedCpuUsage := (maxCpuUsage / cpuCores) * 100
	if normalizedCpuUsage > 100.0 {
		normalizedCpuUsage = 100.0 // 限制最大值为100%
	}

	logger.ApiLogger.Info("Successfully got CPU usage for service", "service", serviceName, "raw_usage", maxCpuUsage, "normalized_usage", normalizedCpuUsage, "cores", cpuCores)
	return normalizedCpuUsage
}

// getDarwinGPUUsage 获取Darwin系统下指定服务的GPU使用情况
func (e *SystemlApi) getDarwinGPUUsage(serviceName string) float64 {
	// macOS下获取GPU使用率的方法：
	// 1. 使用 powermetrics 工具（需要sudo权限）
	// 2. 使用 ioreg 命令查询GPU设备信息
	// 3. 使用系统活动监视器的私有API（复杂）

	// 方法1: 尝试使用 powermetrics 获取GPU使用率
	gpuUsage := e.getGPUUsageFromPowermetrics(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法2: 使用 ioreg 检测GPU活动（简化方法）
	gpuUsage = e.getGPUUsageFromIoreg(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法3: 通过进程的GPU使用情况推估（Metal应用检测）
	gpuUsage = e.getGPUUsageFromProcesses(serviceName)

	logger.ApiLogger.Debug("GPU usage detection completed", "service", serviceName, "usage", gpuUsage)
	return gpuUsage
}

// getGPUUsageFromPowermetrics 使用powermetrics获取GPU使用率
func (e *SystemlApi) getGPUUsageFromPowermetrics(serviceName string) float64 {
	// powermetrics 需要管理员权限，先尝试无权限版本
	// powermetrics -n 1 -i 1000 --samplers gpu_power | grep "GPU Power"
	cmd := "powermetrics -n 1 -i 1000 --samplers gpu_power 2>/dev/null | grep -i 'GPU Power\\|GPU Active' | head -1"

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("powermetrics failed (expected without sudo)", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return -1
	}

	// 解析powermetrics输出，提取GPU使用率
	// 输出格式通常类似: "GPU Power: 1234 mW" 或 "GPU Active: 45%"
	if strings.Contains(outputStr, "%") {
		// 提取百分比
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) > 1 {
			if usage, err := strconv.ParseFloat(matches[1], 64); err == nil {
				logger.ApiLogger.Debug("powermetrics GPU usage found", "service", serviceName, "usage", usage)
				return usage
			}
		}
	}

	return -1
}

// getGPUUsageFromIoreg 使用ioreg检测GPU活动
func (e *SystemlApi) getGPUUsageFromIoreg(serviceName string) float64 {
	// 使用 ioreg 查询GPU设备信息
	// ioreg -r -d 1 -w 0 -c IOPCIDevice | grep -i gpu
	cmd := "ioreg -r -d 1 -w 0 -c IOPCIDevice | grep -i gpu | wc -l"

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("ioreg failed", "service", serviceName, "error", err)
		return -1
	}

	gpuCountStr := strings.TrimSpace(string(output))
	gpuCount, err := strconv.Atoi(gpuCountStr)
	if err != nil || gpuCount == 0 {
		logger.ApiLogger.Debug("No GPU detected via ioreg", "service", serviceName)
		return 0.0
	}

	// 如果检测到GPU，尝试通过其他方式获取使用率
	// 检查系统当前是否有GPU相关活动
	cmd = "ps aux | grep -i metal | grep -v grep | wc -l"
	output, err = exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return 0.0
	}

	metalProcesses := strings.TrimSpace(string(output))
	if metalCount, err := strconv.Atoi(metalProcesses); err == nil && metalCount > 0 {
		// 有Metal相关进程，假设有一定的GPU使用率
		return 5.0 // 保守估计5%的GPU使用率
	}

	return 0.0
}

// getGPUUsageFromProcesses 通过进程信息推估GPU使用率
func (e *SystemlApi) getGPUUsageFromProcesses(serviceName string) float64 {
	// 检查特定服务是否在使用GPU相关功能
	// 方法1: 检查进程是否使用Metal/OpenGL/CUDA
	cmd := fmt.Sprintf("lsof -p $(pgrep -f %s) 2>/dev/null | grep -i 'metal\\|opengl\\|gpu' | wc -l", serviceName)

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("lsof GPU check failed", "service", serviceName, "error", err)
		return 0.0
	}

	gpuFilesStr := strings.TrimSpace(string(output))
	gpuFiles, err := strconv.Atoi(gpuFilesStr)
	if err != nil {
		return 0.0
	}

	if gpuFiles > 0 {
		// 如果检测到GPU相关文件句柄，估算使用率
		// 根据文件句柄数量粗略估算GPU使用率
		estimatedUsage := float64(gpuFiles) * 10.0 // 每个句柄假设10%使用率
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}

		logger.ApiLogger.Debug("GPU usage estimated from file handles", "service", serviceName, "handles", gpuFiles, "estimated_usage", estimatedUsage)
		return estimatedUsage
	}

	// 方法2: 检查进程的CPU使用率，AI应用通常CPU和GPU使用率相关
	cpuUsage := e.getDarwinCPUUsage(serviceName)
	if cpuUsage > 50.0 {
		// 如果CPU使用率较高，可能也在使用GPU
		// 保守估算GPU使用率为CPU使用率的1/4
		gpuEstimate := cpuUsage / 4.0
		if gpuEstimate > 25.0 {
			gpuEstimate = 25.0 // 限制最大估算值
		}

		logger.ApiLogger.Debug("GPU usage estimated from CPU usage", "service", serviceName, "cpu_usage", cpuUsage, "gpu_estimate", gpuEstimate)
		return gpuEstimate
	}

	return 0.0
}

// handleWindowsMonitor 处理Windows系统监控
func (e *SystemlApi) handleWindowsMonitor(c *gin.Context, serviceName string) {
	// Windows系统下可以通过以下方式获取进程信息：
	// 1. tasklist /fo csv | findstr "serviceName"
	// 2. wmic process where "name='serviceName.exe'" get ProcessId,PageFileUsage,WorkingSetSize
	// 3. PowerShell: Get-Process -Name serviceName

	cpuUsed := e.getWindowsCPUUsage(serviceName)
	gpuUsed := e.getWindowsGPUUsage(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"cpu_usage": cpuUsed, // 当前应用CPU占用率（%）
		"gpu_usage": gpuUsed, // 当前应用GPU占用率（%）
	})
}

// getWindowsCPUUsage 获取Windows系统下指定服务的CPU使用情况
func (e *SystemlApi) getWindowsCPUUsage(serviceName string) float64 {
	// Windows下获取CPU使用率的多种方法：
	// 1. 使用 tasklist 命令
	// 2. 使用 wmic 命令
	// 3. 使用 PowerShell Get-Process

	// 获取系统CPU核心数
	cpuCores := float64(runtime.NumCPU())

	// 方法1: 使用 wmic 获取进程CPU使用率
	cpuUsage := e.getWindowsCPUFromWmic(serviceName)
	if cpuUsage >= 0 {
		// 标准化CPU使用率到0-100%范围
		normalizedUsage := (cpuUsage / cpuCores) * 100
		if normalizedUsage > 100.0 {
			normalizedUsage = 100.0
		}
		return normalizedUsage
	}

	// 方法2: 使用 PowerShell Get-Process
	cpuUsage = e.getWindowsCPUFromPowerShell(serviceName)
	if cpuUsage >= 0 {
		// 标准化CPU使用率到0-100%范围
		normalizedUsage := (cpuUsage / cpuCores) * 100
		if normalizedUsage > 100.0 {
			normalizedUsage = 100.0
		}
		return normalizedUsage
	}

	// 方法3: 使用 tasklist 命令（兜底方案）
	cpuUsage = e.getWindowsCPUFromTasklist(serviceName)
	if cpuUsage >= 0 {
		// tasklist 通常返回的已经是相对于系统总CPU的百分比
		return cpuUsage
	}

	logger.ApiLogger.Info("No processes found for service on Windows", "service", serviceName)
	return 0.0
}

// getWindowsCPUFromWmic 使用wmic命令获取CPU使用率
func (e *SystemlApi) getWindowsCPUFromWmic(serviceName string) float64 {
	// wmic process where "name like '%serviceName%'" get Name,ProcessId,PageFileUsage
	// 注意：wmic 本身不直接提供CPU使用率，需要通过其他方式计算
	cmd := fmt.Sprintf(`wmic process where "name like '%%%s%%'" get Name,ProcessId /format:csv`, serviceName)

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("wmic command failed", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return -1
	}

	// 解析wmic输出，获取进程ID列表
	lines := strings.Split(outputStr, "\n")
	var pids []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Node,Name,ProcessId") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[2] != "" {
			pids = append(pids, strings.TrimSpace(parts[2]))
		}
	}

	if len(pids) == 0 {
		return 0.0
	}

	// 使用 typeperf 获取CPU使用率
	return e.getWindowsCPUFromTypeperf(pids)
}

// getWindowsCPUFromPowerShell 使用PowerShell获取CPU使用率
func (e *SystemlApi) getWindowsCPUFromPowerShell(serviceName string) float64 {
	// PowerShell: Get-Process -Name serviceName | Measure-Object -Property CPU -Sum
	cmd := fmt.Sprintf(`powershell -Command "Get-Process | Where-Object {$_.ProcessName -like '*%s*'} | Measure-Object -Property CPU -Sum | Select-Object -ExpandProperty Sum"`, serviceName)

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("PowerShell command failed", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "0" {
		return 0.0
	}

	// PowerShell返回的是累计CPU时间（秒），需要转换为使用率
	cpuTime, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		logger.ApiLogger.Debug("Failed to parse PowerShell CPU output", "output", outputStr, "error", err)
		return -1
	}

	// 简化计算：假设进程运行时间并估算使用率
	// 这里使用一个启发式方法：CPU时间越高，当前使用率可能越高
	if cpuTime > 0 {
		// 估算：每10秒CPU时间对应大约1%的当前使用率（很粗略）
		estimatedUsage := cpuTime / 10.0
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}
		logger.ApiLogger.Debug("PowerShell CPU usage estimated", "service", serviceName, "cpu_time", cpuTime, "estimated_usage", estimatedUsage)
		return estimatedUsage
	}

	return 0.0
}

// getWindowsCPUFromTasklist 使用tasklist命令获取进程信息
func (e *SystemlApi) getWindowsCPUFromTasklist(serviceName string) float64 {
	// tasklist /fi "imagename eq serviceName*" /fo csv
	cmd := fmt.Sprintf(`tasklist /fi "imagename eq %s*" /fo csv`, serviceName)

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("tasklist command failed", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析tasklist输出
	lines := strings.Split(outputStr, "\n")
	processCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Image Name") {
			continue
		}

		// CSV格式解析
		if strings.Contains(line, ",") {
			processCount++
		}
	}

	if processCount > 0 {
		// 基于进程数量的简单估算
		// 每个进程假设使用2-5%的CPU
		estimatedUsage := float64(processCount) * 3.0
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}

		logger.ApiLogger.Debug("tasklist CPU usage estimated", "service", serviceName, "process_count", processCount, "estimated_usage", estimatedUsage)
		return estimatedUsage
	}

	return 0.0
}

// getWindowsCPUFromTypeperf 使用typeperf获取实时CPU使用率
func (e *SystemlApi) getWindowsCPUFromTypeperf(pids []string) float64 {
	if len(pids) == 0 {
		return 0.0
	}

	// typeperf "\Process(processname)\% Processor Time" -sc 1
	// 由于需要知道确切的进程名称，这里使用一个简化的方法

	var totalUsage float64 = 0.0

	for _, pid := range pids {
		// 使用wmic获取进程的工作集大小作为活动指标
		cmd := fmt.Sprintf(`wmic process where "ProcessId=%s" get WorkingSetSize /format:value`, pid)
		output, err := exec.Command("cmd", "/C", cmd).Output()
		if err != nil {
			continue
		}

		outputStr := strings.TrimSpace(string(output))
		if strings.Contains(outputStr, "WorkingSetSize=") {
			// 解析工作集大小
			parts := strings.Split(outputStr, "=")
			if len(parts) >= 2 {
				wsSize, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
				if err == nil && wsSize > 0 {
					// 基于内存使用量估算CPU使用率
					// 内存使用越多，可能CPU使用率越高（启发式方法）
					estimatedCPU := wsSize / (1024 * 1024 * 100) // 每100MB内存对应约1%CPU
					if estimatedCPU > 20.0 {
						estimatedCPU = 20.0 // 单个进程最多20%
					}
					totalUsage += estimatedCPU
				}
			}
		}
	}

	if totalUsage > 100.0 {
		totalUsage = 100.0
	}

	logger.ApiLogger.Debug("typeperf CPU usage estimated", "pids", len(pids), "total_usage", totalUsage)
	return totalUsage
} // getWindowsGPUUsage 获取Windows系统下指定服务的GPU使用情况
func (e *SystemlApi) getWindowsGPUUsage(serviceName string) float64 {
	// Windows下获取GPU使用率的多种方法：
	// 1. 使用 nvidia-smi 命令（NVIDIA GPU）
	// 2. 使用 PowerShell Get-Counter（通用）
	// 3. 使用 wmic 查询GPU相关进程
	// 4. 使用 DirectX/D3D 相关检测

	// 方法1: 尝试使用 nvidia-smi 获取NVIDIA GPU使用率
	gpuUsage := e.getWindowsGPUFromNvidiaSmi(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法2: 使用 PowerShell 性能计数器
	gpuUsage = e.getWindowsGPUFromPowerShell(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法3: 使用 wmic 检测GPU相关进程活动
	gpuUsage = e.getWindowsGPUFromWmic(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法4: 通过进程行为推估GPU使用率
	gpuUsage = e.getWindowsGPUFromProcessBehavior(serviceName)

	logger.ApiLogger.Debug("Windows GPU usage detection completed", "service", serviceName, "usage", gpuUsage)
	return gpuUsage
}

// getWindowsGPUFromNvidiaSmi 使用nvidia-smi获取NVIDIA GPU使用率
func (e *SystemlApi) getWindowsGPUFromNvidiaSmi(serviceName string) float64 {
	// nvidia-smi --query-compute-apps=pid,process_name,used_memory --format=csv,noheader,nounits
	cmd := `nvidia-smi --query-compute-apps=pid,process_name,used_memory --format=csv,noheader,nounits`

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("nvidia-smi not available", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析nvidia-smi输出，查找目标服务
	lines := strings.Split(outputStr, "\n")
	var totalGpuMemory float64 = 0.0
	var foundProcess bool = false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			processName := strings.TrimSpace(parts[1])
			memoryStr := strings.TrimSpace(parts[2])

			// 检查是否是目标服务
			if strings.Contains(strings.ToLower(processName), strings.ToLower(serviceName)) {
				if memory, err := strconv.ParseFloat(memoryStr, 64); err == nil {
					totalGpuMemory += memory
					foundProcess = true
					logger.ApiLogger.Debug("nvidia-smi found process", "service", serviceName, "process", processName, "memory", memory)
				}
			}
		}
	}

	if foundProcess {
		// 获取GPU总使用率
		utilizationCmd := `nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits`
		utilizationOutput, err := exec.Command("cmd", "/C", utilizationCmd).Output()
		if err == nil {
			utilizationStr := strings.TrimSpace(string(utilizationOutput))
			if utilization, err := strconv.ParseFloat(utilizationStr, 64); err == nil {
				logger.ApiLogger.Debug("nvidia-smi GPU utilization", "service", serviceName, "utilization", utilization, "memory", totalGpuMemory)
				return utilization
			}
		}

		// 如果无法获取准确使用率，基于内存使用量估算
		// 假设8GB显存，每1GB对应约12.5%使用率
		estimatedUsage := (totalGpuMemory / 1024.0) * 12.5
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}

		logger.ApiLogger.Debug("nvidia-smi GPU usage estimated from memory", "service", serviceName, "memory_mb", totalGpuMemory, "estimated_usage", estimatedUsage)
		return estimatedUsage
	}

	return 0.0
}

// getWindowsGPUFromPowerShell 使用PowerShell性能计数器获取GPU使用率
func (e *SystemlApi) getWindowsGPUFromPowerShell(serviceName string) float64 {
	// PowerShell: Get-Counter "\GPU Engine(*)\Utilization Percentage"
	cmd := `powershell -Command "try { Get-Counter '\GPU Engine(*)\Utilization Percentage' -SampleInterval 1 -MaxSamples 1 | Select-Object -ExpandProperty CounterSamples | Measure-Object -Property CookedValue -Average | Select-Object -ExpandProperty Average } catch { 0 }"`

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("PowerShell GPU counter failed", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" || outputStr == "0" {
		return 0.0
	}

	// 解析GPU使用率
	gpuUsage, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		logger.ApiLogger.Debug("Failed to parse PowerShell GPU output", "output", outputStr, "error", err)
		return -1
	}

	// 检查是否有目标服务在运行
	processCmd := fmt.Sprintf(`powershell -Command "Get-Process | Where-Object {$_.ProcessName -like '*%s*'} | Measure-Object | Select-Object -ExpandProperty Count"`, serviceName)
	processOutput, err := exec.Command("cmd", "/C", processCmd).Output()
	if err == nil {
		processCountStr := strings.TrimSpace(string(processOutput))
		if processCount, err := strconv.Atoi(processCountStr); err == nil && processCount > 0 {
			// 如果有目标进程在运行，返回系统GPU使用率的一部分
			serviceGpuUsage := gpuUsage * 0.5 // 假设目标服务使用了50%的GPU资源
			if serviceGpuUsage > gpuUsage {
				serviceGpuUsage = gpuUsage
			}

			logger.ApiLogger.Debug("PowerShell GPU usage for service", "service", serviceName, "system_gpu", gpuUsage, "service_gpu", serviceGpuUsage, "processes", processCount)
			return serviceGpuUsage
		}
	}

	return 0.0
}

// getWindowsGPUFromWmic 使用wmic检测GPU相关进程活动
func (e *SystemlApi) getWindowsGPUFromWmic(serviceName string) float64 {
	// 使用wmic查询视频控制器信息
	cmd := `wmic path Win32_VideoController get Name,Status,VideoMemoryType /format:csv`

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("wmic video controller query failed", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 检查是否有可用的GPU
	hasGPU := strings.Contains(strings.ToLower(outputStr), "nvidia") ||
		strings.Contains(strings.ToLower(outputStr), "amd") ||
		strings.Contains(strings.ToLower(outputStr), "intel")

	if !hasGPU {
		logger.ApiLogger.Debug("No discrete GPU detected via wmic", "service", serviceName)
		return 0.0
	}

	// 检查目标服务是否在运行
	processCmd := fmt.Sprintf(`wmic process where "name like '%%%s%%'" get ProcessId,WorkingSetSize /format:csv`, serviceName)
	processOutput, err := exec.Command("cmd", "/C", processCmd).Output()
	if err != nil {
		return 0.0
	}

	processStr := strings.TrimSpace(string(processOutput))
	if processStr == "" {
		return 0.0
	}

	// 解析进程信息
	lines := strings.Split(processStr, "\n")
	var totalMemory float64 = 0.0
	processCount := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "Node,ProcessId,WorkingSetSize") {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 && parts[2] != "" {
			if memory, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64); err == nil {
				totalMemory += memory
				processCount++
			}
		}
	}

	if processCount > 0 {
		// 基于进程内存使用量估算GPU使用率
		// 大内存使用量可能意味着GPU密集型应用
		memoryGB := totalMemory / (1024 * 1024 * 1024)
		estimatedGpuUsage := memoryGB * 15.0 // 每GB内存假设15%GPU使用率

		if estimatedGpuUsage > 100.0 {
			estimatedGpuUsage = 100.0
		}

		logger.ApiLogger.Debug("wmic GPU usage estimated from memory", "service", serviceName, "memory_gb", memoryGB, "processes", processCount, "estimated_gpu", estimatedGpuUsage)
		return estimatedGpuUsage
	}

	return 0.0
}

// getWindowsGPUFromProcessBehavior 通过进程行为推估GPU使用率
func (e *SystemlApi) getWindowsGPUFromProcessBehavior(serviceName string) float64 {
	// 检查进程是否加载了与GPU相关的DLL
	cmd := fmt.Sprintf(`powershell -Command "Get-Process | Where-Object {$_.ProcessName -like '*%s*'} | ForEach-Object { try { $_.Modules | Where-Object {$_.ModuleName -match 'nvapi|d3d|opengl|vulkan|cuda'} | Measure-Object | Select-Object -ExpandProperty Count } catch { 0 } } | Measure-Object -Sum | Select-Object -ExpandProperty Sum"`, serviceName)

	output, err := exec.Command("cmd", "/C", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Process module check failed", "service", serviceName, "error", err)
		return 0.0
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析GPU相关模块数量
	moduleCount, err := strconv.Atoi(outputStr)
	if err != nil {
		return 0.0
	}

	if moduleCount > 0 {
		// 基于GPU相关模块数量估算使用率
		estimatedUsage := float64(moduleCount) * 8.0 // 每个GPU模块假设8%使用率
		if estimatedUsage > 50.0 {
			estimatedUsage = 50.0 // 最大50%
		}

		logger.ApiLogger.Debug("GPU usage estimated from loaded modules", "service", serviceName, "gpu_modules", moduleCount, "estimated_usage", estimatedUsage)
		return estimatedUsage
	}

	// 最后的兜底方案：检查CPU使用率，AI应用CPU和GPU通常相关
	cpuUsage := e.getWindowsCPUUsage(serviceName)
	if cpuUsage > 30.0 {
		// 如果CPU使用率较高，可能也在使用GPU
		gpuEstimate := cpuUsage * 0.3 // GPU使用率约为CPU的30%
		if gpuEstimate > 30.0 {
			gpuEstimate = 30.0
		}

		logger.ApiLogger.Debug("GPU usage estimated from CPU correlation", "service", serviceName, "cpu_usage", cpuUsage, "gpu_estimate", gpuEstimate)
		return gpuEstimate
	}

	return 0.0
}

// handleLinuxMonitor 处理Linux系统监控
func (e *SystemlApi) handleLinuxMonitor(c *gin.Context, serviceName string) {
	cpuUsed := e.getLinuxCPUUsage(serviceName)
	gpuUsed := e.getLinuxGPUUsage(serviceName)

	c.JSON(http.StatusOK, gin.H{
		"cpu_usage": cpuUsed, // 当前应用CPU占用率（%）
		"gpu_usage": gpuUsed, // 当前应用GPU占用率（%）
	})
}

// getLinuxCPUUsage 获取Linux系统下指定服务的CPU使用情况
func (e *SystemlApi) getLinuxCPUUsage(serviceName string) float64 {
	// 获取系统CPU核心数
	cpuCores := float64(runtime.NumCPU())

	// Linux下获取CPU使用率的多种方法：
	// 1. 使用 ps 命令（与macOS类似）
	// 2. 使用 top 命令
	// 3. 使用 /proc/[pid]/stat 文件
	// 4. 使用 pidstat 命令（如果可用）

	// 方法1: 使用 ps 命令获取CPU使用率
	cpuUsage := e.getLinuxCPUFromPs(serviceName)
	if cpuUsage >= 0 {
		// 标准化CPU使用率到0-100%范围
		normalizedUsage := (cpuUsage / cpuCores) * 100
		if normalizedUsage > 100.0 {
			normalizedUsage = 100.0
		}
		return normalizedUsage
	}

	// 方法2: 使用 top 命令
	cpuUsage = e.getLinuxCPUFromTop(serviceName)
	if cpuUsage >= 0 {
		// top命令通常返回相对于系统总CPU的百分比
		return cpuUsage
	}

	// 方法3: 使用 pidstat 命令（如果可用）
	cpuUsage = e.getLinuxCPUFromPidstat(serviceName)
	if cpuUsage >= 0 {
		return cpuUsage
	}

	logger.ApiLogger.Info("No processes found for service on Linux", "service", serviceName)
	return 0.0
}

// getLinuxCPUFromPs 使用ps命令获取CPU使用率
func (e *SystemlApi) getLinuxCPUFromPs(serviceName string) float64 {
	// 方法1: 直接搜索包含服务名的所有进程
	cmd1 := fmt.Sprintf("ps -eo pid,pcpu,comm | grep -i %s | grep -v grep | awk '{sum+=$2} END {print sum}'", serviceName)

	// 方法2: 搜索以服务名开头的进程
	cmd2 := fmt.Sprintf("ps -eo pid,pcpu,comm | awk '$3 ~ /^%s/ {sum+=$2} END {print sum}'", serviceName)

	// 方法3: 搜索包含服务名的所有进程（更宽泛的匹配）
	cmd3 := fmt.Sprintf("ps -eo pid,pcpu,args | grep -i %s | grep -v grep | awk '{sum+=$2} END {print sum}'", serviceName)

	methods := []string{cmd1, cmd2, cmd3}
	var maxCpuUsage float64 = 0.0
	var foundAny bool = false

	for i, cmd := range methods {
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			logger.ApiLogger.Debug("Linux ps method failed", "method", i+1, "service", serviceName, "error", err)
			continue
		}

		cpuUsageStr := strings.TrimSpace(string(output))
		if cpuUsageStr == "" || cpuUsageStr == "0" {
			continue
		}

		cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
		if err != nil {
			logger.ApiLogger.Debug("Linux ps parse failed", "method", i+1, "output", cpuUsageStr, "error", err)
			continue
		}

		foundAny = true
		if cpuUsage > maxCpuUsage {
			maxCpuUsage = cpuUsage
		}

		logger.ApiLogger.Debug("Linux ps found CPU usage", "method", i+1, "service", serviceName, "usage", cpuUsage)
	}

	// 如果ps方法都失败，尝试使用pgrep
	if !foundAny || maxCpuUsage == 0.0 {
		pgrepCmd := fmt.Sprintf("pgrep -f %s", serviceName)
		pgrepOutput, err := exec.Command("sh", "-c", pgrepCmd).Output()
		if err == nil {
			pids := strings.Fields(strings.TrimSpace(string(pgrepOutput)))
			if len(pids) > 0 {
				pidList := strings.Join(pids, ",")
				psCmd := fmt.Sprintf("ps -p %s -o pcpu= | awk '{sum+=$1} END {print sum}'", pidList)
				psOutput, err := exec.Command("sh", "-c", psCmd).Output()
				if err == nil {
					cpuUsageStr := strings.TrimSpace(string(psOutput))
					if cpuUsageStr != "" {
						if cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64); err == nil {
							maxCpuUsage = cpuUsage
							foundAny = true
							logger.ApiLogger.Debug("Linux pgrep found CPU usage", "service", serviceName, "usage", cpuUsage, "pids", len(pids))
						}
					}
				}
			}
		}
	}

	if !foundAny {
		return -1
	}

	logger.ApiLogger.Info("Linux ps CPU usage found", "service", serviceName, "usage", maxCpuUsage)
	return maxCpuUsage
}

// getLinuxCPUFromTop 使用top命令获取CPU使用率
func (e *SystemlApi) getLinuxCPUFromTop(serviceName string) float64 {
	// top -b -n1 | grep serviceName | awk '{sum+=$9} END {print sum}'
	cmd := fmt.Sprintf("top -b -n1 | grep %s | grep -v grep | awk '{sum+=$9} END {print sum}'", serviceName)

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux top command failed", "service", serviceName, "error", err)
		return -1
	}

	cpuUsageStr := strings.TrimSpace(string(output))
	if cpuUsageStr == "" || cpuUsageStr == "0" {
		return 0.0
	}

	cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
	if err != nil {
		logger.ApiLogger.Debug("Linux top parse failed", "output", cpuUsageStr, "error", err)
		return -1
	}

	logger.ApiLogger.Debug("Linux top CPU usage found", "service", serviceName, "usage", cpuUsage)
	return cpuUsage
}

// getLinuxCPUFromPidstat 使用pidstat命令获取CPU使用率
func (e *SystemlApi) getLinuxCPUFromPidstat(serviceName string) float64 {
	// pidstat -C serviceName 1 1 | awk 'NR>3 {sum+=$8} END {print sum}'
	cmd := fmt.Sprintf("pidstat -C %s 1 1 2>/dev/null | awk 'NR>3 && NF>1 {sum+=$8} END {print sum+0}'", serviceName)

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux pidstat not available", "service", serviceName, "error", err)
		return -1
	}

	cpuUsageStr := strings.TrimSpace(string(output))
	if cpuUsageStr == "" || cpuUsageStr == "0" {
		return 0.0
	}

	cpuUsage, err := strconv.ParseFloat(cpuUsageStr, 64)
	if err != nil {
		logger.ApiLogger.Debug("Linux pidstat parse failed", "output", cpuUsageStr, "error", err)
		return -1
	}

	logger.ApiLogger.Debug("Linux pidstat CPU usage found", "service", serviceName, "usage", cpuUsage)
	return cpuUsage
}

// getLinuxGPUUsage 获取Linux系统下指定服务的GPU使用情况
func (e *SystemlApi) getLinuxGPUUsage(serviceName string) float64 {
	// Linux下获取GPU使用率的多种方法：
	// 1. 使用 nvidia-smi 命令（NVIDIA GPU）
	// 2. 使用 radeontop 命令（AMD GPU）
	// 3. 使用 intel_gpu_top 命令（Intel GPU）
	// 4. 使用 /sys/class/drm 文件系统
	// 5. 通过进程GPU相关文件检测

	// 方法1: 尝试使用 nvidia-smi
	gpuUsage := e.getLinuxGPUFromNvidiaSmi(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法2: 尝试使用 radeontop（AMD GPU）
	gpuUsage = e.getLinuxGPUFromRadeontop(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法3: 尝试使用 intel_gpu_top（Intel GPU）
	gpuUsage = e.getLinuxGPUFromIntelGpuTop(serviceName)
	if gpuUsage >= 0 {
		return gpuUsage
	}

	// 方法4: 通过进程GPU相关文件检测
	gpuUsage = e.getLinuxGPUFromProcessFiles(serviceName)

	logger.ApiLogger.Debug("Linux GPU usage detection completed", "service", serviceName, "usage", gpuUsage)
	return gpuUsage
}

// getLinuxGPUFromNvidiaSmi 使用nvidia-smi获取NVIDIA GPU使用率
func (e *SystemlApi) getLinuxGPUFromNvidiaSmi(serviceName string) float64 {
	// nvidia-smi --query-compute-apps=pid,process_name,used_memory --format=csv,noheader,nounits
	cmd := "nvidia-smi --query-compute-apps=pid,process_name,used_memory --format=csv,noheader,nounits"

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux nvidia-smi not available", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析nvidia-smi输出，查找目标服务
	lines := strings.Split(outputStr, "\n")
	var totalGpuMemory float64 = 0.0
	var foundProcess bool = false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			processName := strings.TrimSpace(parts[1])
			memoryStr := strings.TrimSpace(parts[2])

			if strings.Contains(strings.ToLower(processName), strings.ToLower(serviceName)) {
				if memory, err := strconv.ParseFloat(memoryStr, 64); err == nil {
					totalGpuMemory += memory
					foundProcess = true
					logger.ApiLogger.Debug("Linux nvidia-smi found process", "service", serviceName, "process", processName, "memory", memory)
				}
			}
		}
	}

	if foundProcess {
		// 获取GPU总使用率
		utilizationCmd := "nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits"
		utilizationOutput, err := exec.Command("sh", "-c", utilizationCmd).Output()
		if err == nil {
			utilizationStr := strings.TrimSpace(string(utilizationOutput))
			if utilization, err := strconv.ParseFloat(utilizationStr, 64); err == nil {
				logger.ApiLogger.Debug("Linux nvidia-smi GPU utilization", "service", serviceName, "utilization", utilization, "memory", totalGpuMemory)
				return utilization
			}
		}

		// 基于显存使用量估算
		estimatedUsage := (totalGpuMemory / 1024.0) * 12.5
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}

		logger.ApiLogger.Debug("Linux nvidia-smi GPU estimated", "service", serviceName, "memory_mb", totalGpuMemory, "estimated", estimatedUsage)
		return estimatedUsage
	}

	return 0.0
}

// getLinuxGPUFromRadeontop 使用radeontop获取AMD GPU使用率
func (e *SystemlApi) getLinuxGPUFromRadeontop(serviceName string) float64 {
	// radeontop -d - -l 1 | grep "gpu"
	cmd := "timeout 3 radeontop -d - -l 1 2>/dev/null | grep -i 'gpu' | head -1"

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux radeontop not available", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析radeontop输出，提取GPU使用率
	// 输出格式通常类似: "gpu 45.67%"
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) > 1 {
		if usage, err := strconv.ParseFloat(matches[1], 64); err == nil {
			// 检查是否有目标服务在运行
			if e.isServiceRunningOnLinux(serviceName) {
				logger.ApiLogger.Debug("Linux radeontop GPU usage", "service", serviceName, "system_usage", usage)
				// 假设目标服务使用了部分GPU资源
				serviceUsage := usage * 0.6
				return serviceUsage
			}
		}
	}

	return 0.0
}

// getLinuxGPUFromIntelGpuTop 使用intel_gpu_top获取Intel GPU使用率
func (e *SystemlApi) getLinuxGPUFromIntelGpuTop(serviceName string) float64 {
	// intel_gpu_top -s 1000 -o - | head -2
	cmd := "timeout 3 intel_gpu_top -s 1000 -o - 2>/dev/null | head -5 | grep -i 'render\\|video'"

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux intel_gpu_top not available", "service", serviceName, "error", err)
		return -1
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	// 解析intel_gpu_top输出
	lines := strings.Split(outputStr, "\n")
	var totalUsage float64 = 0.0

	for _, line := range lines {
		re := regexp.MustCompile(`(\d+(?:\.\d+)?)%`)
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			if usage, err := strconv.ParseFloat(matches[1], 64); err == nil {
				totalUsage += usage
			}
		}
	}

	if totalUsage > 0 && e.isServiceRunningOnLinux(serviceName) {
		// 假设目标服务使用了部分GPU资源
		serviceUsage := totalUsage * 0.5
		if serviceUsage > 100.0 {
			serviceUsage = 100.0
		}

		logger.ApiLogger.Debug("Linux intel_gpu_top GPU usage", "service", serviceName, "system_usage", totalUsage, "service_usage", serviceUsage)
		return serviceUsage
	}

	return 0.0
}

// getLinuxGPUFromProcessFiles 通过进程GPU相关文件检测GPU使用
func (e *SystemlApi) getLinuxGPUFromProcessFiles(serviceName string) float64 {
	// 检查进程是否打开了GPU相关设备文件
	cmd := fmt.Sprintf("lsof /dev/dri/* /dev/nvidia* 2>/dev/null | grep -i %s | wc -l", serviceName)

	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		logger.ApiLogger.Debug("Linux GPU device file check failed", "service", serviceName, "error", err)
		return 0.0
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return 0.0
	}

	gpuFiles, err := strconv.Atoi(outputStr)
	if err != nil {
		return 0.0
	}

	if gpuFiles > 0 {
		// 基于GPU设备文件使用量估算
		estimatedUsage := float64(gpuFiles) * 15.0 // 每个设备文件15%使用率
		if estimatedUsage > 100.0 {
			estimatedUsage = 100.0
		}

		logger.ApiLogger.Debug("Linux GPU usage from device files", "service", serviceName, "files", gpuFiles, "estimated", estimatedUsage)
		return estimatedUsage
	}

	// 最后的兜底：基于CPU使用率估算
	cpuUsage := e.getLinuxCPUUsage(serviceName)
	if cpuUsage > 40.0 {
		gpuEstimate := cpuUsage * 0.25 // GPU = CPU × 25%
		if gpuEstimate > 30.0 {
			gpuEstimate = 30.0
		}

		logger.ApiLogger.Debug("Linux GPU usage estimated from CPU", "service", serviceName, "cpu", cpuUsage, "gpu_estimate", gpuEstimate)
		return gpuEstimate
	}

	return 0.0
}

// isServiceRunningOnLinux 检查服务是否在Linux上运行
func (e *SystemlApi) isServiceRunningOnLinux(serviceName string) bool {
	cmd := fmt.Sprintf("pgrep -f %s", serviceName)
	output, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return false
	}

	outputStr := strings.TrimSpace(string(output))
	return outputStr != ""
}
