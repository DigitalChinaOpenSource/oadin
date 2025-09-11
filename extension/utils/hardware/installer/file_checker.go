package installer

import (
	"os"
	"path/filepath"
	"runtime"

	"oadin/config"
)

func RuntimePath() string {
	return filepath.Join(config.GlobalEnvironment.RootDir, "runtime")
}

func getBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func getBinaryPath(name *string) string {
	if name == nil || *name == "" {
		return RuntimePath()
	}
	binaryName := getBinaryName(*name)
	binariesDir := RuntimePath()
	// Check if the binariesDir exists
	_, err := os.Stat(binariesDir)
	if err == nil {
		return filepath.Join(binariesDir, binaryName)
	} else {
		return binaryName
	}
}

func isBinaryExists(name string) bool {
	binaryPath := getBinaryPath(&name)
	_, err := os.Stat(binaryPath)
	return err == nil
}
