package tray

import (
	"byze/config"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"syscall"
)

func ShowLogs() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows implementation (your existing code)
		cmd_path := "c:\\Windows\\system32\\cmd.exe"
		cmd = exec.Command(cmd_path, "/c", "start", config.GlobalByzeEnvironment.LogDir)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false, CreationFlags: 0x08000000}

	case "darwin": // macOS
		// On macOS, we use 'open' command to open the directory in Finder
		cmd = exec.Command("open", config.GlobalByzeEnvironment.LogDir)

	case "linux":
		// On Linux, we try common file managers (xdg-open works on most distributions)
		cmd = exec.Command("xdg-open", config.GlobalByzeEnvironment.LogDir)

	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
