package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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
	"darwin-arm64":  "uv-aarch64-apple-darwin.tar.gz",
	"darwin-x64":    "uv-x86_64-apple-darwin.tar.gz",
	"win32-arm64":   "uv-aarch64-pc-windows-msvc.zip",
	"win32-ia32":    "uv-i686-pc-windows-msvc.zip",
	"win32-x64":     "uv-x86_64-pc-windows-msvc.zip",
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
	if strings.HasSuffix(packageName, ".zip") {
		zr, err := zip.OpenReader(tempFilename)
		if err != nil {
			return fmt.Errorf("failed to open zip file: %v", err)
		}
		defer zr.Close()
		for _, file := range zr.File {
			if file.FileInfo().IsDir() {
				continue
			}
			filePath := filepath.Join(binDir, file.Name)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			src, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open file from zip: %v", err)
			}
			defer src.Close()
			dest, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer dest.Close()
			if _, err := io.Copy(dest, src); err != nil {
				return fmt.Errorf("failed to copy file contents: %v", err)
			}
		}
	} else {
		// Extract .tar.gz
		src, err := os.Open(tempFilename)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer src.Close()
		gr, err := gzip.NewReader(src)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gr.Close()
		tr := tar.NewReader(gr)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read tar entry: %v", err)
			}
			filePath := filepath.Join(binDir, header.Name)
			fileMode := header.FileInfo().Mode()
			if header.FileInfo().IsDir() {
				if err := os.MkdirAll(filePath, fileMode); err != nil {
					return fmt.Errorf("failed to create directory: %v", err)
				}
				continue
			}
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			file, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer file.Close()
			if _, err := io.Copy(file, tr); err != nil {
				return fmt.Errorf("failed to write file: %v", err)
			}
			if platform != "windows" {
				if err := os.Chmod(filePath, fileMode); err != nil {
					fmt.Printf("Warning: Failed to set executable permissions for %s: %v\n", filePath, err)
				}
			}
		}
	}
	// Clean up
	if err := os.Remove(tempFilename); err != nil {
		return fmt.Errorf("failed to remove temporary file: %v", err)
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
