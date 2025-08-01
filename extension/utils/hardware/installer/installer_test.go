package installer

import (
	"path/filepath"
	"testing"
)

func TestDownloadBunBinary(t *testing.T) {
	platform := "darwin"
	arch := "arm64"
	version := "1.2.9"
	isMusl := false
	isBaseline := false
	downloadBunBinary(platform, arch, version, isMusl, isBaseline)
}

func TestDownloadBunBinaryWindows(t *testing.T) {
	platform := "windows"
	arch := "amd64"
	version := "1.2.9"
	isMusl := false
	isBaseline := false
	downloadBunBinary(platform, arch, version, isMusl, isBaseline)
}

func TestDownloadUvBinary(t *testing.T) {
	platform := "darwin"
	arch := "arm64"
	version := "0.6.14"
	isMusl := false
	downloadUvBinary(platform, arch, version, isMusl)
}

func TestDownloadUvBinaryWindows(t *testing.T) {
	platform := "windows"
	arch := "amd64"
	version := "0.6.14"
	isMusl := false
	downloadUvBinary(platform, arch, version, isMusl)
}

func TestFileName(t *testing.T) {
	name := "bun-darwin-aarch64/bun"
	res := filepath.Base(name)
	t.Log(res)

	if res != "bun" {
		t.Error("error")
	}

	name = "bun-windows-x64/bun.exe"
	res = filepath.Base(name)
	t.Log(res)
	if res != "bun.exe" {
		t.Error("error")
	}
}
