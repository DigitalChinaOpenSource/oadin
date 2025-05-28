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
	"darwin-arm64":         "bun-darwin-aarch64.zip",
	"darwin-x64":           "bun-darwin-x64.zip",
	"win32-x64":            "bun-windows-x64.zip",
	"win32-x64-baseline":   "bun-windows-x64-baseline.zip",
	"win32-arm64":          "bun-windows-x64.zip",
	"win32-arm64-baseline": "bun-windows-x64-baseline.zip",
	"linux-x64":            "bun-linux-x64.zip",
	"linux-x64-baseline":   "bun-linux-x64-baseline.zip",
	"linux-arm64":          "bun-linux-aarch64.zip",
	// MUSL variants
	"linux-musl-x64":          "bun-linux-x64-musl.zip",
	"linux-musl-x64-baseline": "bun-linux-x64-musl-baseline.zip",
	"linux-musl-arm64":        "bun-linux-aarch64-musl.zip",
}

func downloadBunBinary(platform, arch, version string, isMusl, isBaseline bool) error {
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
	// Create output directory structure
	binDir := RuntimePath()
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	// Download URL for the specific binary
	downloadURL := fmt.Sprintf("%s/bun-v%s/%s", bunReleaseBaseURL, version, packageName)
	tempDir := os.TempDir()
	tempFilename := filepath.Join(tempDir, packageName)
	// Download the file
	err := DownloadWithRedirects(downloadURL, tempFilename)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	// Extract the zip file
	fmt.Printf("Extracting %s to %s...\n", packageName, binDir)
	r, err := zip.OpenReader(tempFilename)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer r.Close()
	for _, file := range r.File {
		err := os.MkdirAll(filepath.Join(binDir, filepath.Dir(file.Name)), 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		// Extract files from zip
		if file.FileInfo().IsDir() {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file from zip: %v", err)
		}
		defer src.Close()
		dest, err := os.Create(filepath.Join(binDir, file.Name))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dest.Close()
		_, err = io.Copy(dest, src)
		if err != nil {
			return fmt.Errorf("failed to copy file contents: %v", err)
		}
	}
	// Clean up
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
