package hardware

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func getPlatformInfo() (string, string, error) {
	osType := runtime.GOOS
	var shell string
	switch osType {
	case "windows":
		shell = "powershell"
	case "linux", "darwin":
		shell = "bash"
	default:
		return "", "", errors.New("unsupported platform")
	}
	return osType, shell, nil
}

func isNodeInstalled() bool {
	_, err := exec.LookPath("node")
	return err == nil
}
func isNpmInstalled() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func installNodeJS() error {
	osType, shell, err := getPlatformInfo()
	if err != nil {
		return err
	}
	var cmd *exec.Cmd
	switch osType {
	case "windows":
		cmd = exec.Command(shell, "-Command",
			`iwr https://nodejs.org/dist/latest/win-x64/node.exe -OutFile node.msi; `+
				`Start-Process msiexec -ArgumentList '/i node.msi /quiet' -Wait; `+
				`rm node.msi`)
	case "darwin":
		cmd = exec.Command(shell, "-c",
			"curl -o node.pkg https://nodejs.org/dist/latest/node-v24.1.0.pkg && "+
				"sudo installer -pkg node.pkg -target / && "+
				"rm node.pkg")
	case "linux":
		cmd = exec.Command(shell, "-c",
			"curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash - && "+
				"sudo apt-get install -y nodejs")
	default:
		return errors.New("unsupported OS")
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func verifyNpx() error {
	cmd := exec.Command("npx", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npx not available: %v, output: %s", err, output)
	}
	return nil
}

func configureEnvironment() error {
	// Windows 需要单独处理用户级环境变量
	if runtime.GOOS == "windows" {
		userProfile := os.Getenv("USERPROFILE")
		npmGlobalPath := filepath.Join(userProfile, "AppData", "Roaming", "npm")
		return os.Setenv("PATH", os.Getenv("PATH")+";"+npmGlobalPath)
	}
	return nil // Unix-like 系统通常自动配置
}

func InstallNpxEnvironment() error {
	// 1. 检查现有环境
	if isNpmInstalled() && isNodeInstalled() {
		fmt.Println("✅ Node.js environment already exists")
		return verifyNpx() // 只需验证npx
	}
	// 2. 安装Node.js
	fmt.Println("⚠️ Installing Node.js...")
	if err := installNodeJS(); err != nil {
		return fmt.Errorf("failed to install Node.js: %w", err)
	}
	// 3. 配置环境
	if err := configureEnvironment(); err != nil {
		return fmt.Errorf("environment configuration failed: %w", err)
	}
	// 4. 验证安装
	fmt.Println("🔍 Verifying npx...")
	return verifyNpx()
}
