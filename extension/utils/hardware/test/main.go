package main

import (
	"fmt"
	"log"

	"oadin/extension/utils/hardware"
)

func main() {
	fmt.Println("🚀 OADIN 系统硬件监控测试")
	fmt.Println("================================")

	// 测试CPU信息获取
	fmt.Println("\n🔧 获取CPU信息...")
	cpuInfo, err := hardware.GetCPUInfo()
	if err != nil {
		log.Printf("❌ 获取CPU信息失败: %v", err)
	} else {
		fmt.Printf("✅ CPU信息获取成功\n")
		fmt.Printf("   型号: %s\n", cpuInfo.ModelName)
		fmt.Printf("   架构: %s\n", cpuInfo.Architecture)
		fmt.Printf("   核心: %d核/%d线程\n", cpuInfo.Cores, cpuInfo.Threads)
		fmt.Printf("   频率: %.0f MHz\n", cpuInfo.MaxFrequency)
		fmt.Printf("   使用率: %.1f%%\n", cpuInfo.CurrentUsage)
		if len(cpuInfo.Features) > 0 {
			fmt.Printf("   指令集: %s", cpuInfo.Features[0])
			if len(cpuInfo.Features) > 1 {
				fmt.Printf(" (等%d个)", len(cpuInfo.Features))
			}
			fmt.Println()
		}
	}

	// 测试内存信息获取
	fmt.Println("\n📊 获取内存信息...")
	memInfo, err := hardware.GetMemoryInfo()
	if err != nil {
		log.Printf("❌ 获取内存信息失败: %v", err)
	} else {
		fmt.Printf("✅ 内存信息获取成功\n")
		fmt.Printf("   系统: %s (%s)\n", memInfo.Platform, memInfo.Hostname)
		fmt.Printf("   RAM: %s / %s (使用率: %.1f%%)\n",
			hardware.FormatBytes(memInfo.RAMUsed),
			hardware.FormatBytes(memInfo.RAMTotal),
			memInfo.RAMUsedPct)

		if memInfo.VRAMTotal > 0 {
			fmt.Printf("   VRAM: %s / %s (使用率: %.1f%%)\n",
				hardware.FormatBytes(memInfo.VRAMUsed),
				hardware.FormatBytes(memInfo.VRAMTotal),
				memInfo.VRAMUsedPct)
		} else {
			fmt.Printf("   VRAM: 未检测到独立显卡\n")
		}
	}

	// 测试GPU信息获取
	fmt.Println("\n🎮 获取GPU信息...")
	gpuInfo, err := hardware.GetGPUInfo()
	if err != nil {
		log.Printf("❌ 获取GPU信息失败: %v", err)
	} else {
		if len(gpuInfo) == 0 {
			fmt.Printf("⚠️  未检测到GPU设备\n")
		} else {
			fmt.Printf("✅ 检测到 %d 个GPU设备:\n", len(gpuInfo))
			for i, gpu := range gpuInfo {
				fmt.Printf("   GPU %d: %s\n", i+1, gpu.Name)
				if gpu.MemoryTotal > 0 {
					fmt.Printf("         显存: %s / %s\n",
						hardware.FormatBytes(gpu.MemoryUsed),
						hardware.FormatBytes(gpu.MemoryTotal))
				}
				if gpu.Utilization >= 0 {
					fmt.Printf("         使用率: %.1f%%\n", gpu.Utilization)
				}
				if gpu.Temperature > 0 {
					fmt.Printf("         温度: %.1f°C\n", gpu.Temperature)
				}
			}
		}
	}

	// 测试完整硬件信息获取
	fmt.Println("\n🖥️  获取完整硬件信息...")
	fullInfo, err := hardware.GetSystemHardwareInfo()
	if err != nil {
		log.Printf("❌ 获取完整硬件信息失败: %v", err)
	} else {
		fmt.Printf("✅ 完整硬件信息获取成功\n")
		fmt.Printf("   CPU: %s (%s)\n", fullInfo.CPU.ModelName, fullInfo.CPU.Architecture)
		fmt.Printf("   内存: %s 总容量\n", hardware.FormatBytes(fullInfo.Memory.RAMTotal))
		fmt.Printf("   GPU: %d 个设备\n", len(fullInfo.GPUs))
	}

	// 打印详细的硬件信息
	fmt.Println("\n📋 详细硬件信息:")
	fmt.Println("================================")
	if err := hardware.PrintSystemHardwareInfo(); err != nil {
		log.Printf("❌ 打印详细信息失败: %v", err)
	}
}
