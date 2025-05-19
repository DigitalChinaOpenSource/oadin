//go:build linux || darwin || freebsd || openbsd || netbsd || solaris
// +build linux darwin freebsd openbsd netbsd solaris

package hardware

import (
	"golang.org/x/sys/unix"
)

func GetDiskUsage(path string) (*DiskUsage, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return nil, err
	}
	// 每块的大小（字节）
	blockSize := uint64(stat.Bsize)
	// 总块数
	totalBlocks := uint64(stat.Blocks)
	// 可用块数（对普通用户）
	freeBlocks := uint64(stat.Bavail)
	// 已使用块数
	usedBlocks := totalBlocks - freeBlocks
	total := totalBlocks * blockSize
	free := freeBlocks * blockSize
	used := usedBlocks * blockSize
	return &DiskUsage{
		Total:     total,
		Free:      free,
		Used:      used,
		UsedRatio: float64(used) / float64(total),
	}, nil
}
