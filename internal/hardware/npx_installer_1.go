package hardware

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// CheckNpxAvailability 检查系统是否已安装 npx
func CheckNpxAvailability() (bool, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "npx --version")
	} else {
		cmd = exec.Command("bash", "-c", "npx --version")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	// 检查输出是否包含版本号
	version := strings.TrimSpace(string(output))
	return len(version) > 0, nil
}

// CheckNodeAvailability 检查系统是否已安装 Node.js
func CheckNodeAvailability() (bool, string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "node --version")
	} else {
		cmd = exec.Command("bash", "-c", "node --version")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", err
	}

	version := strings.TrimSpace(string(output))
	return len(version) > 0, version, nil
}

func InstallNode() error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// 使用PowerShell下载并运行Windows安装程序
		cmd = exec.Command("powershell", "-Command",
			"Invoke-WebRequest https://nodejs.org/dist/latest-v16.x/node-v16.x.x-x64.msi -OutFile node.msi; Start-Process msiexec.exe -Wait -ArgumentList '/i node.msi /quiet /norestart'")
	case "darwin": // macOS
		// 使用Homebrew安装
		cmd = exec.Command("bash", "-c", "brew install node")
	case "linux":
		// 使用apt或yum安装
		// 需要区分不同的Linux发行版
		if isDebianBased() {
			cmd = exec.Command("bash", "-c", "sudo apt-get update && sudo apt-get install -y nodejs npm")
		} else if isRedHatBased() {
			cmd = exec.Command("bash", "-c", "sudo yum install -y nodejs npm")
		} else {
			return fmt.Errorf("unsupported Linux distribution")
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Run()
}

// 帮助函数，检测Linux发行版类型
func isDebianBased() bool {
	cmd := exec.Command("bash", "-c", "command -v apt-get")
	return cmd.Run() == nil
}
func isRedHatBased() bool {
	cmd := exec.Command("bash", "-c", "command -v yum")
	return cmd.Run() == nil
}

// SetupNpxEnvironment 检查并设置 npx 环境
func SetupNpxEnvironment(progressCallback func(stage string, percent int)) error {
	progressCallback("检查 Node.js 环境", 10)

	// 检查 Node.js
	nodeAvailable, nodeVersion, err := CheckNodeAvailability()
	if err != nil || !nodeAvailable {
		progressCallback("未检测到 Node.js，开始安装", 20)
		if err := InstallNode(); err != nil {
			return fmt.Errorf("安装 Node.js 失败: %w", err)
		}
		progressCallback("Node.js 安装完成", 50)
	} else {
		progressCallback("检测到 Node.js: "+nodeVersion, 30)
	}

	// 检查 npx
	progressCallback("检查 npx 环境", 60)
	npxAvailable, err := CheckNpxAvailability()
	if err != nil || !npxAvailable {
		progressCallback("未检测到 npx，开始安装", 70)
		if err := InstallNpx(); err != nil {
			return fmt.Errorf("安装 npx 失败: %w", err)
		}
		progressCallback("npx 安装完成", 90)
	} else {
		progressCallback("检测到 npx 环境", 80)
	}

	// 再次验证
	npxAvailable, err = CheckNpxAvailability()
	if err != nil || !npxAvailable {
		return fmt.Errorf("npx 安装后验证失败")
	}

	progressCallback("环境检查完成", 100)
	return nil
}

// InstallNpx 安装 npx
func InstallNpx() error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "npm install -g npx")
	} else {
		cmd = exec.Command("bash", "-c", "npm install -g npx")
	}

	return cmd.Run()
}
