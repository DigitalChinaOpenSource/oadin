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
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	cli "oadin/cmd/cli/core"
)

// oadinService defines the Windows service
type oadinService struct{}

// Execute is the entry point of the Windows service
func (m *oadinService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// Notify SCM: service is starting
	s <- svc.Status{State: svc.StartPending}

	// open EventLog
	elog, err := eventlog.Open("OadinService")
	if err == nil {
		defer elog.Close()
		elog.Info(1, "OadinService 启动中...")
	}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runErrChan := make(chan error, 1)

	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	if elog != nil {
		elog.Info(1, "OadinService 已进入运行状态（后台继续初始化）")
	}

	// start CLI server
	go func() {
		command := cli.NewCommand()
		command.SetContext(ctx)

		if elog != nil {
			elog.Info(1, "OadinService 开始执行 CLI 服务器...")
		}

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
				if elog != nil {
					elog.Info(1, "OadinService 收到停止信号，正在关闭...")
				}
				s <- svc.Status{State: svc.StopPending}
				cancel()

				// 等待服务关闭
				select {
				case <-runErrChan:
				case <-time.After(5 * time.Second):
				}

				s <- svc.Status{State: svc.Stopped}
				if elog != nil {
					elog.Info(1, "OadinService 已停止")
				}
				return false, 0
			}
		case err := <-runErrChan:
			if elog != nil {
				elog.Error(1, "OadinService 出现错误: "+err.Error())
			}
			s <- svc.Status{State: svc.Stopped}
			return false, 1
		case <-ctx.Done():
			if elog != nil {
				elog.Info(1, "OadinService 上下文取消，准备退出")
			}
			s <- svc.Status{State: svc.Stopped}
			return false, 0
		}
	}
}

func main() {
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
	command := cli.NewCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
