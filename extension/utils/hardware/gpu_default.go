//go:build !windows && !linux && !darwin

package hardware

import "fmt"

// getVRAMInfo 默认实现，返回0值
func getVRAMInfo() (total, used, free uint64) {
	return 0, 0, 0
}

// getWindowsGPUInfo 默认实现（在非Windows平台）
func getWindowsGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Windows GPU info not available on this platform")
}

// getLinuxGPUInfo 默认实现（在非Linux平台）
func getLinuxGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("Linux GPU info not available on this platform")
}

// getMacOSGPUInfo 默认实现（在非macOS平台）
func getMacOSGPUInfo() ([]GPUInfo, error) {
	return nil, fmt.Errorf("macOS GPU info not available on this platform")
}
