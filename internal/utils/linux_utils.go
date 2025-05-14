//go:build linux

package utils

import "github.com/shirou/gopsutil/mem"

func GetMemoryInfo() (*MemoryInfo, error) {
	v, _ := mem.VirtualMemory()
	memorySize := int(v.Total / 1024 / 1024 / 1024)
	memoryInfo := &MemoryInfo{
		Size: memorySize,
	}
	return memoryInfo, nil
}

func GetSystemVersion() int {
	systemVersion := 0
	return systemVersion
}
