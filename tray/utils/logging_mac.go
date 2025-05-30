//go:build darwin

package tray

import (
	"byze/config"
	"fmt"
	"log/slog"
	"os/exec"
	"syscall"
)

func ShowLogs() error {
	cmd_path := "c:\\Windows\\system32\\cmd.exe"
	cmd := exec.Command(cmd_path, "/c", "start", config.GlobalByzeEnvironment.LogDir)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false, CreationFlags: 0x08000000}
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
}
