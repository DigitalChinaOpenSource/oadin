package main

import (
	"fmt"
	"log"

	"oadin/extension/utils/hardware"
)

func main() {
	fmt.Println("🚀 OADIN 系统硬件监控测试")
	fmt.Println("================================")

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

	// 打印详细的内存信息
	fmt.Println("\n📋 详细内存信息:")
	fmt.Println("================================")
	if err := hardware.PrintMemoryInfo(); err != nil {
		log.Printf("❌ 打印详细信息失败: %v", err)
	}
}
