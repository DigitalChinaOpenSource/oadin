//go:build linux

package hardware

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// getWindowsGPUInfo 在Linux平台返回错误
func getWindowsGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Windows GPU info not available on Linux")
}

// getMacOSGPUInfo 在Linux平台返回错误
func getMacOSGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("macOS GPU info not available on Linux")
}

// getVRAMInfo 获取Linux VRAM信息
func getVRAMInfo() (total, used, free uint64) {
	// 优先使用nvidia-smi
	if nvTotal, nvUsed := getNvidiaVRAMLinux(); nvTotal > 0 {
		return nvTotal, nvUsed, nvTotal - nvUsed
	}

	// 尝试AMD GPU
	if amdTotal, amdUsed := getAMDVRAMLinux(); amdTotal > 0 {
		return amdTotal, amdUsed, amdTotal - amdUsed
	}

	// 尝试Intel GPU
	if intelTotal, intelUsed := getIntelVRAMLinux(); intelTotal > 0 {
		return intelTotal, intelUsed, intelTotal - intelUsed
	}

	return 0, 0, 0
}

// getNvidiaVRAMLinux 获取NVIDIA GPU显存信息
func getNvidiaVRAMLinux() (total, used uint64) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.total,memory.used", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 {
		return 0, 0
	}

	// 处理第一个GPU的信息
	parts := strings.Split(strings.TrimSpace(lines[0]), ",")
	if len(parts) != 2 {
		return 0, 0
	}

	totalMB, err1 := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
	usedMB, err2 := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)

	if err1 != nil || err2 != nil {
		return 0, 0
	}

	// 转换为字节
	return totalMB * 1024 * 1024, usedMB * 1024 * 1024
}

// getAMDVRAMLinux 获取AMD GPU显存信息
func getAMDVRAMLinux() (total, used uint64) {
	// 尝试读取 /sys/class/drm/cardX/device/mem_info_vram_total
	cards, err := filepath.Glob("/sys/class/drm/card*/device/mem_info_vram_total")
	if err != nil || len(cards) == 0 {
		return 0, 0
	}

	for _, card := range cards {
		if totalBytes := readSysfsValue(card); totalBytes > 0 {
			usedPath := strings.Replace(card, "vram_total", "vram_used", 1)
			usedBytes := readSysfsValue(usedPath)
			return totalBytes, usedBytes
		}
	}

	return 0, 0
}

// getIntelVRAMLinux 获取Intel GPU显存信息
func getIntelVRAMLinux() (total, used uint64) {
	// Intel GPU通常使用系统内存，检查i915驱动信息
	if totalBytes := readIntelGpuMemory(); totalBytes > 0 {
		return totalBytes, 0 // Intel GPU使用量难以准确获取
	}

	return 0, 0
}

// readSysfsValue 读取sysfs值
func readSysfsValue(path string) uint64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0
	}

	return value
}

// readIntelGpuMemory 读取Intel GPU内存信息
func readIntelGpuMemory() uint64 {
	// 尝试从debugfs读取
	debugPaths := []string{
		"/sys/kernel/debug/dri/0/i915_gem_objects",
		"/proc/driver/i915/gem_objects",
	}

	for _, path := range debugPaths {
		if mem := parseIntelDebugInfo(path); mem > 0 {
			return mem
		}
	}

	return 0
}

// parseIntelDebugInfo 解析Intel GPU调试信息
func parseIntelDebugInfo(path string) uint64 {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "total") && strings.Contains(line, "bytes") {
			// 尝试解析总内存大小
			fields := strings.Fields(line)
			for i, field := range fields {
				if field == "bytes" && i > 0 {
					if size, err := strconv.ParseUint(fields[i-1], 10, 64); err == nil {
						return size
					}
				}
			}
		}
	}

	return 0
}

// getLinuxGPUInfo 获取Linux GPU详细信息
func getLinuxGPUInfo() ([]GPUInfo, error) {
	var gpus []GPUInfo

	// 尝试NVIDIA GPU
	if nvGPUs := getNvidiaGPUInfoLinux(); len(nvGPUs) > 0 {
		gpus = append(gpus, nvGPUs...)
	}

	// 尝试AMD GPU
	if amdGPUs := getAMDGPUInfoLinux(); len(amdGPUs) > 0 {
		gpus = append(gpus, amdGPUs...)
	}

	// 尝试Intel GPU
	if intelGPUs := getIntelGPUInfoLinux(); len(intelGPUs) > 0 {
		gpus = append(gpus, intelGPUs...)
	}

	return gpus, nil
}

