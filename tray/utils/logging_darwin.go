//go:build darwin

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"

	"oadin/config"
)

func ShowLogs() error {
	cmd := exec.Command("open", config.GlobalOadinEnvironment.LogDir)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
