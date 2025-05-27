package utils

type MemoryInfo struct {
	Size       int
	MemoryType string
}

type EnvVariables struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PathDiskSizeInfo struct {
	FreeSize  int `json:"free_size"`
	TotalSize int `json:"total_size"`
	UsageSize int `json:"usage_size"`
}