// getNvidiaGPUInfoLinux 获取NVIDIA GPU详细信息
func getNvidiaGPUInfoLinux() []GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,memory.total,memory.used,utilization.gpu,temperature.gpu",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) != 5 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		totalMB, _ := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		usedMB, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		util, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		temp, _ := strconv.ParseFloat(strings.TrimSpace(parts[4]), 64)

		total := totalMB * 1024 * 1024
		used := usedMB * 1024 * 1024

		gpu := GPUInfo{
			Name:        name,
			MemoryTotal: total,
			MemoryUsed:  used,
			MemoryFree:  total - used,
			Utilization: util,
			Temperature: temp,
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// getAMDGPUInfoLinux 获取AMD GPU详细信息
func getAMDGPUInfoLinux() []GPUInfo {
	var gpus []GPUInfo

	// 查找AMD GPU设备
	cards, err := filepath.Glob("/sys/class/drm/card*/device/vendor")
	if err != nil {
		return nil
	}

	for _, vendorPath := range cards {
		vendor := readSysfsStringValue(vendorPath)
		// AMD vendor ID: 0x1002
		if vendor != "0x1002" {
			continue
		}

		cardDir := filepath.Dir(vendorPath)

		// 读取设备名称
		devicePath := filepath.Join(cardDir, "device")
		deviceID := readSysfsStringValue(devicePath)

		// 读取内存信息
		vramTotalPath := filepath.Join(cardDir, "mem_info_vram_total")
		vramUsedPath := filepath.Join(cardDir, "mem_info_vram_used")

		total := readSysfsValue(vramTotalPath)
		used := readSysfsValue(vramUsedPath)

		if total > 0 {
			gpu := GPUInfo{
				Name:        fmt.Sprintf("AMD GPU (Device: %s)", deviceID),
				MemoryTotal: total,
				MemoryUsed:  used,
				MemoryFree:  total - used,
				Utilization: -1, // AMD GPU使用率需要特殊获取
				Temperature: getAMDTemperature(cardDir),
			}

			gpus = append(gpus, gpu)
		}
	}

	return gpus
}

// getIntelGPUInfoLinux 获取Intel GPU详细信息
func getIntelGPUInfoLinux() []GPUInfo {
	var gpus []GPUInfo

	// 查找Intel GPU设备
	cards, err := filepath.Glob("/sys/class/drm/card*/device/vendor")
	if err != nil {
		return nil
	}

	for _, vendorPath := range cards {
		vendor := readSysfsStringValue(vendorPath)
		// Intel vendor ID: 0x8086
		if vendor != "0x8086" {
			continue
		}

		cardDir := filepath.Dir(vendorPath)

		// 读取设备名称
		devicePath := filepath.Join(cardDir, "device")
		deviceID := readSysfsStringValue(devicePath)

		// Intel GPU通常使用系统内存
		memTotal := readIntelGpuMemory()

		if memTotal > 0 {
			gpu := GPUInfo{
				Name:        fmt.Sprintf("Intel GPU (Device: %s)", deviceID),
				MemoryTotal: memTotal,
				MemoryUsed:  0, // Intel GPU使用量难以准确获取
				MemoryFree:  memTotal,
				Utilization: -1,
				Temperature: 0,
			}

			gpus = append(gpus, gpu)
		}
	}

	return gpus
}

// readSysfsStringValue 读取sysfs字符串值
func readSysfsStringValue(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// getAMDTemperature 获取AMD GPU温度
func getAMDTemperature(cardDir string) float64 {
	tempPaths := []string{
		"hwmon/hwmon*/temp1_input",
		"hwmon/hwmon*/temp2_input",
	}

	for _, tempPattern := range tempPaths {
		tempPath := filepath.Join(cardDir, tempPattern)
		matches, err := filepath.Glob(tempPath)
		if err != nil || len(matches) == 0 {
			continue
		}

		for _, match := range matches {
			if temp := readSysfsValue(match); temp > 0 {
				// 温度值通常以毫摄氏度为单位
				return float64(temp) / 1000.0
			}
		}
	}

	return 0
}
