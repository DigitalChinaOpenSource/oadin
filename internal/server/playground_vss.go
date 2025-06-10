package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)


func InitPlaygroundVSS(ctx context.Context, dbPath string) error {
	slog.Info("开始初始化VSS数据库", "db_path", dbPath)


	if err := initVSSDB(dbPath); err != nil {
		return fmt.Errorf("初始化VSS数据库失败: %w", err)
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
