package hardware

type DiskUsage struct {
	Total     uint64  `json:"total"`      // 磁盘总空间（字节）
	Free      uint64  `json:"free"`       // 磁盘剩余空间（字节）
	Used      uint64  `json:"used"`       // 磁盘已使用空间（字节）
	UsedRatio float64 `json:"used_ratio"` // 磁盘使用率 (0-1)
}

// GetDiskUsage 获取指定路径所在磁盘的使用情况
func GetDiskUsageInformation(path string) (*DiskUsage, error) {
	return GetDiskUsage(path)
}
