package main

import (
	"context"
	"fmt"
	"oadin/internal/utils"
	"oadin/tray"
	"os"
	"path/filepath"
	"sync"

	"oadin/config"
	"oadin/internal/logger"
)

var (
	mutex    sync.Mutex
	lockFile *os.File
)

// 检查是否已有实例在运行
func checkSingleInstance() bool {
	lockPath := filepath.Join(os.TempDir(), "oadin-tray.lock")

	var err error
	lockFile, err = os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		// 文件已存在，说明可能有实例在运行
		fmt.Printf("Another instance may be running: %v\n", err)
		return false
	}

	// 写入进程ID
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	return true
}

// 清理锁文件
func cleanup() {
	if lockFile != nil {
		lockFile.Close()
		os.Remove(lockFile.Name())
	}
}

func main() {
	fmt.Println("Starting Oadin Tray Application...")

	// // 检查单实例
	// if !checkSingleInstance() {
	// 	fmt.Println("Another instance is already running. Exiting...")
	// 	return
	// }

	// // 确保程序退出时清理
	// defer cleanup()

	// 获取用户目录
	rootDir, err := utils.GetOADINDataDir()
	if err != nil {
		fmt.Printf("Failed to get oadin home directory: %v\n", err)
		return
	}

	logDirPath := filepath.Join(rootDir, "logs")
	logFilePath := filepath.Join(logDirPath, "oadin.log")
	pidPath := rootDir

	// 确保目录存在
	os.MkdirAll(logDirPath, 0755)
	os.MkdirAll(pidPath, 0755)
	os.Create(logFilePath)

	// 创建托盘管理器
	trayManager := tray.NewManager(true, logFilePath, pidPath)
	ctx := context.Background()

	config.GlobalEnvironment = config.NewOADINEnvironment()
	logger.InitLogger(logger.LogConfig{LogLevel: config.GlobalEnvironment.LogLevel, LogPath: config.GlobalEnvironment.LogDir})
	// check updates
	go tray.StartCheckUpdate(ctx, trayManager)
	// 启动托盘
	trayManager.Start()
}
