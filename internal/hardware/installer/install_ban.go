package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	bunReleaseBaseURL = "https://gitcode.com/CherryHQ/bun/releases/download"
	defaultBunVersion = "1.2.9"
)

var bunPackages = map[string]string{
	"darwin-arm64": "bun-darwin-aarch64.zip",
	"darwin-x64":   "bun-darwin-x64.zip",
	"win32-x64":    "bun-windows-x64.zip",
	// 增加新支持
	"windows-amd64":          "bun-windows-x64.zip",
	"windows-amd64-baseline": "bun-windows-x64.zip",
	"win32-x64-baseline":     "bun-windows-x64-baseline.zip",
	"win32-arm64":            "bun-windows-x64.zip",
	"win32-arm64-baseline":   "bun-windows-x64-baseline.zip",
	"linux-x64":              "bun-linux-x64.zip",
	"linux-x64-baseline":     "bun-linux-x64-baseline.zip",
	"linux-arm64":            "bun-linux-aarch64.zip",
	// MUSL variants
	"linux-musl-x64":          "bun-linux-x64-musl.zip",
	"linux-musl-x64-baseline": "bun-linux-x64-musl-baseline.zip",
	"linux-musl-arm64":        "bun-linux-aarch64-musl.zip",
}

func downloadBunBinary(platform, arch, version string, isMusl, isBaseline bool) error {
	// 确定平台和包名
	var platformKey string
	if isMusl {
		platformKey = fmt.Sprintf("%s-musl-%s", platform, arch)
		if isBaseline {
			platformKey += "-baseline"
		}
	} else {
		platformKey = fmt.Sprintf("%s-%s", platform, arch)
		if isBaseline {
			platformKey += "-baseline"
		}
	}
	packageName, ok := bunPackages[platformKey]
	if !ok {
		return fmt.Errorf("no binary available for %s", platformKey)
	}
	// 创建输出目录
	binDir := RuntimePath()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// 下载URL
	downloadURL := fmt.Sprintf("%s/bun-v%s/%s", bunReleaseBaseURL, version, packageName)
	tempDir := os.TempDir()
	tempFilename := filepath.Join(tempDir, packageName)
	// 下载文件
	if err := DownloadWithRedirects(downloadURL, tempFilename, nil); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	fmt.Printf("Downloaded %s to %s\n", packageName, tempFilename)
	// 打开ZIP文件
	r, err := zip.OpenReader(tempFilename)
	if err != nil {
		// 删除临时文件
		if errRemove := os.Remove(tempFilename); errRemove != nil {
			return fmt.Errorf("failed to remove temporary file: %v", errRemove)
		}
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()
	fmt.Printf("Extracting %s to %s...\n", packageName, binDir)
	// 解压缩文件
	for _, file := range r.File {
		// 跳过目录
		if file.FileInfo().IsDir() {
			continue
		}
		// 创建文件路径
		// ps:这里制定固定的名字,防止解压到子级文件夹下面
		destPath := filepath.Join(binDir, filepath.Base(file.Name))
		destDir := filepath.Dir(destPath)
		// 创建目标目录
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		// 打开ZIP文件流
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file from zip: %v", err)
		}
		defer src.Close()
		// 创建目标文件
		dest, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dest.Close()
		// 拷贝文件内容
		if _, err := io.Copy(dest, src); err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}
		// 确保数据同步到磁盘
		if err := dest.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %v", err)
		}
		fmt.Printf("Extracted %s\n", destPath)
	}
	// 删除临时文件
	if err := os.Remove(tempFilename); err != nil {
		return fmt.Errorf("failed to remove temporary file: %v", err)
	}
	fmt.Printf("Successfully installed bun %s for %s\n", version, platformKey)
	return nil
}
func detectPlatformAndArch() (string, string, bool, bool) {
	platform := runtime.GOOS
	arch := runtime.GOARCH
	var isMusl bool
	if platform == "linux" {
		// Simple check for Alpine Linux which uses MUSL
		output, err := exec.Command("cat", "/etc/os-release").Output()
		if err == nil {
			isMusl = strings.Contains(strings.ToLower(string(output)), "alpine")
		}
	}
	isBaseline := platform == "windows"
	return platform, arch, isMusl, isBaseline
}
func InstallBun() error {
	version := defaultBunVersion
	fmt.Printf("Using bun version: %s\n", version)
	platform, arch, isMusl, isBaseline := detectPlatformAndArch()
	fmt.Printf("Installing bun %s for %s-%s%s%s...\n",
		version,
		platform,
		arch,
		func() string {
			if isMusl {
				return " (MUSL)"
			}
			return ""
		}(),
		func() string {
			if isBaseline {
				return " (baseline)"
			}
			return ""
		}())
	return downloadBunBinary(platform, arch, version, isMusl, isBaseline)
}
