//go:build windows
// +build windows

package directory

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

// WindowsPaths 存储Windows系统各种路径信息
type WindowsPaths struct {
	// 程序目录
	ProgramFiles    string // C:\Program Files
	ProgramFilesX86 string // C:\Program Files (x86)
	ProgramData     string // C:\ProgramData
	WindowsDir      string // C:\Windows
	SystemDir       string // C:\Windows\System32
	System32Dir     string // C:\Windows\System32
	SysWOW64Dir     string // C:\Windows\SysWOW64

	// 用户数据目录
	UserProfile    string // C:\Users\{username}
	AppData        string // C:\Users\{username}\AppData
	AppDataLocal   string // C:\Users\{username}\AppData\Local
	AppDataRoaming string // C:\Users\{username}\AppData\Roaming
	LocalAppData   string // C:\Users\{username}\AppData\Local
	Documents      string // C:\Users\{username}\Documents
	Desktop        string // C:\Users\{username}\Desktop
	Downloads      string // C:\Users\{username}\Downloads

	// 公共目录
	PublicDesktop   string // C:\Users\Public\Desktop
	PublicDocuments string // C:\Users\Public\Documents
	PublicDownloads string // C:\Users\{username}\Downloads

	// 临时目录
	TempDir     string // C:\Users\{username}\AppData\Local\Temp
	WindowsTemp string // C:\Windows\Temp

	// 预装软件常见目录
	PreinstalledApps []string // 预装软件可能的安装位置
}

// Windows API 常量
const (
	CSIDL_PROGRAM_FILES           = 0x0026 // C:\Program Files
	CSIDL_PROGRAM_FILESX86        = 0x002a // C:\Program Files (x86)
	CSIDL_PROGRAM_FILES_COMMON    = 0x002b // C:\Program Files\Common Files
	CSIDL_WINDOWS                 = 0x0024 // C:\Windows
	CSIDL_SYSTEM                  = 0x0025 // C:\Windows\System32
	CSIDL_APPDATA                 = 0x001a // C:\Users\{username}\AppData\Roaming
	CSIDL_LOCAL_APPDATA           = 0x001c // C:\Users\{username}\AppData\Local
	CSIDL_COMMON_APPDATA          = 0x0023 // C:\ProgramData
	CSIDL_PROFILE                 = 0x0028 // C:\Users\{username}
	CSIDL_MYDOCUMENTS             = 0x0005 // C:\Users\{username}\Documents
	CSIDL_DESKTOP                 = 0x0000 // C:\Users\{username}\Desktop
	CSIDL_COMMON_DESKTOPDIRECTORY = 0x0019 // C:\Users\Public\Desktop
	CSIDL_COMMON_DOCUMENTS        = 0x002e // C:\Users\Public\Documents
)

var (
	shell32         = syscall.NewLazyDLL("shell32.dll")
	shGetFolderPath = shell32.NewProc("SHGetFolderPathW")
)

// getSpecialFolder 获取Windows特殊文件夹路径
func getSpecialFolder(csidl int) (string, error) {
	if runtime.GOOS != "windows" {
		return "", fmt.Errorf("this function only works on Windows")
	}

	var path [260]uint16
	r1, _, err := shGetFolderPath.Call(
		0,
		uintptr(csidl),
		0,
		0,
		uintptr(unsafe.Pointer(&path[0])),
	)

	if r1 != 0 {
		return "", fmt.Errorf("SHGetFolderPath failed: %v", err)
	}

	return syscall.UTF16ToString(path[:]), nil
}

