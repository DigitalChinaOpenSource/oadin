//go:build linux

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"

	"byze/config"
)

func ShowLogs() error {
	cmd = exec.Command("xdg-open", config.GlobalByzeEnvironment.LogDir)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
