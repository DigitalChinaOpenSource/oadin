//go:build darwin

package utils

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"syscall"

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

func SamePartitionStatus(srcPath string, targetPath string) (bool, error) {
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

func ModifySystemUserVariables(envInfo *EnvVariables) error {
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
	contentBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return err
	}
	content := string(contentBytes)

	lines := strings.Split(content, "\n")
	updated := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export "+envInfo.Name+"=") {
			lines[i] = variableLine // Replace existing line
			updated = true
			break
		}
	}

	if !updated {
		lines = append(lines, variableLine) // Append new line
	}

	newContent := strings.Join(lines, "\n")

	// append
	err = os.WriteFile(fullPath, []byte(newContent), 0o644)
	if err != nil {
		return err
	}

	return nil
}

func SetCmdSysProcAttr(cmd *exec.Cmd) {
	fmt.Printf("not implement")
}