// GetWindowsPaths 获取Windows系统所有重要路径
func GetWindowsPaths() (*WindowsPaths, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("this function only works on Windows")
	}

	paths := &WindowsPaths{}
	var err error

	// 获取程序目录
	if paths.ProgramFiles, err = getSpecialFolder(CSIDL_PROGRAM_FILES); err != nil {
		paths.ProgramFiles = os.Getenv("PROGRAMFILES")
	}

	if paths.ProgramFilesX86, err = getSpecialFolder(CSIDL_PROGRAM_FILESX86); err != nil {
		paths.ProgramFilesX86 = os.Getenv("PROGRAMFILES(X86)")
	}

	if paths.ProgramData, err = getSpecialFolder(CSIDL_COMMON_APPDATA); err != nil {
		paths.ProgramData = os.Getenv("PROGRAMDATA")
	}

	if paths.WindowsDir, err = getSpecialFolder(CSIDL_WINDOWS); err != nil {
		paths.WindowsDir = os.Getenv("WINDIR")
	}

	if paths.SystemDir, err = getSpecialFolder(CSIDL_SYSTEM); err != nil {
		paths.SystemDir = filepath.Join(paths.WindowsDir, "System32")
	}
	paths.System32Dir = paths.SystemDir

	// SysWOW64 目录 (64位系统上的32位程序目录)
	paths.SysWOW64Dir = filepath.Join(paths.WindowsDir, "SysWOW64")

	// 获取用户目录
	if paths.UserProfile, err = getSpecialFolder(CSIDL_PROFILE); err != nil {
		paths.UserProfile = os.Getenv("USERPROFILE")
	}

	if paths.AppDataRoaming, err = getSpecialFolder(CSIDL_APPDATA); err != nil {
		paths.AppDataRoaming = filepath.Join(paths.UserProfile, "AppData", "Roaming")
	}

	if paths.AppDataLocal, err = getSpecialFolder(CSIDL_LOCAL_APPDATA); err != nil {
		paths.AppDataLocal = filepath.Join(paths.UserProfile, "AppData", "Local")
	}

	paths.AppData = filepath.Join(paths.UserProfile, "AppData")
	paths.LocalAppData = paths.AppDataLocal

	if paths.Documents, err = getSpecialFolder(CSIDL_MYDOCUMENTS); err != nil {
		paths.Documents = filepath.Join(paths.UserProfile, "Documents")
	}

	if paths.Desktop, err = getSpecialFolder(CSIDL_DESKTOP); err != nil {
		paths.Desktop = filepath.Join(paths.UserProfile, "Desktop")
	}

	paths.Downloads = filepath.Join(paths.UserProfile, "Downloads")

	// 公共目录
	if paths.PublicDesktop, err = getSpecialFolder(CSIDL_COMMON_DESKTOPDIRECTORY); err != nil {
		paths.PublicDesktop = "C:\\Users\\Public\\Desktop"
	}

	if paths.PublicDocuments, err = getSpecialFolder(CSIDL_COMMON_DOCUMENTS); err != nil {
		paths.PublicDocuments = "C:\\Users\\Public\\Documents"
	}

	paths.PublicDownloads = "C:\\Users\\Public\\Downloads"

	// 临时目录
	paths.TempDir = os.Getenv("TEMP")
	if paths.TempDir == "" {
		paths.TempDir = filepath.Join(paths.AppDataLocal, "Temp")
	}

	paths.WindowsTemp = filepath.Join(paths.WindowsDir, "Temp")

	// 预装软件常见目录
	paths.PreinstalledApps = []string{
		paths.ProgramFiles,
		paths.ProgramFilesX86,
		filepath.Join(paths.ProgramData, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(paths.AppDataRoaming, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(paths.UserProfile, "AppData", "Local", "Programs"),
		filepath.Join(paths.WindowsDir, "System32"),
		filepath.Join(paths.WindowsDir, "SysWOW64"),
		"C:\\Windows\\Microsoft.NET\\Framework",
		"C:\\Windows\\Microsoft.NET\\Framework64",
		"C:\\Windows\\WinSxS",
	}

	return paths, nil
}

// GetPreinstalledSoftwarePaths 获取Windows预装软件可能的安装路径
func GetPreinstalledSoftwarePaths() ([]string, error) {
	paths, err := GetWindowsPaths()
	if err != nil {
		return nil, err
	}

	preinstalledPaths := []string{
		// 主要程序目录
		paths.ProgramFiles,
		paths.ProgramFilesX86,

		// Windows内置应用
		filepath.Join(paths.WindowsDir, "System32"),
		filepath.Join(paths.WindowsDir, "SysWOW64"),

		// Microsoft Store 应用
		filepath.Join(paths.ProgramFiles, "WindowsApps"),

		// .NET Framework
		filepath.Join(paths.WindowsDir, "Microsoft.NET", "Framework"),
		filepath.Join(paths.WindowsDir, "Microsoft.NET", "Framework64"),

		// Windows 功能和组件
		filepath.Join(paths.WindowsDir, "WinSxS"),

		// 用户级预装应用
		filepath.Join(paths.AppDataLocal, "Programs"),
		filepath.Join(paths.AppDataLocal, "Microsoft", "WindowsApps"),

		// 公共程序数据
		paths.ProgramData,
		filepath.Join(paths.ProgramData, "Microsoft"),

		// 开始菜单程序
		filepath.Join(paths.ProgramData, "Microsoft", "Windows", "Start Menu", "Programs"),
		filepath.Join(paths.AppDataRoaming, "Microsoft", "Windows", "Start Menu", "Programs"),
	}

	// 过滤存在的目录
	var existingPaths []string
	for _, path := range preinstalledPaths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	return existingPaths, nil
}

// GetDataDirectories 获取应用程序数据目录
func GetDataDirectories() (map[string]string, error) {
	paths, err := GetWindowsPaths()
	if err != nil {
		return nil, err
	}

	dataDirectories := map[string]string{
		"AppData":        paths.AppData,
		"LocalAppData":   paths.AppDataLocal,
		"RoamingAppData": paths.AppDataRoaming,
		"ProgramData":    paths.ProgramData,
		"Documents":      paths.Documents,
		"Desktop":        paths.Desktop,
		"Downloads":      paths.Downloads,
		"Temp":           paths.TempDir,
		"WindowsTemp":    paths.WindowsTemp,
		"UserProfile":    paths.UserProfile,
	}

	return dataDirectories, nil
}

// 更新原有函数
func getProgramPath() (string, error) {
	if runtime.GOOS != "windows" {
		// 对于非Windows系统，返回当前执行文件的目录
		executable, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.Dir(executable), nil
	}

	// Windows系统获取程序安装目录
	paths, err := GetWindowsPaths()
	if err != nil {
		return "", err
	}

	// 返回当前程序可能的安装位置
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	execDir := filepath.Dir(executable)

	// 检查是否在标准程序目录中
	if strings.HasPrefix(execDir, paths.ProgramFiles) ||
		strings.HasPrefix(execDir, paths.ProgramFilesX86) {
		return execDir, nil
	}

	// 如果不在标准目录，返回可执行文件所在目录
	return execDir, nil
}

// PrintWindowsPaths 打印所有Windows路径信息（用于调试）
func PrintWindowsPaths() error {
	paths, err := GetWindowsPaths()
	if err != nil {
		return err
	}

	fmt.Println("=== Windows 系统路径信息 ===")
	fmt.Printf("程序文件目录: %s\n", paths.ProgramFiles)
	fmt.Printf("程序文件目录(x86): %s\n", paths.ProgramFilesX86)
	fmt.Printf("程序数据目录: %s\n", paths.ProgramData)
	fmt.Printf("Windows目录: %s\n", paths.WindowsDir)
	fmt.Printf("System32目录: %s\n", paths.System32Dir)
	fmt.Printf("SysWOW64目录: %s\n", paths.SysWOW64Dir)

	fmt.Println("\n=== 用户目录 ===")
	fmt.Printf("用户配置文件: %s\n", paths.UserProfile)
	fmt.Printf("应用数据目录: %s\n", paths.AppData)
	fmt.Printf("本地应用数据: %s\n", paths.AppDataLocal)
	fmt.Printf("漫游应用数据: %s\n", paths.AppDataRoaming)
	fmt.Printf("文档目录: %s\n", paths.Documents)
	fmt.Printf("桌面目录: %s\n", paths.Desktop)
	fmt.Printf("下载目录: %s\n", paths.Downloads)

	fmt.Println("\n=== 预装软件目录 ===")
	for i, path := range paths.PreinstalledApps {
		fmt.Printf("%d. %s\n", i+1, path)
	}

	return nil
}
