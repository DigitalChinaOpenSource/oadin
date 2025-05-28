package hardware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	minPythonVersion = 3.6
	minUVVersion     = 0.1
)

// EnvManager 环境管理器
type EnvManager struct {
	pythonPath string
	uvPath     string
	logger     Logger // 自定义日志接口
}
type Logger interface {
	Info(msg string)
	Debug(msg string)
	Error(msg string)
}

// NewEnvManager 创建环境管理器实例
func NewEnvManager() *EnvManager {
	return &EnvManager{
		logger: &SimpleLogger{},
	}
}

type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string)  { fmt.Println("[INFO]", msg) }
func (l *SimpleLogger) Debug(msg string) { fmt.Println("[DEBUG]", msg) }
func (l *SimpleLogger) Error(msg string) { fmt.Println("[ERROR]", msg) }

// CheckAndInstall 主入口：检查并安装所需环境
func (m *EnvManager) InstallUVWithPython() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// 1. 检查Python
	if err := m.checkPython(ctx); err != nil {
		m.logger.Info("Python environment not found, attempting installation...")
		if err := m.installPython(ctx); err != nil {
			return fmt.Errorf("Python installation failed: %v", err)
		}
	}
	// 2. 检查UV
	if err := m.checkUV(ctx); err != nil {
		m.logger.Info("UV not found, attempting installation...")
		if err := m.installUV(ctx); err != nil {
			return fmt.Errorf("UV installation failed: %v", err)
		}
	}
	m.logger.Info("All dependencies are ready!")
	return nil
}

// checkPython 检查Python环境
func (m *EnvManager) checkPython(ctx context.Context) error {
	m.logger.Debug("Checking Python installation...")
	// 首先尝试从缓存中获取
	if m.pythonPath != "" {
		if err := m.verifyPythonVersion(ctx, m.pythonPath); err == nil {
			return nil
		}
	}
	// 检查常见Python可执行文件名称
	candidates := []string{"python3", "python"}
	if runtime.GOOS == "windows" {
		candidates = append([]string{"python"}, candidates...) // Windows下优先尝试python.exe
	}
	for _, name := range candidates {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		if err := m.verifyPythonVersion(ctx, path); err != nil {
			m.logger.Debug(fmt.Sprintf("Found Python at %s but version check failed: %v", path, err))
			continue
		}
		m.pythonPath = path
		m.logger.Info(fmt.Sprintf("Python found at: %s", path))
		return nil
	}
	m.pythonPath = ""
	return errors.New("no valid Python installation found")
}

// verifyPythonVersion 验证Python版本
func (m *EnvManager) verifyPythonVersion(ctx context.Context, pythonPath string) error {
	// 测试多个功能以确保完整安装
	tests := []struct {
		args   []string
		check  func(string) bool
		errMsg string
	}{
		{
			args:   []string{"--version"},
			check:  func(out string) bool { return strings.Contains(out, "Python 3") },
			errMsg: "version check failed",
		},
		{
			args:   []string{"-c", "import ssl; print(ssl.OPENSSL_VERSION)"},
			check:  func(out string) bool { return strings.Contains(out, "OpenSSL") },
			errMsg: "SSL模块缺失",
		},
		{
			args:   []string{"-c", "import pip; print(pip.__version__)"},
			check:  func(out string) bool { return len(out) > 0 },
			errMsg: "pip未安装",
		},
	}
	for _, test := range tests {
		cmd := exec.CommandContext(ctx, pythonPath, test.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Python功能测试失败[%s]: %v", test.errMsg, err)
		}
		if !test.check(string(output)) {
			return fmt.Errorf("Python完整性检查未通过[%s]: %s", test.errMsg, string(output))
		}
	}
	return nil
}

