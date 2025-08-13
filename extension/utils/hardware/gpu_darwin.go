//go:build darwin

package hardware

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/mem"
)

// getWindowsGPUInfo 在macOS平台返回错误
func getWindowsGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Windows GPU info not available on macOS")
}

// getLinuxGPUInfo 在macOS平台返回错误
func getLinuxGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Linux GPU info not available on macOS")
}

// getVRAMInfo 获取macOS VRAM信息
func getVRAMInfo() (total, used, free uint64) {
	// 优先检测Apple Silicon GPU
	if runtime.GOARCH == "arm64" {
		if appleTotal, appleUsed := getAppleSiliconGPUMemory(); appleTotal > 0 {
			return appleTotal, appleUsed, appleTotal - appleUsed
		}
	}

	// 检测独立显卡
	if dedicatedTotal, dedicatedUsed := getDedicatedGPUMemory(); dedicatedTotal > 0 {
		return dedicatedTotal, dedicatedUsed, dedicatedTotal - dedicatedUsed
	}

	// 使用system_profiler获取显卡信息
	if profilerTotal := getSystemProfilerVRAM(); profilerTotal > 0 {
		return profilerTotal, 0, profilerTotal
	}

	return 0, 0, 0
}

// getMetalVRAM 使用Metal框架获取GPU内存信息
func getMetalVRAM() (total, used uint64) {
	// 尝试使用Metal工具（如果可用）
	cmd := exec.Command("system_profiler", "SPDisplaysDataType", "-xml")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	// 解析XML输出获取显存信息
	totalMem := parseDisplayMemoryFromXML(string(output))
	if totalMem > 0 {
		// macOS较难获取准确的已使用显存，返回总量
		return totalMem, 0
	}

	return 0, 0
}

// getAppleSiliconGPUMemory 获取Apple Silicon GPU内存信息
func getAppleSiliconGPUMemory() (total, used uint64) {
	// Apple Silicon使用统一内存架构，GPU内存是系统内存的一部分
	// 通过vm_stat命令获取内存使用情况
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")
	var pageSize, freePages, inactivePages, speculative, cached uint64

	for _, line := range lines {
		if strings.Contains(line, "page size of") {
			// 例如: "Mach Virtual Memory Statistics: (page size of 16384 bytes)"
			fmt.Sscanf(line, "Mach Virtual Memory Statistics: (page size of %d bytes)", &pageSize)
		} else if strings.Contains(line, "Pages free:") {
			fmt.Sscanf(line, "Pages free: %d.", &freePages)
		} else if strings.Contains(line, "Pages inactive:") {
			fmt.Sscanf(line, "Pages inactive: %d.", &inactivePages)
		} else if strings.Contains(line, "Pages speculative:") {
			fmt.Sscanf(line, "Pages speculative: %d.", &speculative)
		} else if strings.Contains(line, "File-backed pages:") {
			fmt.Sscanf(line, "File-backed pages: %d.", &cached)
		}
	}

	if pageSize > 0 {
		// Apple Silicon通常为GPU预留系统内存的25-30%
		totalSystemMemory := getTotalSystemMemory()
		estimatedGPUMemory := totalSystemMemory * 3 / 10 // 30%

		// 估算GPU使用的内存（非精确）
		usedMemory := estimatedGPUMemory * 2 / 10 // 估算20%被GPU使用

		return estimatedGPUMemory, usedMemory
	}

	return 0, 0
}

