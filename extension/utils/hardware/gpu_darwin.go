//go:build darwin

package hardware

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
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
	// macOS主要使用Metal和系统信息工具
	if metalTotal, metalUsed := getMetalVRAM(); metalTotal > 0 {
		return metalTotal, metalUsed, metalTotal - metalUsed
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

	// 使用system_profiler获取显卡信息
	cmd := exec.Command("system_profiler", "SPDisplaysDataType")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	gpus = parseMacGPUInfo(string(output))

	// 如果找到了GPU，尝试获取更详细的信息
	for i := range gpus {
		if gpus[i].MemoryTotal == 0 {
			// 尝试其他方法获取内存信息
			gpus[i].MemoryTotal = getSystemProfilerVRAM()
		}
	}

	return gpus, nil
}

// parseMacGPUInfo 解析macOS GPU信息
func parseMacGPUInfo(output string) []GPUInfo {
	var gpus []GPUInfo

	lines := strings.Split(output, "\n")
	var currentGPU *GPUInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否是新的GPU条目
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "Displays") {
			// 保存之前的GPU
			if currentGPU != nil {
				gpus = append(gpus, *currentGPU)
			}

			// 创建新的GPU
			gpuName := strings.TrimSuffix(line, ":")
			currentGPU = &GPUInfo{
				Name:        gpuName,
				Utilization: -1, // macOS较难获取实时使用率
				Temperature: 0,  // macOS需要特殊权限获取温度
			}
		} else if currentGPU != nil {
			// 解析GPU属性
			if strings.Contains(line, "VRAM") || strings.Contains(line, "Memory") {
				if mem := extractMemorySize(line); mem > 0 {
					currentGPU.MemoryTotal = mem
					currentGPU.MemoryFree = mem // 假设空闲内存等于总内存
				}
			}
		}
	}

	// 添加最后一个GPU
	if currentGPU != nil {
		gpus = append(gpus, *currentGPU)
	}

	// 如果没有找到GPU，添加一个默认的集成显卡
	if len(gpus) == 0 {
		totalMem := getSystemProfilerVRAM()
		if totalMem == 0 {
			// 对于Apple Silicon，尝试获取统一内存的一部分
			totalMem = getAppleSiliconGPUMemory()
		}

		gpus = append(gpus, GPUInfo{
			Name:        "Integrated Graphics",
			MemoryTotal: totalMem,
			MemoryUsed:  0,
			MemoryFree:  totalMem,
			Utilization: -1,
			Temperature: 0,
		})
	}

	return gpus
}

// getAppleSiliconGPUMemory 获取Apple Silicon GPU内存（统一内存架构）
func getAppleSiliconGPUMemory() uint64 {
	// Apple Silicon使用统一内存架构，GPU和CPU共享内存
	// 尝试获取系统总内存的一部分作为GPU可用内存

	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	totalMemStr := strings.TrimSpace(string(output))
	totalMem, err := strconv.ParseUint(totalMemStr, 10, 64)
	if err != nil {
		return 0
	}

	// 假设GPU可以使用总内存的1/4到1/2
	// 这是一个估算值，实际可用内存取决于系统配置
	return totalMem / 3 // 约33%的系统内存用于GPU
}

// getTemperature 尝试获取macOS GPU温度（需要特殊权限）
func getTemperature() float64 {
	// macOS获取硬件温度需要特殊权限，通常需要第三方工具
	// 这里返回0表示无法获取
	return 0
}
