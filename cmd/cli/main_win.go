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
	"log"
	"os"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	cli "oadin/cmd/cli/core"
)

// oadinService 定义 Windows 服务
type oadinService struct{}

// Execute 是 Windows 服务的入口
func (m *oadinService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// 通知服务管理器：启动中
	s <- svc.Status{State: svc.StartPending}

	elog, _ := eventlog.Open("OadinService")
	if elog != nil {
		defer elog.Close()
		elog.Info(1, "OadinService 启动中...")
	}

	// 通知服务管理器：运行中
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// 启动 CLI 命令（参数来自 os.Args，例如 "server start -d"）
	command := cli.NewCommand()

	// goroutine 监听 Stop/Shutdown
	go func() {
		for c := range r {
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				if elog != nil {
					elog.Info(1, "OadinService 收到停止信号，准备退出")
				}
				// ⚠️ 这里用 os.Exit(0) 暴力退出
				// 如果 cli 支持 context，可以在这里发 cancel 让其优雅退出
				os.Exit(0)
			}
		}
	}()

	// 阻塞执行 CLI（直到进程退出）
	if err := command.Execute(); err != nil {
		if elog != nil {
			elog.Error(1, "OadinService 执行失败: "+err.Error())
		}
	}

	// 通知服务管理器：已停止
	s <- svc.Status{State: svc.Stopped}
	return false, 0
}

func main() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("无法判断是否为 Windows 服务: %v", err)
	}

	if isService {
		// 服务模式
		err = svc.Run("OadinService", &oadinService{})
		if err != nil {
			log.Fatalf("服务运行失败: %v", err)
		}
		return
	}

	// 控制台模式（调试用）
	command := cli.NewCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
