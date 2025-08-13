package hardware

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// MemoryInfo 内存信息结构体
type MemoryInfo struct {
	// RAM 信息
	RAMTotal   uint64  `json:"ram_total"`    // 总内存 (bytes)
	RAMUsed    uint64  `json:"ram_used"`     // 已使用内存 (bytes)
	RAMFree    uint64  `json:"ram_free"`     // 可用内存 (bytes)
	RAMUsedPct float64 `json:"ram_used_pct"` // 内存使用率 (%)

	// VRAM 信息 (显存)
	VRAMTotal   uint64  `json:"vram_total"`    // 总显存 (bytes)
	VRAMUsed    uint64  `json:"vram_used"`     // 已使用显存 (bytes)
	VRAMFree    uint64  `json:"vram_free"`     // 可用显存 (bytes)
	VRAMUsedPct float64 `json:"vram_used_pct"` // 显存使用率 (%)

	// 系统信息
	Platform   string `json:"platform"`    // 操作系统平台
	Hostname   string `json:"hostname"`    // 主机名
	UpdateTime string `json:"update_time"` // 更新时间
}

// GPUInfo GPU信息结构体
type GPUInfo struct {
	Name        string  `json:"name"`         // GPU名称
	MemoryTotal uint64  `json:"memory_total"` // 总显存 (bytes)
	MemoryUsed  uint64  `json:"memory_used"`  // 已使用显存 (bytes)
	MemoryFree  uint64  `json:"memory_free"`  // 可用显存 (bytes)
	Utilization float64 `json:"utilization"`  // GPU使用率 (%)
	Temperature float64 `json:"temperature"`  // GPU温度 (°C)
}

// CPUInfo CPU信息结构体
type CPUInfo struct {
	ModelName    string   `json:"model_name"`    // CPU型号名称
	Brand        string   `json:"brand"`         // CPU品牌
	Architecture string   `json:"architecture"`  // CPU架构 (x86_64, arm64, etc.)
	Cores        int32    `json:"cores"`         // 物理核心数
	Threads      int32    `json:"threads"`       // 逻辑核心数 (线程数)
	MaxFrequency float64  `json:"max_frequency"` // 最大频率 (MHz)
	CurrentUsage float64  `json:"current_usage"` // 当前使用率 (%)
	Features     []string `json:"features"`      // 支持的指令集/特性
	CacheSize    int32    `json:"cache_size"`    // 缓存大小 (KB)
	Family       string   `json:"family"`        // CPU系列
	Model        string   `json:"model"`         // CPU型号
	Stepping     string   `json:"stepping"`      // CPU步进
	Microcode    string   `json:"microcode"`     // 微码版本
}

// SystemHardwareInfo 系统硬件综合信息
type SystemHardwareInfo struct {
	CPU    *CPUInfo    `json:"cpu"`    // CPU信息
	Memory *MemoryInfo `json:"memory"` // 内存信息
	GPUs   []GPUInfo   `json:"gpus"`   // GPU信息列表
}

// GetMemoryInfo 获取内存信息
func GetMemoryInfo() (*MemoryInfo, error) {
	// 获取RAM信息
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory info: %w", err)
	}

	// 获取系统信息
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	// 获取VRAM信息
	vramTotal, vramUsed, vramFree := getVRAMInfo()
	vramUsedPct := 0.0
	if vramTotal > 0 {
		vramUsedPct = float64(vramUsed) / float64(vramTotal) * 100
	}

	memInfo := &MemoryInfo{
		// RAM信息
		RAMTotal:   vmStat.Total,
		RAMUsed:    vmStat.Used,
		RAMFree:    vmStat.Free,
		RAMUsedPct: vmStat.UsedPercent,

		// VRAM信息
		VRAMTotal:   vramTotal,
		VRAMUsed:    vramUsed,
		VRAMFree:    vramFree,
		VRAMUsedPct: vramUsedPct,

		// 系统信息
		Platform:   hostInfo.Platform,
		Hostname:   hostInfo.Hostname,
		UpdateTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	return memInfo, nil
}

// GetCPUInfo 获取CPU信息
func GetCPUInfo() (*CPUInfo, error) {
	// 获取CPU基本信息
	cpuInfos, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	if len(cpuInfos) == 0 {
		return nil, fmt.Errorf("no CPU information available")
	}

	// 获取第一个CPU的信息（通常所有核心信息相同）
	cpuInfo := cpuInfos[0]

	// 获取逻辑核心数
	logicalCores, err := cpu.Counts(true)
	if err != nil {
		logicalCores = int(cpuInfo.Cores) // 降级处理
	}

	// 获取物理核心数
	physicalCores, err := cpu.Counts(false)
	if err != nil {
		physicalCores = int(cpuInfo.Cores) // 降级处理
	}

	// 获取当前CPU使用率
	cpuPercent, err := cpu.Percent(time.Second, false)
	var currentUsage float64
	if err == nil && len(cpuPercent) > 0 {
		currentUsage = cpuPercent[0]
	}

	// 构建CPU信息结构
	result := &CPUInfo{
		ModelName:    cpuInfo.ModelName,
		Brand:        cpuInfo.VendorID,
		Architecture: runtime.GOARCH,
		Cores:        int32(physicalCores),
		Threads:      int32(logicalCores),
		MaxFrequency: cpuInfo.Mhz,
		CurrentUsage: currentUsage,
		Features:     cpuInfo.Flags,
		CacheSize:    cpuInfo.CacheSize,
		Family:       cpuInfo.Family,
		Model:        cpuInfo.Model,
		Stepping:     fmt.Sprintf("%d", cpuInfo.Stepping),
		Microcode:    cpuInfo.Microcode,
	}

	return result, nil
}

