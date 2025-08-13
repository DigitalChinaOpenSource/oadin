//go:build windows

package hardware

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
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

	// GPU信息缓存
	gpuCacheMutex sync.RWMutex
	gpuCacheData  []GPUInfo
	gpuCacheTime  time.Time
	gpuCacheTTL   = 10 * time.Second // 缓存10秒
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

// getVRAMInfo 获取Windows VRAM信息 - 优化版本
func getVRAMInfo() (total, used, free uint64) {
	// 使用并发方式同时尝试多种方法，取最快返回的结果
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type result struct {
		total, used uint64
	}

	resultChan := make(chan result, 3)

	// 并发执行三种方法
	go func() {
		if nvTotal, nvUsed := getNvidiaVRAMWithTimeout(ctx); nvTotal > 0 {
			resultChan <- result{nvTotal, nvUsed}
		}
	}()

	go func() {
		if wmicTotal, wmicUsed := getWMICVRAMWithTimeout(ctx); wmicTotal > 0 {
			resultChan <- result{wmicTotal, wmicUsed}
		}
	}()

	go func() {
		if psTotal, psUsed := getPowerShellVRAMWithTimeout(ctx); psTotal > 0 {
			resultChan <- result{psTotal, psUsed}
		}
	}()

	// 返回第一个成功的结果
	select {
	case res := <-resultChan:
		return res.total, res.used, res.total - res.used
	case <-ctx.Done():
		return 0, 0, 0
	}
}

// getNvidiaVRAMWithTimeout 使用nvidia-smi获取NVIDIA GPU显存信息（带超时）
func getNvidiaVRAMWithTimeout(ctx context.Context) (total, used uint64) {
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=memory.total,memory.used", "--format=csv,noheader,nounits")
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

// getWMICVRAMWithTimeout 使用WMIC获取显卡显存信息（带超时）
func getWMICVRAMWithTimeout(ctx context.Context) (total, used uint64) {
	cmd := exec.CommandContext(ctx, "wmic", "path", "win32_VideoController", "get", "AdapterRAM", "/format:value")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "AdapterRAM=") {
			ramStr := strings.TrimPrefix(line, "AdapterRAM=")
			if ramStr != "" && ramStr != "0" {
				if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil && ram > 0 {
					return ram, 0
				}
			}
		}
	}

	return 0, 0
}

