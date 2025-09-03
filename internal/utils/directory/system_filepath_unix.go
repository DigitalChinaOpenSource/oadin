//go:build !windows
// +build !windows

package directory

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// UnixPaths 存储 Unix/Linux/macOS 系统各种路径信息
type UnixPaths struct {
	// 系统目录
	HomeDir     string // /home/{username} 或 /Users/{username}
	TempDir     string // /tmp
	ConfigDir   string // ~/.config
	CacheDir    string // ~/.cache
	DataDir     string // ~/.local/share
	DocumentsDir string // ~/Documents
	DesktopDir   string // ~/Desktop
	DownloadsDir string // ~/Downloads
	
	// 系统级目录
	UsrLocalDir  string // /usr/local
	UsrBinDir    string // /usr/bin
	OptDir       string // /opt
	EtcDir       string // /etc
	VarDir       string // /var
	
	// 应用数据目录
	ApplicationDirs []string // 应用程序可能的安装位置
}

// GetUnixPaths 获取 Unix/Linux/macOS 系统所有重要路径
func GetUnixPaths() (*UnixPaths, error) {
	paths := &UnixPaths{}
	
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %v", err)
	}
	paths.HomeDir = homeDir
	
	// 系统临时目录
	paths.TempDir = os.TempDir()
	
	// XDG 基本目录规范的目录
	paths.ConfigDir = filepath.Join(homeDir, ".config")
	paths.CacheDir = filepath.Join(homeDir, ".cache")
	paths.DataDir = filepath.Join(homeDir, ".local", "share")
	
	// 常见用户目录
	paths.DocumentsDir = filepath.Join(homeDir, "Documents")
	paths.DesktopDir = filepath.Join(homeDir, "Desktop")
	paths.DownloadsDir = filepath.Join(homeDir, "Downloads")
	
	// 系统级目录
	paths.UsrLocalDir = "/usr/local"
	paths.UsrBinDir = "/usr/bin"
	paths.OptDir = "/opt"
	paths.EtcDir = "/etc"
	paths.VarDir = "/var"
	
	// 应用程序可能的安装位置
	paths.ApplicationDirs = []string{
		"/usr/local/bin",
		"/usr/bin",
		"/opt",
		"/Applications",              // macOS 专用
		filepath.Join(homeDir, "Applications"), // macOS 用户级应用
		"/usr/local/opt",            // Homebrew 安装位置 (macOS)
		"/snap/bin",                 // Snap 包 (Linux)
		"/usr/share/applications",   // Linux 应用程序
	}
	
	return paths, nil
}

// 以下是与 Windows 版本保持兼容的函数

// GetWindowsPaths 在非 Windows 系统上返回错误
func GetWindowsPaths() (interface{}, error) {
	return nil, fmt.Errorf("不支持在 %s 平台上获取 Windows 路径", runtime.GOOS)
}

// GetPreinstalledSoftwarePaths 获取系统预装软件可能的安装路径
func GetPreinstalledSoftwarePaths() ([]string, error) {
	paths, err := GetUnixPaths()
	if err != nil {
		return nil, err
	}
	
	// 过滤存在的目录
	var existingPaths []string
	for _, path := range paths.ApplicationDirs {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}
	
	return existingPaths, nil
}

// GetDataDirectories 获取应用程序数据目录
func GetDataDirectories() (map[string]string, error) {
	paths, err := GetUnixPaths()
	if err != nil {
		return nil, err
	}
	
	dataDirectories := map[string]string{
		"Home":      paths.HomeDir,
		"Config":    paths.ConfigDir,
		"Cache":     paths.CacheDir,
		"Data":      paths.DataDir,
		"Documents": paths.DocumentsDir,
		"Desktop":   paths.DesktopDir,
		"Downloads": paths.DownloadsDir,
		"Temp":      paths.TempDir,
	}
	
	return dataDirectories, nil
}

// getProgramPath 获取程序安装路径
func getProgramPath() (string, error) {
	// 返回当前执行文件的目录
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(executable), nil
}

// PrintWindowsPaths 打印系统路径信息（用于调试）
func PrintWindowsPaths() error {
	paths, err := GetUnixPaths()
	if err != nil {
		return err
	}
	
	fmt.Println("=== Unix/Linux/macOS 系统路径信息 ===")
	fmt.Printf("用户主目录: %s\n", paths.HomeDir)
	fmt.Printf("临时目录: %s\n", paths.TempDir)
	fmt.Printf("配置目录: %s\n", paths.ConfigDir)
	fmt.Printf("缓存目录: %s\n", paths.CacheDir)
	fmt.Printf("数据目录: %s\n", paths.DataDir)
	
	fmt.Println("\n=== 用户目录 ===")
	fmt.Printf("文档目录: %s\n", paths.DocumentsDir)
	fmt.Printf("桌面目录: %s\n", paths.DesktopDir)
	fmt.Printf("下载目录: %s\n", paths.DownloadsDir)
	
	fmt.Println("\n=== 系统目录 ===")
	fmt.Printf("/usr/local: %s\n", paths.UsrLocalDir)
	fmt.Printf("/usr/bin: %s\n", paths.UsrBinDir)
	fmt.Printf("/opt: %s\n", paths.OptDir)
	fmt.Printf("/etc: %s\n", paths.EtcDir)
	fmt.Printf("/var: %s\n", paths.VarDir)
	
	fmt.Println("\n=== 应用程序目录 ===")
	for i, path := range paths.ApplicationDirs {
		fmt.Printf("%d. %s\n", i+1, path)
	}
	
	return nil
}
