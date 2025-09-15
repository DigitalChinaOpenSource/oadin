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
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"

	cli "oadin/cmd/cli/core"
)

// Service 定义 Windows 服务
type oadinService struct{}

// Execute 方法是服务的入口
func (m *oadinService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	// 通知服务管理器：启动中
	s <- svc.Status{State: svc.StartPending}

	// 打开事件日志
	elog, err := eventlog.Open("OadinService")
	if err != nil {
		log.Printf("无法打开事件日志: %v", err)
	} else {
		elog.Info(1, "服务启动中...")
		defer elog.Close()
	}

	// 通知服务管理器：运行中
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	// 启动原来的 CLI 命令
	done := make(chan struct{})
	go func() {
		command := cli.NewStartApiServerCommand()
		if err := command.Execute(); err != nil {
			if elog != nil {
				elog.Error(1, "服务启动失败："+err.Error())
			}
		}
		close(done) // run 完成后通知
	}()

	stop := false
	for !stop {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				stop = true
				if elog != nil {
					elog.Info(1, "服务收到停止信号")
				}
			default:
				// 忽略其他事件
			}
		case <-time.After(1 * time.Second):
			// 可用于健康检查或周期任务
		case <-done:
			// CLI 命令退出，服务可以停止或保持运行
			stop = true
		}
	}

	s <- svc.Status{State: svc.Stopped}
	return false, 0
}

func main() {
	// 判断是否交互模式
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("无法判断运行模式: %v", err)
	}

	if !isInteractive {
		// 服务模式
		err = svc.Run("OadinService", &oadinService{})
		if err != nil {
			log.Fatalf("服务运行失败: %v", err)
		}
		return
	}

	// 控制台模式（调试用）
	command := cli.NewStartApiServerCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
