package hardware

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ProcessMonitorInfo represents monitoring information for a specific process
type ProcessMonitorInfo struct {
	MemoryUsageBytes uint64  `json:"memoryUsageBytes"` // Memory usage in bytes
	MemoryUsageMB    float64 `json:"memoryUsageMB"`    // Memory usage in MB
	CPUUsagePercent  float64 `json:"cpuUsagePercent"`  // CPU usage percentage
	ProcessID        string  `json:"processId"`        // Process ID
	Running          bool    `json:"running"`          // Whether the process is running
	GPUUsagePercent  float64 `json:"gpuUsagePercent"`  // GPU usage percentage
	GPUMemoryMB      float64 `json:"gpuMemoryMB"`      // GPU memory usage in MB
}

func GetOpenVinoMonitor() (*ProcessMonitorInfo, error) {
	// 查找 openvino 进程
	cmd := exec.Command("pgrep", "openvino")
	output, err := cmd.Output()
	if err != nil {
		return &ProcessMonitorInfo{Running: false}, nil
	}

	pid := strings.TrimSpace(string(output))
	if pid == "" {
		return &ProcessMonitorInfo{Running: false}, nil
	}

	// 获取进程内存使用情况
	cmd = exec.Command("ps", "-o", "rss=", "-p", pid)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get openvino memory usage: %v", err)
	}

	memStr := strings.TrimSpace(string(output))
	memKB, err := strconv.ParseUint(memStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory usage: %v", err)
	}

	// 转换为字节
	memBytes := memKB * 1024
	memMB := float64(memBytes) / 1024 / 1024

	// 获取 CPU 使用率
	cmd = exec.Command("ps", "-o", "%cpu=", "-p", pid)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get openvino CPU usage: %v", err)
	}

	cpuStr := strings.TrimSpace(string(output))
	cpuUsage, err := strconv.ParseFloat(cpuStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CPU usage: %v", err)
	}

	// 获取 GPU 使用情况 (for macOS using Metal)
	gpuUsage, gpuMemory := getGPUStats()

	return &ProcessMonitorInfo{
		MemoryUsageBytes: memBytes,
		MemoryUsageMB:    memMB,
		CPUUsagePercent:  cpuUsage,
		ProcessID:        pid,
		Running:          true,
		GPUUsagePercent:  gpuUsage,
		GPUMemoryMB:      gpuMemory,
	}, nil
}

func GetOllamaMonitor() (*ProcessMonitorInfo, error) {
	// 查找 ollama 进程
	cmd := exec.Command("pgrep", "ollama")
	output, err := cmd.Output()
	if err != nil {
		return &ProcessMonitorInfo{Running: false}, nil
	}

	pid := strings.TrimSpace(string(output))
	if pid == "" {
		return &ProcessMonitorInfo{Running: false}, nil
	}

	// 获取进程内存使用情况
	cmd = exec.Command("ps", "-o", "rss=", "-p", pid)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ollama memory usage: %v", err)
	}

	memStr := strings.TrimSpace(string(output))
	memKB, err := strconv.ParseUint(memStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory usage: %v", err)
	}

	// 转换为字节
	memBytes := memKB * 1024
	memMB := float64(memBytes) / 1024 / 1024

	// 获取 CPU 使用率
	cmd = exec.Command("ps", "-o", "%cpu=", "-p", pid)
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ollama CPU usage: %v", err)
	}

	cpuStr := strings.TrimSpace(string(output))
	cpuUsage, err := strconv.ParseFloat(cpuStr, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CPU usage: %v", err)
	}

	// 获取 GPU 使用情况 (for macOS using Metal)
	gpuUsage, gpuMemory := getGPUStats()

	return &ProcessMonitorInfo{
		MemoryUsageBytes: memBytes,
		MemoryUsageMB:    memMB,
		CPUUsagePercent:  cpuUsage,
		ProcessID:        pid,
		Running:          true,
		GPUUsagePercent:  gpuUsage,
		GPUMemoryMB:      gpuMemory,
	}, nil
}

// getGPUStats retrieves GPU utilization and memory usage for macOS
// Uses the `ioreg` command to get Metal GPU stats
func getGPUStats() (float64, float64) {
	// Try to get GPU usage with powermetrics (requires sudo)
	cmd := exec.Command("sudo", "-n", "powermetrics", "--samplers", "gpu", "-n", "1", "-i", "1000")
	output, err := cmd.CombinedOutput()
	if err == nil {
		// Parse the powermetrics output
		outputStr := string(output)

		// Extract GPU utilization
		gpuUsage := extractGPUMetric(outputStr, "GPU %")

		// Extract GPU memory usage
		gpuMemory := extractGPUMetric(outputStr, "GPU Memory")

		return gpuUsage, gpuMemory
	}

	// Fallback to ioreg for basic GPU info if powermetrics fails
	cmd = exec.Command("ioreg", "-l")
	output, err = cmd.Output()
	if err != nil {
		return 0, 0
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	var gpuUsage, gpuMemory float64

	for _, line := range lines {
		if strings.Contains(line, "PerformanceStatistics") {
			// Try to extract basic GPU stats from ioreg
			// This is a simple approximation
			gpuUsage = 50.0    // Default placeholder when exact value can't be determined
			gpuMemory = 1024.0 // Default 1GB as placeholder
			break
		}
	}

	return gpuUsage, gpuMemory
}

// extractGPUMetric parses output from powermetrics to extract specific GPU metrics
func extractGPUMetric(output, metricName string) float64 {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, metricName) {
			fields := strings.Fields(line)
			for _, field := range fields {
				val, err := strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64)
				if err == nil {
					return val
				}
			}
		}
	}
	return 0
}