// getPowerShellVRAMWithTimeout 使用PowerShell获取显存信息（带超时）
func getPowerShellVRAMWithTimeout(ctx context.Context) (total, used uint64) {
	cmd := exec.CommandContext(ctx, "powershell", "-Command",
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

// getWindowsGPUInfo 获取Windows GPU详细信息 - 优化版本（带缓存）
func getWindowsGPUInfo() ([]GPUInfo, error) {
	// 检查缓存
	gpuCacheMutex.RLock()
	if time.Since(gpuCacheTime) < gpuCacheTTL && len(gpuCacheData) > 0 {
		cached := make([]GPUInfo, len(gpuCacheData))
		copy(cached, gpuCacheData)
		gpuCacheMutex.RUnlock()
		return cached, nil
	}
	gpuCacheMutex.RUnlock()

	// 缓存过期或为空，重新获取
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 减少超时时间
	defer cancel()

	var gpus []GPUInfo
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 并发尝试不同的方法
	wg.Add(2)

	// 尝试nvidia-smi
	go func() {
		defer wg.Done()
		if nvGPUs := getNvidiaGPUInfoWithTimeout(ctx); len(nvGPUs) > 0 {
			mu.Lock()
			gpus = append(gpus, nvGPUs...)
			mu.Unlock()
		}
	}()

	// 尝试通用方法
	go func() {
		defer wg.Done()
		if genericGPUs := getGenericWindowsGPUInfoWithTimeout(ctx); len(genericGPUs) > 0 {
			mu.Lock()
			if len(gpus) == 0 { // 只有在没有nvidia GPU时才添加通用GPU
				gpus = append(gpus, genericGPUs...)
			}
			mu.Unlock()
		}
	}()

	// 等待所有goroutine完成，但不超过超时时间
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 更新缓存
		gpuCacheMutex.Lock()
		gpuCacheData = make([]GPUInfo, len(gpus))
		copy(gpuCacheData, gpus)
		gpuCacheTime = time.Now()
		gpuCacheMutex.Unlock()

		return gpus, nil
	case <-ctx.Done():
		// 如果超时，返回缓存的数据（如果有的话）
		gpuCacheMutex.RLock()
		if len(gpuCacheData) > 0 {
			cached := make([]GPUInfo, len(gpuCacheData))
			copy(cached, gpuCacheData)
			gpuCacheMutex.RUnlock()
			return cached, fmt.Errorf("timeout getting GPU info, returned cached data")
		}
		gpuCacheMutex.RUnlock()
		return nil, fmt.Errorf("timeout getting GPU info")
	}
}

// getNvidiaGPUInfoWithTimeout 获取NVIDIA GPU详细信息（带超时）
func getNvidiaGPUInfoWithTimeout(ctx context.Context) []GPUInfo {
	cmd := exec.CommandContext(ctx, "nvidia-smi",
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

// getGenericWindowsGPUInfoWithTimeout 获取通用Windows GPU信息（带超时，优化版本）
func getGenericWindowsGPUInfoWithTimeout(ctx context.Context) []GPUInfo {
	// 使用一次PowerShell调用获取所有GPU信息，而不是每个GPU单独调用
	cmd := exec.CommandContext(ctx, "powershell", "-Command", `
		$gpus = Get-WmiObject -Class Win32_VideoController | Where-Object {$_.Name -notlike '*RDP*' -and $_.Name -notlike '*Oray*' -and $_.AdapterRAM -gt 0}
		foreach ($gpu in $gpus) {
			$name = $gpu.Name
			$ram = $gpu.AdapterRAM
			if ($ram -eq $null) { $ram = 0 }
			Write-Output "$name|$ram"
		}
	`)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		// 如果PowerShell失败，尝试WMIC作为备用
		return getWMICGPUInfoWithTimeout(ctx)
	}

	var gpus []GPUInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		ramStr := strings.TrimSpace(parts[1])

		if name == "" {
			continue
		}

		var memTotal uint64 = 0
		if ramStr != "" && ramStr != "0" {
			if ram, err := strconv.ParseUint(ramStr, 10, 64); err == nil {
				memTotal = ram
			}
		}

		// 估算内存使用情况（避免额外的PowerShell调用）
		var memUsed uint64 = 0
		if memTotal > 0 {
			memUsed = memTotal / 10 // 假设使用10%
		}

		gpu := GPUInfo{
			Name:        name,
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			MemoryFree:  memTotal - memUsed,
			Utilization: -1, // 避免额外的调用来获取使用率
			Temperature: 0,
		}

		gpus = append(gpus, gpu)
	}

	return gpus
}

// getWMICGPUInfoWithTimeout WMIC备用方法（带超时）
func getWMICGPUInfoWithTimeout(ctx context.Context) []GPUInfo {
	cmd := exec.CommandContext(ctx, "wmic", "path", "win32_VideoController", "get", "Name,AdapterRAM", "/format:csv")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	return parseWMICGPUInfo(string(output))
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

// getGPUUtilization 获取GPU使用率
// Deprecated: 这个函数被认为是性能瓶颈，每次调用都会执行多个PowerShell命令
// 建议使用nvidia-smi或缓存机制来获取GPU使用率
func getGPUUtilization(gpuName string) float64 {
	// 快速返回估算值，避免性能瓶颈
	return -1 // 表示无法获取，避免调用慢的PowerShell命令
}

// getGPUMemoryUsage 获取GPU内存使用情况
// Deprecated: 这个函数被认为是性能瓶颈，每次调用都会执行PowerShell命令
// 建议使用nvidia-smi或估算方法
func getGPUMemoryUsage(gpuName string, totalMemory uint64) uint64 {
	// 快速返回估算值，避免性能瓶颈
	if totalMemory > 0 {
		return totalMemory / 10 // 估算使用10%
	}
	return 0
}

// getPerformanceCounterGPUUsage 使用性能计数器获取GPU使用率
// Deprecated: 性能瓶颈函数 - typeperf命令执行慢，建议使用nvidia-smi或估算
func getPerformanceCounterGPUUsage(gpuName string) float64 {
	return -1 // 直接返回，避免执行慢的typeperf命令
}

// getPowerShellGPUUsage 使用PowerShell获取GPU使用率
// Deprecated: 性能瓶颈函数 - PowerShell Get-Counter命令执行慢
func getPowerShellGPUUsage(gpuName string) float64 {
	return -1 // 直接返回，避免执行慢的PowerShell命令
}

// getWMIGPUUsage 使用WMI获取GPU使用率
// Deprecated: 性能瓶颈函数 - 复杂的进程查询执行慢
func getWMIGPUUsage() float64 {
	return -1 // 直接返回，避免执行慢的进程查询
}

// getTaskManagerGPUMemory 使用类似任务管理器的方式获取GPU内存使用
// Deprecated: 性能瓶颈函数 - PowerShell Get-Counter命令执行慢
func getTaskManagerGPUMemory() uint64 {
	return 0 // 直接返回，避免执行慢的PowerShell命令
}

// getIntelGPUMemoryUsage 获取Intel GPU内存使用情况
// Deprecated: 性能瓶颈函数 - 复杂的进程查询执行慢
func getIntelGPUMemoryUsage() uint64 {
	return 0 // 直接返回，避免执行慢的进程查询
}