// GetSystemHardwareInfo 获取完整的系统硬件信息
func GetSystemHardwareInfo() (*SystemHardwareInfo, error) {
	// 获取CPU信息
	cpuInfo, err := GetCPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	// 获取内存信息
	memoryInfo, err := GetMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// 获取GPU信息
	gpuInfo, err := GetGPUInfo()
	if err != nil {
		// GPU信息获取失败不影响整体，记录错误但继续
		gpuInfo = []GPUInfo{}
	}

	return &SystemHardwareInfo{
		CPU:    cpuInfo,
		Memory: memoryInfo,
		GPUs:   gpuInfo,
	}, nil
}

// GetGPUInfo 获取GPU信息（支持多GPU）
func GetGPUInfo() ([]GPUInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return getWindowsGPUInfo()
	case "linux":
		return getLinuxGPUInfo()
	case "darwin":
		return getMacOSGPUInfo()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// FormatBytes 格式化字节数为人类可读格式
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PrintMemoryInfo 打印内存信息（用于调试）
func PrintMemoryInfo() error {
	memInfo, err := GetMemoryInfo()
	if err != nil {
		return err
	}

	fmt.Println("=== 内存信息 ===")
	fmt.Printf("系统: %s (%s)\n", memInfo.Platform, memInfo.Hostname)
	fmt.Printf("更新时间: %s\n", memInfo.UpdateTime)
	fmt.Println()

	fmt.Println("RAM 信息:")
	fmt.Printf("  总容量: %s\n", FormatBytes(memInfo.RAMTotal))
	fmt.Printf("  已使用: %s (%.1f%%)\n", FormatBytes(memInfo.RAMUsed), memInfo.RAMUsedPct)
	fmt.Printf("  可用量: %s\n", FormatBytes(memInfo.RAMFree))
	fmt.Println()

	if memInfo.VRAMTotal > 0 {
		fmt.Println("VRAM 信息:")
		fmt.Printf("  总容量: %s\n", FormatBytes(memInfo.VRAMTotal))
		fmt.Printf("  已使用: %s (%.1f%%)\n", FormatBytes(memInfo.VRAMUsed), memInfo.VRAMUsedPct)
		fmt.Printf("  可用量: %s\n", FormatBytes(memInfo.VRAMFree))
		fmt.Println()
	}

	// 显示GPU详细信息
	gpuInfo, err := GetGPUInfo()
	if err == nil && len(gpuInfo) > 0 {
		fmt.Println("GPU 信息:")
		for i, gpu := range gpuInfo {
			fmt.Printf("  GPU %d: %s\n", i+1, gpu.Name)
			if gpu.MemoryTotal > 0 {
				fmt.Printf("    显存: %s / %s (%.1f%%)\n",
					FormatBytes(gpu.MemoryUsed),
					FormatBytes(gpu.MemoryTotal),
					float64(gpu.MemoryUsed)/float64(gpu.MemoryTotal)*100)
			}
			if gpu.Utilization >= 0 {
				fmt.Printf("    使用率: %.1f%%\n", gpu.Utilization)
			}
			if gpu.Temperature > 0 {
				fmt.Printf("    温度: %.1f°C\n", gpu.Temperature)
			}
		}
	}

	return nil
}

// PrintCPUInfo 打印CPU信息（用于调试）
func PrintCPUInfo() error {
	cpuInfo, err := GetCPUInfo()
	if err != nil {
		return err
	}

	fmt.Println("=== CPU信息 ===")
	fmt.Printf("型号: %s\n", cpuInfo.ModelName)
	fmt.Printf("品牌: %s\n", cpuInfo.Brand)
	fmt.Printf("架构: %s\n", cpuInfo.Architecture)
	fmt.Printf("物理核心: %d 核\n", cpuInfo.Cores)
	fmt.Printf("逻辑核心: %d 线程\n", cpuInfo.Threads)
	fmt.Printf("最大频率: %.0f MHz\n", cpuInfo.MaxFrequency)
	fmt.Printf("当前使用率: %.1f%%\n", cpuInfo.CurrentUsage)

	if cpuInfo.CacheSize > 0 {
		fmt.Printf("缓存大小: %d KB\n", cpuInfo.CacheSize)
	}

	if len(cpuInfo.Features) > 0 {
		fmt.Printf("支持的指令集: ")
		// 只显示前10个重要的指令集特性
		count := 0
		for _, feature := range cpuInfo.Features {
			if count >= 10 {
				fmt.Printf("... (共%d个)", len(cpuInfo.Features))
				break
			}
			if count > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", feature)
			count++
		}
		fmt.Println()
	}
	fmt.Println()

	return nil
}

// PrintSystemHardwareInfo 打印完整的系统硬件信息
func PrintSystemHardwareInfo() error {
	fmt.Println("🖥️  OADIN 系统硬件信息")
	fmt.Println("================================")

	// 打印CPU信息
	if err := PrintCPUInfo(); err != nil {
		fmt.Printf("❌ 获取CPU信息失败: %v\n", err)
	}

	// 打印内存信息
	if err := PrintMemoryInfo(); err != nil {
		fmt.Printf("❌ 获取内存信息失败: %v\n", err)
	}

	return nil
}

func getSystemInformation() {
	// 保持原有函数签名，可以调用 PrintSystemHardwareInfo
	if err := PrintSystemHardwareInfo(); err != nil {
		fmt.Printf("获取系统信息失败: %v\n", err)
	}
}