// getDedicatedGPUMemory 获取独立显卡内存信息
func getDedicatedGPUMemory() (total, used uint64) {
	// 使用system_profiler检查是否有独立显卡
	cmd := exec.Command("system_profiler", "SPDisplaysDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	return parseGPUMemoryFromJSON(output)
}

// parseGPUMemoryFromJSON 解析system_profiler JSON输出获取GPU内存
func parseGPUMemoryFromJSON(jsonData []byte) (total, used uint64) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return 0, 0
	}

	displays, ok := data["SPDisplaysDataType"].([]interface{})
	if !ok {
		return 0, 0
	}

	for _, display := range displays {
		displayInfo, ok := display.(map[string]interface{})
		if !ok {
			continue
		}

		// 检查VRAM大小
		if vramStr, exists := displayInfo["spdisplays_vram"].(string); exists {
			// 解析VRAM字符串，如 "8 GB", "512 MB"
			if memory := extractMemorySize(vramStr); memory > 0 {
				return memory, memory / 10 // 假设使用10%
			}
		}

		// 检查共享内存
		if vramShared, exists := displayInfo["spdisplays_vram_shared"].(string); exists {
			if memory := extractMemorySize(vramShared); memory > 0 {
				return memory, memory / 20 // 假设使用5%
			}
		}
	}

	return 0, 0
}

// getTotalSystemMemory 获取系统总内存
func getTotalSystemMemory() uint64 {
	if v, err := mem.VirtualMemory(); err == nil {
		return v.Total
	}
	return 0
}

// getMacOSGPUUtilization 获取macOS GPU使用率
func getMacOSGPUUtilization() float64 {
	// 尝试通过Activity Monitor的PowerMetrics获取GPU使用率
	cmd := exec.Command("powermetrics", "-n", "1", "-s", "gpu_power", "--show-usage-summary")
	output, err := cmd.Output()
	if err != nil {
		// 如果powermetrics不可用，尝试使用ioreg
		return getGPUUtilizationFromIOReg()
	}

	// 解析powermetrics输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "GPU HW active residency") || strings.Contains(line, "GPU active") {
			// 尝试解析百分比
			if strings.Contains(line, "%") {
				var usage float64
				if n, err := fmt.Sscanf(line, "%*s %f%%", &usage); n == 1 && err == nil {
					return usage
				}
			}
		}
	}

	return -1 // 无法获取
}

// getGPUUtilizationFromIOReg 通过ioreg获取GPU使用率（备用方法）
func getGPUUtilizationFromIOReg() float64 {
	cmd := exec.Command("ioreg", "-r", "-d", "1", "-w", "0", "-c", "IOAccelerator")
	output, err := cmd.Output()
	if err != nil {
		return -1
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 查找GPU相关的活动指标
		if strings.Contains(line, "PerformanceStatistics") {
			// 这是一个简化的实现，实际解析会更复杂
			// 返回一个估算值
			return 0.0 // 默认返回0%，表示空闲
		}
	}

	return -1 // 无法获取
}

// getSystemProfilerVRAM 使用system_profiler获取显存信息
func getSystemProfilerVRAM() uint64 {
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	return parseDisplayMemory(string(output))
}

// parseDisplayMemory 解析显示器信息中的显存大小
func parseDisplayMemory(output string) uint64 {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 查找VRAM相关信息
		if strings.Contains(line, "VRAM") || strings.Contains(line, "Memory") {
			// 尝试提取内存大小
			if mem := extractMemorySize(line); mem > 0 {
				return mem
			}
		}
	}

	return 0
}

// parseDisplayMemoryFromXML 从XML输出解析显存信息
func parseDisplayMemoryFromXML(output string) uint64 {
	// 简单的XML解析，查找内存相关字段
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if strings.Contains(line, "spdisplays_vram") || strings.Contains(line, "spdisplays_memory") {
			// 查找下一行的值
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(nextLine, "<string>") && strings.HasSuffix(nextLine, "</string>") {
					memStr := strings.TrimPrefix(nextLine, "<string>")
					memStr = strings.TrimSuffix(memStr, "</string>")
					if mem := extractMemorySize(memStr); mem > 0 {
						return mem
					}
				}
			}
		}
	}

	return 0
}

