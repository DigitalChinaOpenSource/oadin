//go:build darwin

package tray

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"oadin/internal/utils"
)

func ShowLogs() error {
	rootDir, _ := utils.GetOADINDataDir()
	logPath := filepath.Join(rootDir, "logs")
	cmd := exec.Command("open", logPath)
	err := cmd.Start()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to open log dir: %s", err))
		return err
	}
	return nil
}
