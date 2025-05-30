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

func TestFileName(t *testing.T) {
	var name = "bun-darwin-aarch64/bun"
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
