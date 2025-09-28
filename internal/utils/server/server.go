package server

import (
	"context"
	"fmt"
	"net/http"
	"oadin/internal/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"oadin/internal/provider"
	"oadin/internal/types"
	"oadin/internal/utils"
)

func IsServerRunning() bool {
	serverUrl := "http://127.0.0.1:16688" + "/health"
	resp, err := http.Get(serverUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func StartOadinServer(logPath string, pidFilePath string) error {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	execCmd := "oadin.exe"
	if runtime.GOOS != "windows" {
		execCmd = "oadin"
	}
	cmd := exec.Command(execCmd, "server", "start")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if runtime.GOOS == "windows" {
		utils.SetCmdSysProcAttr(cmd)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Oadin server: %v", err)
	}
	// Save PID to file.
	pid := cmd.Process.Pid
	pidFile := filepath.Join(pidFilePath, "oadin.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0o644); err != nil {
		return fmt.Errorf("failed to save PID to file: %v", err)
	}

	logger.EngineLogger.Info(fmt.Sprintf("Oadin server started with PID: %d", cmd.Process.Pid))
	return nil
}

func StopOadinServer(pidFilePath string) error {
	files, err := filepath.Glob(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to list pid files: %v", err)
	}

	if len(files) == 0 {
		logger.EngineLogger.Info("No running processes found")
		return nil
	}
	// stop model engine
	for _, modelEngine := range types.SupportModelEngine {
		engine := provider.GetModelEngine(modelEngine)
		err = engine.StopEngine(context.Background())
		if err != nil {
			logger.EngineLogger.Info(fmt.Sprintf("failed to stop engine %s: %v", modelEngine, err))
		}
		logger.EngineLogger.Info(fmt.Sprintf("Stop engine successfully %s", modelEngine))
	}

	// Traverse all pid files.
	for _, pidFile := range files {
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			logger.EngineLogger.Info(fmt.Sprintf("Failed to read PID file %s: %v", pidFile, err))
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if err != nil {
			logger.EngineLogger.Info(fmt.Sprintf("Invalid PID in file %s: %v", pidFile, err))
			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			logger.EngineLogger.Info(fmt.Sprintf("Failed to find process with PID %d: %v", pid, err))
			continue
		}

		if err := process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				logger.EngineLogger.Info("Process with PID %d is already stopped\n", pid)
			} else {
				logger.EngineLogger.Info("Failed to kill process with PID %d: %v\n", pid, err)
				continue
			}
		} else {
			logger.EngineLogger.Info("Successfully stopped process with PID %d\n", pid)
		}

		// remove pid file
		if err := os.Remove(pidFile); err != nil {
			logger.EngineLogger.Info("Failed to remove PID file %s: %v\n", pidFile, err)
		}
	}
	if runtime.GOOS == "windows" {
		if utils.IpexOllamaSupportGPUStatus() {
			extraProcessName := "ollama-lib.exe"
			extraCmd := exec.Command("taskkill", "/IM", extraProcessName, "/F")
			_, err := extraCmd.CombinedOutput()
			if err != nil {
				logger.EngineLogger.Info("failed to kill process: %s", extraProcessName)
				return nil
			}
			logger.EngineLogger.Info("Successfully killed process: %s\n", extraProcessName)
		}

		ovmsProcessName := "ovms.exe"
		ovmsCmd := exec.Command("taskkill", "/IM", ovmsProcessName, "/F")
		_, err = ovmsCmd.CombinedOutput()
		if err != nil {
			logger.EngineLogger.Info("failed to kill process: %s", ovmsProcessName)
			return nil
		}
		logger.EngineLogger.Info("Successfully killed process: %s\n", ovmsProcessName)

	}

	return nil
}
