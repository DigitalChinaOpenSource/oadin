//go:build linux

package utils

import "github.com/jaypipes/ghw"

func GetMemoryInfo() (*MemoryInfo, error) {
	memoryObj, err := ghw.Memory()
	if err != nil {
		return &MemoryInfo{}, err
	}
	memorySize := int(memoryObj.TotalUsableBytes / 1024 / 1024 / 1024)
	memoryInfo = &MemoryInfo{
		Size: memorySize,
	}
	return memoryInfo, nil
}

func GetSystemVersion() int {
	systemVersion := 0
	return systemVersion
}
