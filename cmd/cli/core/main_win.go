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

package cli

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/windows/svc"
)

// oadinService defines the Windows service
type oadinService struct{}

// Execute is the entry point of the Windows service
func (m *oadinService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// Notify SCM: service is starting
	s <- svc.Status{State: svc.StartPending}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErrChan := make(chan error, 1)

	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// start CLI server
	go func() {
		command := NewCommand()
		command.SetContext(ctx)
		if err := command.Execute(); err != nil {
			runErrChan <- err
		}
	}()

	// Main loop
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				cancel()
				stopCmd := exec.Command("oadin", "server", "stop")
				stopCmd.Run()

				// 等待服务关闭
				select {
				case <-runErrChan:
				case <-time.After(5 * time.Second):
				}

				s <- svc.Status{State: svc.Stopped}
				return false, 0
			}
		case err := <-runErrChan:
			if err != nil {
				s <- svc.Status{State: svc.Stopped}
				return false, 1
			}
		case <-ctx.Done():
			s <- svc.Status{State: svc.Stopped}
			return false, 0
		}
	}
}

func MainPlatform() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("无法检测是否在服务模式下运行: %v", err)
	}

	if isService {
		err = svc.Run("OadinService", &oadinService{})
		if err != nil {
			log.Fatalf("服务运行失败: %v", err)
		}
		return
	}

	// Console 模式（方便调试）
	command := NewCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