// installPython 安装Python
func (m *EnvManager) installPython(ctx context.Context) error {
	m.logger.Info("Starting Python installation...")

	var err error
	switch runtime.GOOS {
	case "windows":
		err = m.installPythonWindows(ctx)
	case "darwin":
		err = m.installPythonMac(ctx)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	if err != nil {
		return err
	}
	// 重新检查安装是否成功
	return m.checkPython(ctx)
}

// windows 安裝python
func (m *EnvManager) installPythonWindows(ctx context.Context) error {
	// 1. 检查现有安装残留
	if err := m.cleanPythonWindows(ctx); err != nil {
		m.logger.Error("清理旧版本失败: " + err.Error())
	}
	// 2. 优先级1: 使用winget安装
	m.logger.Info("尝试通过winget安装...")
	if err := m.runCommand(ctx, "winget", "install", "--id=Python.Python.3.10", "-e"); err == nil {
		return m.postInstallWindows(ctx)
	}
	// 3. 优先级2: 下载官方安装包
	m.logger.Info("通过离线安装包安装...")
	tempDir, err := os.MkdirTemp("", "python_install")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	installerPath := filepath.Join(tempDir, "python_installer.exe")
	if err := m.downloadPythonWindows(ctx, installerPath); err != nil {
		return err
	}
	// 使用精确的安装参数
	installCmd := fmt.Sprintf(
		`%s /quiet InstallAllUsers=1 PrependPath=1 Include_launcher=1 Include_test=0`,
		installerPath)

	if output, err := exec.CommandContext(ctx, "cmd", "/C", installCmd).CombinedOutput(); err != nil {
		return fmt.Errorf("安装失败: %v\n输出: %s", err, string(output))
	}
	return m.postInstallWindows(ctx)
}
func (m *EnvManager) downloadPythonWindows(ctx context.Context, dest string) error {
	// 根据架构动态选择版本
	arch := "amd64"
	if runtime.GOARCH == "386" {
		arch = "win32"
	}
	url := fmt.Sprintf(
		"https://www.python.org/ftp/python/3.10.11/python-3.10.11-%s.exe",
		arch)

	return m.downloadFile(ctx, url, dest)
}
func (m *EnvManager) postInstallWindows(ctx context.Context) error {
	// 确保Python出现在PATH中
	paths := []string{
		`C:\Program Files\Python310`,
		`C:\Program Files\Python310\Scripts`,
		`C:\Users\` + os.Getenv("USERNAME") + `\AppData\Roaming\Python\Python310`,
		`C:\Users\` + os.Getenv("USERNAME") + `\AppData\Roaming\Python\Python310\Scripts`,
	}
	// 更新当前进程的环境变量
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			os.Setenv("PATH", fmt.Sprintf("%s;%s", p, os.Getenv("PATH")))
		}
	}
	// 验证安装结果
	return m.checkPython(ctx)
}
func (m *EnvManager) cleanPythonWindows(ctx context.Context) error {
	// 卸载所有现有版本
	if _, err := exec.LookPath("where"); err == nil {
		if output, err := exec.CommandContext(
			ctx, "where", "python.exe").CombinedOutput(); err == nil {

			for _, path := range strings.Split(string(output), "\n") {
				path = strings.TrimSpace(path)
				if path != "" {
					m.logger.Info("发现已安装的Python: " + path)
					uninstallCmd := fmt.Sprintf(
						`wmic product where "name like 'Python%%'" call uninstall /nointeractive`)

					exec.CommandContext(ctx, "cmd", "/C", uninstallCmd).Run()
				}
			}
		}
	}

	// 清理残留目录
	dirs := []string{
		`C:\Python27`,
		`C:\Python310`,
		`C:\Program Files\Python310`,
		filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Programs", "Python"),
	}

	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
	return nil
}

// mac 安裝python
func (m *EnvManager) installPythonMac(ctx context.Context) error {
	// 确保Homebrew是最新版本
	if err := m.runCommand(ctx, "brew", "update"); err != nil {
		m.logger.Error("brew更新失败: " + err.Error())
	}
	// 检查arch - 如果是M1芯片需要特殊处理
	if output, err := exec.CommandContext(ctx, "uname", "-m").CombinedOutput(); err == nil {
		if strings.Contains(string(output), "arm64") {
			// 确保在Rosetta下运行
			os.Setenv("PATH", "/usr/local/bin:"+os.Getenv("PATH"))
		}
	}
	// 安装核心依赖
	if err := m.runCommand(ctx, "brew", "install", "openssl readline sqlite3 xz zlib"); err != nil {
		return err
	}
	// 使用pyenv获得更干净的安装
	if _, err := exec.LookPath("pyenv"); err == nil {
		return m.installWithPyenvMac(ctx)
	}

	// 回退到brew直接安装
	if err := m.runCommand(ctx, "brew", "install", "python@3.10"); err != nil {
		return err
	}
	// 处理M1芯片的特殊路径
	if output, err := exec.CommandContext(ctx, "uname", "-m").CombinedOutput(); err == nil {
		if strings.Contains(string(output), "arm64") {
			pythonPath := "/opt/homebrew/bin/python3"
			if _, err := os.Stat(pythonPath); err == nil {
				m.pythonPath = pythonPath
			}
		}
	}
	return m.checkPython(ctx)
}
func (m *EnvManager) installWithPyenvMac(ctx context.Context) error {
	version := "3.10.11"
	if err := m.runCommand(ctx, "pyenv", "install", "--skip-existing", version); err != nil {
		return err
	}

	// 设置全局Python
	if err := m.runCommand(ctx, "pyenv", "global", version); err != nil {
		return err
	}

	// 更新PATH
	pyenvPath := filepath.Join(os.Getenv("HOME"), ".pyenv", "shims")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", pyenvPath, os.Getenv("PATH")))

	return m.checkPython(ctx)
}

func (m *EnvManager) installPythonDebian(ctx context.Context) error {
	// 安装编译依赖
	if err := m.runCommand(ctx, "sudo", "apt-get", "update"); err != nil {
		return err
	}
	deps := []string{
		"build-essential", "zlib1g-dev", "libncurses5-dev",
		"libgdbm-dev", "libnss3-dev", "libssl-dev",
		"libreadline-dev", "libffi-dev", "libsqlite3-dev",
		"wget", "libbz2-dev",
	}
	if err := m.runCommand(ctx, "sudo", append([]string{"apt-get", "install", "-y"}, deps...)...); err != nil {
		return err
	}
	// 使用dead snakes PPA
	if err := m.runCommand(ctx, "sudo", "add-apt-repository", "-y", "ppa:deadsnakes/ppa"); err == nil {
		m.runCommand(ctx, "sudo", "apt-get", "update")
		if err := m.runCommand(ctx, "sudo", "apt-get", "install", "-y", "python3.10-full"); err == nil {
			return m.checkPython(ctx)
		}
	}
	// 回退到源码编译
	return m.buildPythonFromSource(ctx)
}
func (m *EnvManager) buildPythonFromSource(ctx context.Context) error {
	version := "3.10.11"
	tempDir, err := os.MkdirTemp("", "python_build")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	// 下载源码
	url := fmt.Sprintf("https://www.python.org/ftp/python/%s/Python-%s.tgz", version, version)
	archivePath := filepath.Join(tempDir, "python.tgz")
	if err := m.downloadFile(ctx, url, archivePath); err != nil {
		return err
	}
	// 解压和编译
	commands := [][]string{
		{"tar", "-xzf", archivePath, "-C", tempDir},
		{"sh", "-c", fmt.Sprintf("cd %s && ./configure --enable-optimizations --with-ensurepip=install", filepath.Join(tempDir, fmt.Sprintf("Python-%s", version)))},
		{"sh", "-c", fmt.Sprintf("cd %s && make -j$(nproc)", filepath.Join(tempDir, fmt.Sprintf("Python-%s", version)))},
		{"sh", "-c", fmt.Sprintf("cd %s && sudo make altinstall", filepath.Join(tempDir, fmt.Sprintf("Python-%s", version)))},
	}
	for _, cmd := range commands {
		if err := m.runCommand(ctx, cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}
	// 查找安装路径
	if path, err := exec.LookPath(fmt.Sprintf("python%s", strings.Join(strings.Split(version, ".")[:2], "."))); err == nil {
		m.pythonPath = path
		return nil
	}
	return errors.New("无法找到安装后的Python")
}

// checkUV 检查UV环境
func (m *EnvManager) checkUV(ctx context.Context) error {
	m.logger.Debug("Checking UV installation...")

	// 优先使用已知的Python环境检查
	if m.pythonPath != "" {
		if path, err := m.checkUVPip(ctx, m.pythonPath); err == nil {
			m.uvPath = path
			return nil
		}
	}

	// 检查全局安装的UV
	if path, err := exec.LookPath("uv"); err == nil {
		if err := m.verifyUVVersion(ctx, path); err == nil {
			m.uvPath = path
			return nil
		}
	}

	return errors.New("UV not found")
}

// checkUVPip 通过pip检查UV
func (m *EnvManager) checkUVPip(ctx context.Context, pythonPath string) (string, error) {
	// 检查是否通过pip安装过uv
	cmd := exec.CommandContext(ctx, pythonPath, "-m", "pip", "show", "uv")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pip check failed: %v", err)
	}

	if !bytes.Contains(output, []byte("Name: uv")) {
		return "", errors.New("uv not installed via pip")
	}

	// 获取uv可执行文件位置
	uvPath := filepath.Join(filepath.Dir(pythonPath), "uv")
	if runtime.GOOS == "windows" {
		uvPath += ".exe"
	}

	if _, err := os.Stat(uvPath); err != nil {
		return "", fmt.Errorf("uv binary not found at expected path: %s", uvPath)
	}

	return uvPath, nil
}

// verifyUVVersion 验证UV版本
func (m *EnvManager) verifyUVVersion(ctx context.Context, uvPath string) error {
	cmd := exec.CommandContext(ctx, uvPath, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check UV version: %v", err)
	}

	m.logger.Debug(fmt.Sprintf("UV version output: %s", string(output)))

	// 简化版版本检查 (实际应根据UV的版本输出格式调整)
	if strings.Contains(string(output), "uv") {
		return nil
	}
	return errors.New("invalid UV version")
}

// installUV 安装UV
func (m *EnvManager) installUV(ctx context.Context) error {
	if m.pythonPath == "" {
		return errors.New("Python not found, cannot install UV")
	}

	m.logger.Info("Installing UV via pip...")

	// 使用系统Python或已安装的Python安装UV
	cmd := exec.CommandContext(ctx, m.pythonPath, "-m", "pip", "install", "--user", "uv")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("UV installation failed: %v\nOutput: %s", err, string(output))
	}

	// 重新检查安装是否成功
	return m.checkUV(ctx)
}

// runCommand 辅助函数：执行命令
func (m *EnvManager) runCommand(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("command failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// downloadFile 使用标准net/http实现(适合小型文件)
func (m *EnvManager) downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request failed: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}
	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create file failed: %v", err)
	}
	defer outFile.Close()
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		os.Remove(dest) // 删除部分下载的文件
		return fmt.Errorf("write file failed: %v", err)
	}
	return nil
}
