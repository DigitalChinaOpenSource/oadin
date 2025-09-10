//go:build windows

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"
	"path/filepath"
	"oadin/internal/utils"
)

func ShowLogs() error {
	// Windows implementation (your existing code)
	cmd_path := "c:\\Windows\\system32\\cmd.exe"
	rootDir, _ := utils.GetOADINDataDir()
	logPath := filepath.Join(rootDir, "logs")
	cmd := exec.Command(cmd_path, "/c", "start", logPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false, CreationFlags: 0x08000000}

	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
