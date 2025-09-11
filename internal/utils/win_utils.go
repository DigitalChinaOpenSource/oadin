//go:build windows

package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/jaypipes/ghw"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
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

func SamePartitionStatus(srcPath, targetPath string) (bool, error) {
	abs1, err := filepath.Abs(srcPath)
	if err != nil {
		return false, err
	}

	abs2, err := filepath.Abs(targetPath)
	if err != nil {
		return false, err
	}

	drive1 := strings.ToUpper(filepath.VolumeName(abs1))
	drive2 := strings.ToUpper(filepath.VolumeName(abs2))

	return drive1 == drive2, nil
}

func ModifySystemUserVariables(envInfo *EnvVariables) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		panic(err)
	}
	defer key.Close()

	// set environment variables
	err = key.SetStringValue(envInfo.Name, envInfo.Value)
	if err != nil {
		return err
	}

	// Notify the system that environment variables have changed
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	hwndBroadcast := uintptr(0xffff)
	wmSettingChange := uintptr(0x001A)
	smtoAbortIfHung := uintptr(0x0002)

	r1, _, _ := sendMessageTimeout.Call(
		hwndBroadcast,
		wmSettingChange,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
		smtoAbortIfHung,
		5000,
		0,
	)
	if r1 == 0 {
		return errors.New("Failed to notify the system of environment variable changes. Please restart or log out for the changes to take effect.")
	}
	return nil
}

func SetCmdSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CERT_TRUST_HAS_NOT_SUPPORTED_CRITICAL_EXT | syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

func CheckDllExists(dllName string) bool {
	winDir := os.Getenv("WINDIR")

	if runtime.GOARCH == "amd64" {
		path := filepath.Join(winDir, "System32", dllName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	} else {
		path := filepath.Join(winDir, "SysWOW64", dllName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
		path = filepath.Join(winDir, "System32", dllName)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}
