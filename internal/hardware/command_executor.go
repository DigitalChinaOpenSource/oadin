package hardware

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// Platform 声明操作系统类型
type Platform int

const (
	Unknown Platform = iota
	Windows
	Darwin // macOS
	Linux
)

// GetPlatform 获取当前操作系统类型
func GetPlatform() Platform {
	switch runtime.GOOS {
	case "windows":
		return Windows
	case "darwin":
		return Darwin
	case "linux":
		return Linux
	default:
		return Unknown
	}
}

// CommandBuilder 跨平台命令构建器
type CommandBuilder struct {
	command    string
	args       []string
	timeout    time.Duration
	workingDir string
	envVars    []string
}

// NewCommandBuilder 创建命令构建器
func NewCommandBuilder(command string) *CommandBuilder {
	return &CommandBuilder{
		command: command,
		envVars: os.Environ(),
	}
}

// WithArgs 添加命令参数
func (cb *CommandBuilder) WithArgs(args ...string) *CommandBuilder {
	cb.args = append(cb.args, args...)
	return cb
}

// WithTimeout 设置超时时间
func (cb *CommandBuilder) WithTimeout(timeout time.Duration) *CommandBuilder {
	cb.timeout = timeout
	return cb
}

// WithWorkingDir 设置工作目录
func (cb *CommandBuilder) WithWorkingDir(dir string) *CommandBuilder {
	cb.workingDir = dir
	return cb
}

// WithEnv 添加环境变量
func (cb *CommandBuilder) WithEnv(key, value string) *CommandBuilder {
	cb.envVars = append(cb.envVars, fmt.Sprintf("%s=%s", key, value))
	return cb
}

// Execute 执行命令
func (cb *CommandBuilder) Execute() (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cb.timeout)
	defer cancel()

	// 构建平台特定命令
	fullCmd, args := cb.buildPlatformCommand()

	// 创建命令对象
	cmd := exec.CommandContext(ctx, fullCmd, args...)

	// 配置环境变量
	cmd.Env = cb.envVars

	// 设置工作目录
	if cb.workingDir != "" {
		cmd.Dir = cb.workingDir
	}

	// 处理输入输出
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		return "", "", fmt.Errorf("启动命令失败: %w", err)
	}

	// 等待执行完成
	if err := cmd.Wait(); err != nil {
		return stdout.String(), stderr.String(), fmt.Errorf("执行出错: %w", err)
	}

	return stdout.String(), stderr.String(), nil
}

// buildPlatformCommand 构建平台特定命令
func (cb *CommandBuilder) buildPlatformCommand() (string, []string) {
	platform := GetPlatform()

	switch platform {
	case Windows:
		// Windows 使用 cmd.exe /C 执行
		return "cmd", append([]string{"/C", cb.command}, cb.args...)
	case Darwin, Linux:
		// macOS/Linux 使用 /bin/sh -c 执行
		return "/bin/sh", append([]string{"-c", cb.command}, cb.args...)
	default:
		panic("不支持的操作系统")
	}
}
