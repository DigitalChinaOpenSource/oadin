//go:build darwin

package tray

import (
	"byze/config"
	"fmt"
	"log/slog"
	"os/exec"
)

func ShowLogs() error {
	cmd := exec.Command("open", config.GlobalByzeEnvironment.LogDir)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
