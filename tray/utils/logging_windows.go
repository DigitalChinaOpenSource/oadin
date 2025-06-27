//go:build windows

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"

	"oadin/config"
)

func ShowLogs() error {
	// Windows implementation (your existing code)
	cmd_path := "c:\\Windows\\system32\\cmd.exe"
	cmd := exec.Command(cmd_path, "/c", "start", config.GlobalOadinEnvironment.LogDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false, CreationFlags: 0x08000000}

	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
