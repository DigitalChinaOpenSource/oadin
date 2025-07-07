//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

//go:build windows

package utils

import (
	"fmt"
	"strconv"

	"github.com/StackExchange/wmi"
	"github.com/jaypipes/ghw"
	"golang.org/x/sys/windows"
)

const (
	// Windows memory type codes
	MemoryTypeDDR        = "20"
	MemoryTypeDDR2       = "21"
	MemoryTypeDDR2FBDIMM = "22"
	MemoryTypeDDR3       = "24"
	MemoryTypeDDR4       = "26"
	MemoryTypeDDR5_1     = "34"
	MemoryTypeDDR5_2     = "35"

	// Memory type names
	MemoryNameDDR        = "DDR"
	MemoryNameDDR2       = "DDR2"
	MemoryNameDDR2FBDIMM = "DDR2 FB-DIMM"
	MemoryNameDDR3       = "DDR3"
	MemoryNameDDR4       = "DDR4"
	MemoryNameDDR5       = "DDR5"
	MemoryNameUnknown    = "Unknown"

	// Memory unit conversions
	BytesToGB = 1024 * 1024 * 1024

	// Windows version identification
	WindowsBuildWindows11    = 22000
	WindowsBuildWindows10Min = 10240
	WindowsBuildWindows10Max = 19045
	WindowsMajorVersion10    = 10
	WindowsVersion10         = 10
	WindowsVersion11         = 11
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
		Size:       int(memory.TotalPhysicalBytes / BytesToGB),
	}
	return &memoryInfo, nil
}

// Convert Windows memory type codes to DDR types.
func memoryTypeFromCode(code string) string {
	switch code {
	case MemoryTypeDDR:
		return MemoryNameDDR
	case MemoryTypeDDR2:
		return MemoryNameDDR2
	case MemoryTypeDDR2FBDIMM:
		return MemoryNameDDR2FBDIMM
	case MemoryTypeDDR3:
		return MemoryNameDDR3
	case MemoryTypeDDR4:
		return MemoryNameDDR4
	case MemoryTypeDDR5_1:
		return MemoryNameDDR5
	case MemoryTypeDDR5_2:
		return MemoryNameDDR5
	default:
		return MemoryNameUnknown + " (" + code + ")"
	}
}

func GetSystemVersion() int {
	systemVersion := 0
	info := windows.RtlGetVersion()
	if info.MajorVersion == WindowsMajorVersion10 {
		if info.BuildNumber >= WindowsBuildWindows11 {
			systemVersion = WindowsVersion11
		} else if info.BuildNumber >= WindowsBuildWindows10Min && info.BuildNumber <= WindowsBuildWindows10Max {
			systemVersion = WindowsVersion10
		}
	}
	return systemVersion
}
