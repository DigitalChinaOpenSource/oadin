//go:build windows

//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	cli "oadin/cmd/cli/core"
)

// oadinService defines the Windows service
type oadinService struct{}

// Execute is the entry point of the Windows service
func (m *oadinService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// Notify the service manager: Starting
	s <- svc.Status{State: svc.StartPending}

	// open EventLog
	elog, err := eventlog.Open("OadinService")
	if err == nil {
		defer elog.Close()
		elog.Info(1, "OadinService 启动中...")
	}

	// Redirect stdout/stderr to a log file to avoid console output
	logDir := "C:\\ProgramData\\Oadin"
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "oadin.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		os.Stdout = logFile
		os.Stderr = logFile
	}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Notify the service manager: Running
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	command := cli.NewCommand()
	command.SetContext(ctx)

	// Goroutine to listen for Stop/Shutdown requests
	go func() {
		for c := range r {
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				if elog != nil {
					elog.Info(1, "OadinService 收到停止信号，准备退出")
				}
				cancel() // Notify the CLI to stop
				// Give the service manager some time to process
				s <- svc.Status{State: svc.StopPending}
			}
		}
	}()

	// Block and execute CLI (until ctx.Done() or command finishes)
	if err := command.Execute(); err != nil {
		if elog != nil {
			elog.Error(1, "OadinService 执行失败: "+err.Error())
		}
	}

	// Wait a short time to ensure resources are released
	time.Sleep(500 * time.Millisecond)

	// Notify the service manager: Stopped
	s <- svc.Status{State: svc.Stopped}
	return false, 0
}

func main() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Unable to determine if running as a Windows service: %v", err)
	}

	if isService {
		err = svc.Run("OadinService", &oadinService{})
		if err != nil {
			log.Fatalf("Service failed to run: %v", err)
		}
		return
	}

	// Console mode (for debugging)
	command := cli.NewCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
