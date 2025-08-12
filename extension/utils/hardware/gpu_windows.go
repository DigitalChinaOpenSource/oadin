//go:build windows

package hardware

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Windows GPU信息获取

// getLinuxGPUInfo 在Windows平台返回错误
func getLinuxGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Linux GPU info not available on Windows")
}

// getMacOSGPUInfo 在Windows平台返回错误
func getMacOSGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("macOS GPU info not available on Windows")
}

var (
	modKernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = modKernel32.NewProc("GlobalMemoryStatusEx")
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

// getVRAMInfo 获取Windows VRAM信息
func getVRAMInfo() (total, used, free uint64) {
	// 使用nvidia-smi获取NVIDIA GPU信息
	if nvTotal, nvUsed := getNvidiaVRAM(); nvTotal > 0 {
		return nvTotal, nvUsed, nvTotal - nvUsed
	}

	// 使用WMIC获取显卡信息
	if wmicTotal, wmicUsed := getWMICVRAM(); wmicTotal > 0 {
		return wmicTotal, wmicUsed, wmicTotal - wmicUsed
	}

	return 0, 0, 0
}

// getNvidiaVRAM 使用nvidia-smi获取NVIDIA GPU显存信息
func getNvidiaVRAM() (total, used uint64) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.total,memory.used", "--format=csv,noheader,nounits")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

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

// getWMICVRAM 使用PowerShell获取显卡显存信息
func getWMICVRAM() (total, used uint64) {
	// 首先尝试传统的wmic命令
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "AdapterRAM", "/format:value")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		// 如果wmic失败，尝试使用PowerShell
		return getPowerShellVRAM()
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "AdapterRAM=") {
			ramStr := strings.TrimPrefix(line, "AdapterRAM=")
			if ramStr != "" && ramStr != "0" {
				if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil && ram > 0 {
					// WMIC返回的是字节数，我们假设使用率为未知，返回0
					return ram, 0
				}
			}
		}
	}

	return 0, 0
}

// getPowerShellVRAM 使用PowerShell获取显存信息
func getPowerShellVRAM() (total, used uint64) {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject -Class Win32_VideoController | Where-Object {$_.AdapterRAM -gt 0} | Select-Object -First 1 -ExpandProperty AdapterRAM")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	ramStr := strings.TrimSpace(string(output))
	if ramStr != "" && ramStr != "0" {
		if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil && ram > 0 {
			return ram, 0
		}
	}

	return 0, 0
}

// getWindowsGPUInfo 获取Windows GPU详细信息
func getWindowsGPUInfo() ([]GPUInfo, error) {
	var gpus []GPUInfo

	// 尝试使用nvidia-smi获取NVIDIA GPU信息
	if nvGPUs := getNvidiaGPUInfo(); len(nvGPUs) > 0 {
		gpus = append(gpus, nvGPUs...)
	}

	// 如果没有NVIDIA GPU，尝试获取通用GPU信息
	if len(gpus) == 0 {
		if genericGPUs := getGenericWindowsGPUInfo(); len(genericGPUs) > 0 {
			gpus = append(gpus, genericGPUs...)
		}
	}

	return gpus, nil
}

// getNvidiaGPUInfo 获取NVIDIA GPU详细信息
func getNvidiaGPUInfo() []GPUInfo {
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=name,memory.total,memory.used,utilization.gpu,temperature.gpu",
		"--format=csv,noheader,nounits")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

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

// getGenericWindowsGPUInfo 获取通用Windows GPU信息
func getGenericWindowsGPUInfo() []GPUInfo {
	// 首先尝试传统的wmic命令
	cmd := exec.Command("wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM", "/format:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		// 如果wmic失败，尝试使用PowerShell
		return getPowerShellGPUInfo()
	}

	return parseWMICGPUInfo(string(output))
}

// getPowerShellGPUInfo 使用PowerShell获取GPU信息
func getPowerShellGPUInfo() []GPUInfo {
	cmd := exec.Command("powershell", "-Command",
		"Get-WmiObject -Class Win32_VideoController | Where-Object {$_.Name -notlike '*RDP*' -and $_.Name -notlike '*Oray*'} | Select-Object Name, AdapterRAM | Format-Table -HideTableHeaders")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parsePowerShellGPUInfo(string(output))
}

// parseWMICGPUInfo 解析WMIC输出的GPU信息
func parseWMICGPUInfo(output string) []GPUInfo {
	var gpus []GPUInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines[1:] { // 跳过标题行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		name := strings.TrimSpace(parts[2])
		ramStr := strings.TrimSpace(parts[1])

		if name == "Name" || name == "" {
			continue
		}

		var memTotal uint64 = 0
		if ramStr != "" && ramStr != "0" {
			if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil {
				memTotal = ram
			}
		}

		gpu := GPUInfo{
			Name:        name,
			MemoryTotal: memTotal,
			MemoryUsed:  0, // Windows通用接口无法获取使用量
			MemoryFree:  memTotal,
			Utilization: -1, // 表示未知
			Temperature: 0,
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// parsePowerShellGPUInfo 解析PowerShell输出的GPU信息
func parsePowerShellGPUInfo(output string) []GPUInfo {
	var gpus []GPUInfo
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// PowerShell输出格式: "Name                             AdapterRAM"
		// 使用空格分割可能不可靠，尝试更智能的解析
		if strings.Contains(line, "GPU") || strings.Contains(line, "Graphics") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// 最后一个字段应该是AdapterRAM
				ramStr := parts[len(parts)-1]

				// 名称是除了最后一个字段之外的所有部分
				name := strings.Join(parts[:len(parts)-1], " ")

				var memTotal uint64 = 0
				if ramStr != "" && ramStr != "0" {
					if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil {
						memTotal = ram
					}
				}

				// 过滤掉无效的GPU名称
				if !strings.Contains(name, "RDP") && !strings.Contains(name, "Oray") {
					gpu := GPUInfo{
						Name:        name,
						MemoryTotal: memTotal,
						MemoryUsed:  0, // Windows通用接口无法获取使用量
						MemoryFree:  memTotal,
						Utilization: -1, // 表示未知
						Temperature: 0,
					}

					gpus = append(gpus, gpu)
				}
			}
		}
	}

	return gpus
}
