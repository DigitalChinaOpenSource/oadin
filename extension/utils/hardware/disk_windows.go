//go:build windows

package hardware

import (
	"syscall"
	"unsafe"
)

func GetDiskUsage(path string) (*DiskUsage, error) {
	// 定义Windows API使用的结构
	var freeBytes, totalBytes, totalFreeBytes uint64
	// 加载kernel32.dll中的函数
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx := kernel32.NewProc("GetDiskFreeSpaceExW")
	// 转换路径为UTF16指针
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	// 调用Windows API
	ret, _, e1 := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return nil, e1
	}
	used := totalBytes - freeBytes
	return &DiskUsage{
		Total:     totalBytes,
		Free:      freeBytes,
		Used:      used,
		UsedRatio: float64(used) / float64(totalBytes),
	}, nil
}
