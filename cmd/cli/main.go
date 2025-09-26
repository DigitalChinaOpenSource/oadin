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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	cli "oadin/cmd/cli/core"
)

func main() {
	command := cli.NewCommand()

	// 只有在执行 server start 命令时才检查管理员权限
	if len(os.Args) >= 3 && os.Args[1] == "server" && os.Args[2] == "start" {
		// 检查是否已经是管理员权限
		err := exec.Command("net", "session").Run()
		if err != nil {
			fmt.Println("检测到需要管理员权限才能启动服务")
			fmt.Print("是否要以管理员身份重新启动? (y/N): ")

			var response string
			fmt.Scanln(&response)

			if response == "y" || response == "Y" || response == "yes" || response == "Yes" {
				fmt.Println("正在以管理员权限重新启动...")
				exe, _ := os.Executable()

				// 构建完整的参数列表
				args := os.Args[1:] // 除了程序名之外的所有参数
				argString := ""
				for i, arg := range args {
					if i > 0 {
						argString += " "
					}
					argString += arg
				}

				// 使用 CMD 的 runas 命令而不是 PowerShell
				cmd := exec.Command("cmd", "/c", "runas", "/user:Administrator",
					fmt.Sprintf("\"%s\" %s", exe, argString))

				// 或者使用更简单的方式，直接调用带提权的批处理
				// 创建临时批处理文件来实现提权
				batchContent := fmt.Sprintf(`@echo off
cd /d "%s"
"%s" %s
pause`, filepath.Dir(exe), exe, argString)

				tempBat := filepath.Join(os.TempDir(), "oadin_elevate.bat")
				err = os.WriteFile(tempBat, []byte(batchContent), 0644)
				if err == nil {
					// 使用 runas 执行批处理文件
					cmd = exec.Command("runas", "/user:Administrator", tempBat)
					defer os.Remove(tempBat) // 清理临时文件
				}

				err = cmd.Run()
				if err != nil {
					fmt.Printf("提权失败: %v\n", err)
					fmt.Println("请手动以管理员身份运行此程序")
					fmt.Println("或者在 CMD 中使用: runas /user:Administrator \"oadin.exe server start\"")
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				fmt.Println("用户取消了提权操作，程序将以当前权限继续运行")
				fmt.Println("注意：某些功能可能需要管理员权限才能正常工作")
			}
		}
	}

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
