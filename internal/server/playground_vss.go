package server

import (
	"context"
	"log/slog"
	"os"
	"strings"
)


func InitPlaygroundVSS(ctx context.Context, dbPath string) error {
	slog.Info("开始初始化VSS数据库", "db_path", dbPath)

	if !UseVSSForPlayground() {
		slog.Info("VSS功能已通过环境变量禁用")
		return nil
	}

	if err := initVSSDB(dbPath); err != nil {
		slog.Warn("初始化VSS数据库失败，将使用标准搜索", "error", err)
		return nil 
	}

	return nil
}


func UseVSSForPlayground() bool {
	// 优先从环境变量中检查
	useVSSEnv := os.Getenv("BYZE_USE_VSS")
	if strings.ToLower(useVSSEnv) == "false" {
		return false
	}

	// 默认启用VSS
	return true
}
