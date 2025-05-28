package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	uvReleaseBaseURL = "https://gitcode.com/CherryHQ/uv/releases/download"
	defaultUVVersion = "0.6.14"
)

var uvPackages = map[string]string{
	"darwin-arm64": "uv-aarch64-apple-darwin.tar.gz",
	"darwin-x64":   "uv-x86_64-apple-darwin.tar.gz",
	"win32-arm64":  "uv-aarch64-pc-windows-msvc.zip",
	"win32-ia32":   "uv-i686-pc-windows-msvc.zip",
	"win32-x64":    "uv-x86_64-pc-windows-msvc.zip",
	// 增加新支持
	"windows-amd64": "uv-x86_64-pc-windows-msvc.zip",

	"linux-arm64":   "uv-aarch64-unknown-linux-gnu.tar.gz",
	"linux-ia32":    "uv-i686-unknown-linux-gnu.tar.gz",
	"linux-ppc64":   "uv-powerpc64-unknown-linux-gnu.tar.gz",
	"linux-ppc64le": "uv-powerpc64le-unknown-linux-gnu.tar.gz",
	"linux-s390x":   "uv-s390x-unknown-linux-gnu.tar.gz",
	"linux-x64":     "uv-x86_64-unknown-linux-gnu.tar.gz",
	"linux-armv7l":  "uv-armv7-unknown-linux-gnueabihf.tar.gz",
	// MUSL variants
	"linux-musl-arm64":  "uv-aarch64-unknown-linux-musl.tar.gz",
	"linux-musl-ia32":   "uv-i686-unknown-linux-musl.tar.gz",
	"linux-musl-x64":    "uv-x86_64-unknown-linux-musl.tar.gz",
	"linux-musl-armv6l": "uv-arm-unknown-linux-musleabihf.tar.gz",
	"linux-musl-armv7l": "uv-armv7-unknown-linux-musleabihf.tar.gz",
}

func downloadWithRedirects(url string, destinationPath string) error {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status code: %d", resp.StatusCode)
	}
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	file, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	return nil
}
func downloadUvBinary(platform, arch, version string, isMusl bool) error {
	var platformKey string
	if isMusl {
		platformKey = fmt.Sprintf("%s-musl-%s", platform, arch)
	} else {
		platformKey = fmt.Sprintf("%s-%s", platform, arch)
	}
	packageName, ok := uvPackages[platformKey]
	if !ok {
		return fmt.Errorf("no binary available for %s", platformKey)
	}
	binDir := RuntimePath()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	downloadURL := fmt.Sprintf("%s/%s/%s", uvReleaseBaseURL, version, packageName)
	tempDir := os.TempDir()
	tempFilename := filepath.Join(tempDir, packageName)
	fmt.Printf("Downloading uv %s for %s...\n", version, platformKey)
	fmt.Printf("URL: %s\n", downloadURL)
	if err := downloadWithRedirects(downloadURL, tempFilename); err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	fmt.Printf("Extracting %s to %s...\n", packageName, binDir)

	var zr *zip.ReadCloser
	if strings.HasSuffix(packageName, ".zip") {
		zr, err := zip.OpenReader(tempFilename)
		if err != nil {
			// 删除临时文件，在打开失败时
			if err := os.Remove(tempFilename); err != nil {
				return fmt.Errorf("无法删除临时文件: %v", err)
			}
			return fmt.Errorf("无法打开ZIP文件: %v", err)
		}
		defer zr.Close() // 确保ZIP读者被关闭

		for _, file := range zr.File {
			if file.FileInfo().IsDir() {
				continue
			}
			filePath := filepath.Join(binDir, file.Name)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("无法创建目录: %v", err)
			}

			// 打开源文件
			src, err := file.Open()
			if err != nil {
				return fmt.Errorf("无法打开文件 %s: %v", file.Name, err)
			}

			// 创建目标文件
			dest, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("无法创建文件 %s: %v", file.Name, err)
			}

			// 复制文件内容
			if _, err := io.Copy(dest, src); err != nil {
				// 关闭文件并删除目标文件，然后返回错误
				defer src.Close()
				defer dest.Close()
				defer os.Remove(filePath)
				return fmt.Errorf("无法复制文件 %s: %v", file.Name, err)
			}

			// 确保所有数据被写入磁盘
			err = dest.Sync()
			if err != nil {
				defer src.Close()
				defer dest.Close()
				defer os.Remove(filePath)
				return fmt.Errorf("无法同步文件 %s: %v", file.Name, err)
			}

			// 关闭文件
			defer src.Close()
			defer dest.Close()
		}
	} else {
		// 处理.tar.gz文件，代码略
	}

	// Explicitly close the zip reader and any other files before deletion
	if zr != nil {
		if err := zr.Close(); err != nil {
			return fmt.Errorf("无法关闭ZIP读者: %v", err)
		}
	}

	// 现在，尝试删除临时文件
	if err := os.Remove(tempFilename); err != nil {
		return fmt.Errorf("无法删除临时文件: %v", err)
	}

	return nil
}
func detectPlatformAndArchUV() (string, string, bool) {
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
	return platform, arch, isMusl
}
func InstallUv() error {
	version := defaultUVVersion
	fmt.Printf("Using uv version: %s\n", version)
	platform, arch, isMusl := detectPlatformAndArchUV()
	fmt.Printf("Installing uv %s for %s-%s%s...\n",
		version,
		platform,
		arch,
		func() string {
			if isMusl {
				return " (MUSL)"
			}
			return ""
		}())
	return downloadUvBinary(platform, arch, version, isMusl)
}
