//go:build linux

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"

	"oadin/config"
)

func ShowLogs() error {
	cmd = exec.Command("xdg-open", config.GlobalOADINEnvironment.LogDir)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