// extractMemorySize 从字符串中提取内存大小
func extractMemorySize(text string) uint64 {
	// 移除多余的空格
	text = strings.TrimSpace(text)

	// 查找数字和单位
	words := strings.Fields(text)
	for i, word := range words {
		// 检查是否包含数字
		if hasDigit(word) {
			// 尝试解析数字
			numStr := extractNumber(word)
			if numStr == "" {
				continue
			}

			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				continue
			}

			// 查找单位
			unit := ""
			if strings.Contains(word, "GB") || strings.Contains(word, "gb") {
				unit = "GB"
			} else if strings.Contains(word, "MB") || strings.Contains(word, "mb") {
				unit = "MB"
			} else if i+1 < len(words) {
				nextWord := strings.ToUpper(words[i+1])
				if nextWord == "GB" || nextWord == "MB" || nextWord == "KB" {
					unit = nextWord
				}
			}

			// 转换为字节
			switch unit {
			case "GB":
				return uint64(num * 1024 * 1024 * 1024)
			case "MB":
				return uint64(num * 1024 * 1024)
			case "KB":
				return uint64(num * 1024)
			}
		}
	}

	return 0
}

// hasDigit 检查字符串是否包含数字
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// extractNumber 从字符串中提取数字部分
func extractNumber(s string) string {
	var result strings.Builder

	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// getMacOSGPUInfo 获取macOS GPU详细信息
func getMacOSGPUInfo() ([]GPUInfo, error) {
	var gpus []GPUInfo

	// 首先尝试检测Apple Silicon GPU
	if runtime.GOARCH == "arm64" {
		appleGPU := getAppleSiliconGPUInfo()
		if appleGPU.MemoryTotal > 0 {
			gpus = append(gpus, appleGPU)
		}
	}

	// 使用system_profiler获取显卡信息（用于检测独立显卡或Intel Mac）
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err == nil {
		additionalGPUs := parseMacGPUInfo(string(output))
		for _, gpu := range additionalGPUs {
			// 避免重复添加相同的GPU
			isDuplicate := false
			for _, existingGPU := range gpus {
				if strings.Contains(existingGPU.Name, "Apple") && strings.Contains(gpu.Name, "Apple") {
					isDuplicate = true
					break
				}
			}
			if !isDuplicate {
				gpus = append(gpus, gpu)
			}
		}
	}

	// 如果仍然没有找到GPU，添加默认的集成显卡
	if len(gpus) == 0 {
		totalMem := getSystemProfilerVRAM()
		var usedMem uint64
		if totalMem == 0 {
			// 对于 Apple Silicon，尝试获取统一内存的一部分
			totalMem, usedMem = getAppleSiliconGPUMemory()
		}

		gpus = append(gpus, GPUInfo{
			Name:        "Integrated Graphics",
			MemoryTotal: totalMem,
			MemoryUsed:  usedMem,
			MemoryFree:  totalMem - usedMem,
			Utilization: getMacOSGPUUtilization(), // 尝试获取实际使用率
			Temperature: 0,
		})
	}

	return gpus, nil
}

// getAppleSiliconGPUInfo 获取Apple Silicon GPU信息
func getAppleSiliconGPUInfo() GPUInfo {
	// 获取CPU型号来确定GPU型号
	cpuName := getAppleSiliconCPUName()
	gpuName := mapCPUToGPU(cpuName)

	// 获取内存信息
	totalMem, usedMem := getAppleSiliconGPUMemory()

	// 获取使用率
	utilization := getMacOSGPUUtilization()
	if utilization < 0 {
		utilization = 0 // 如果无法获取，设为0而不是-1
	}

	return GPUInfo{
		Name:        gpuName,
		MemoryTotal: totalMem,
		MemoryUsed:  usedMem,
		MemoryFree:  totalMem - usedMem,
		Utilization: utilization,
		Temperature: 0, // Apple Silicon温度获取需要特殊权限
	}
}

// getAppleSiliconCPUName 获取Apple Silicon CPU名称
func getAppleSiliconCPUName() string {
	cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string")
	output, err := cmd.Output()
	if err != nil {
		return "Apple Silicon"
	}
	return strings.TrimSpace(string(output))
}

// mapCPUToGPU 根据CPU型号映射对应的GPU名称
func mapCPUToGPU(cpuName string) string {
	cpuName = strings.ToLower(cpuName)

	// Apple Silicon GPU映射
	if strings.Contains(cpuName, "m1 pro") {
		return "Apple M1 Pro GPU"
	} else if strings.Contains(cpuName, "m1 max") {
		return "Apple M1 Max GPU"
	} else if strings.Contains(cpuName, "m1 ultra") {
		return "Apple M1 Ultra GPU"
	} else if strings.Contains(cpuName, "m1") {
		return "Apple M1 GPU"
	} else if strings.Contains(cpuName, "m2 pro") {
		return "Apple M2 Pro GPU"
	} else if strings.Contains(cpuName, "m2 max") {
		return "Apple M2 Max GPU"
	} else if strings.Contains(cpuName, "m2 ultra") {
		return "Apple M2 Ultra GPU"
	} else if strings.Contains(cpuName, "m2") {
		return "Apple M2 GPU"
	} else if strings.Contains(cpuName, "m3 pro") {
		return "Apple M3 Pro GPU"
	} else if strings.Contains(cpuName, "m3 max") {
		return "Apple M3 Max GPU"
	} else if strings.Contains(cpuName, "m3") {
		return "Apple M3 GPU"
	} else if strings.Contains(cpuName, "m4 pro") {
		return "Apple M4 Pro GPU"
	} else if strings.Contains(cpuName, "m4 max") {
		return "Apple M4 Max GPU"
	} else if strings.Contains(cpuName, "m4") {
		return "Apple M4 GPU"
	}

	return "Apple Silicon GPU"
}

// parseMacGPUInfo 解析macOS GPU信息
func parseMacGPUInfo(output string) []GPUInfo {
	var gpus []GPUInfo

	lines := strings.Split(output, "\n")
	var currentGPU *GPUInfo
	inDisplaysSection := false // 标记是否在 Displays 部分

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否进入 Displays 部分
		if strings.HasPrefix(line, "Displays:") {
			inDisplaysSection = true
			continue
		}

		// 如果在 Displays 部分，跳过解析
		if inDisplaysSection {
			// 检查是否离开 Displays 部分
			if line == "" {
				inDisplaysSection = false
			}
			continue
		}

		// 检查是否是新的 GPU 条目
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "Displays") {
			// 保存之前的 GPU
			if currentGPU != nil {
				gpus = append(gpus, *currentGPU)
			}

			// 创建新的 GPU
			gpuName := strings.TrimSuffix(line, ":")
			currentGPU = &GPUInfo{
				Name:        gpuName,
				Utilization: getMacOSGPUUtilization(), // 尝试获取实际使用率
				Temperature: 0,                        // macOS 需要特殊权限获取温度
			}
		} else if currentGPU != nil {
			// 解析 GPU 属性
			if strings.Contains(line, "VRAM") || strings.Contains(line, "Memory") {
				if mem := extractMemorySize(line); mem > 0 {
					currentGPU.MemoryTotal = mem
					currentGPU.MemoryFree = mem // 假设空闲内存等于总内存
				}
			}
		}
	}

	// 添加最后一个 GPU
	if currentGPU != nil {
		gpus = append(gpus, *currentGPU)
	}

	// 如果没有找到 GPU，添加一个默认的集成显卡
	if len(gpus) == 0 {
		totalMem := getSystemProfilerVRAM()
		var usedMem uint64
		if totalMem == 0 {
			// 对于 Apple Silicon，尝试获取统一内存的一部分
			totalMem, usedMem = getAppleSiliconGPUMemory()
		}

		gpus = append(gpus, GPUInfo{
			Name:        "Integrated Graphics",
			MemoryTotal: totalMem,
			MemoryUsed:  usedMem,
			MemoryFree:  totalMem - usedMem,
			Utilization: getMacOSGPUUtilization(), // 尝试获取实际使用率
			Temperature: 0,
		})
	}

	return gpus
}

// getTemperature 尝试获取macOS GPU温度（需要特殊权限）
func getTemperature() float64 {
	// macOS获取硬件温度需要特殊权限，通常需要第三方工具
	// 这里返回0表示无法获取
	return 0
}
