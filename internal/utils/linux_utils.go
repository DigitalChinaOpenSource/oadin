//go:build linux

package utils

import (
	"fmt"
	"github.com/shirou/gopsutil/mem"
)

func GetMemoryInfo() (*MemoryInfo, error) {
	v, _ := mem.VirtualMemory()
	memorySize := int(v.Total / 1024 / 1024 / 1024)
	memoryInfo := &MemoryInfo{
		Size: memorySize,
	}
	return memoryInfo, nil
}

func GetSystemVersion() int {
	systemVersion := 0
	return systemVersion
}

func SamePartitionStatus(srcPath, targetPath string) (bool, error) {
	info1, err := os.Stat(srcPath)
	if err != nil {
		return false, err
	}

	info2, err := os.Stat(targetPath)
	if err != nil {
		return false, err
	}

	stat1 := info1.Sys().(*syscall.Stat_t)
	stat2 := info2.Sys().(*syscall.Stat_t)

	return stat1.Dev == stat2.Dev, nil
}

func ModifySystemUserVariables(envInfo *types.EnvVariables) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	shell := os.Getenv("SHELL")
	var rcFile string

	switch {
	case strings.Contains(shell, "zsh"):
		rcFile = ".zshrc"
	case strings.Contains(shell, "bash"):
		rcFile = ".bashrc"
	default:
		rcFile = ".profile" // fallback
	}

	fullPath := currentUser.HomeDir + "/" + rcFile
	variableLine := fmt.Sprintf("export %s=\"%s\"", envInfo.Name, envInfo.Value)

	// check exists
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), variableLine) {
		return nil
	}

	// append
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + variableLine + "\n"); err != nil {
		return err
	}
	return nil
}
