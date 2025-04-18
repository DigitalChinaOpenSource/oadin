//go:build windows

package utils

import (
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/jaypipes/ghw"
	"golang.org/x/sys/windows"
	"strconv"
)

type Win32_PhysicalMemory struct {
	SMBIOSMemoryType int
}

func GetMemoryInfo() (*MemoryInfo, error) {
	var win32Memories []Win32_PhysicalMemory
	q := wmi.CreateQuery(&win32Memories, "")
	err := wmi.Query(q, &win32Memories)
	if err != nil {
		fmt.Println(err)
	}
	memory, err := ghw.Memory()
	if err != nil {
		return nil, err
	}
	memoryType := strconv.Itoa(win32Memories[0].SMBIOSMemoryType)
	finalMemoryType := memoryTypeFromCode(memoryType)
	memoryInfo := MemoryInfo{
		MemoryType: finalMemoryType,
		Size:       int(memory.TotalPhysicalBytes / 1024 / 1024 / 1024),
	}
	return &memoryInfo, nil
}

// Convert Windows memory type codes to DDR types.
func memoryTypeFromCode(code string) string {
	switch code {
	case "20":
		return "DDR"
	case "21":
		return "DDR2"
	case "22":
		return "DDR2 FB-DIMM"
	case "24":
		return "DDR3"
	case "26":
		return "DDR4"
	case "34":
		return "DDR5"
	case "35":
		return "DDR5"
	default:
		return "Unknown (" + code + ")"
	}
}

func GetSystemVersion() int {
	systemVersion := 0
	info := windows.RtlGetVersion()
	if info.MajorVersion == 10 {
		if info.BuildNumber >= 22000 {
			systemVersion = 11
		} else if info.BuildNumber >= 10240 && info.BuildNumber <= 19045 {
			systemVersion = 10
		}
	}
	return systemVersion
}
