package hardware

import (
	"bytes"
	"byze/internal/hardware/installer"
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

// CommandBuilder 跨平台命令构建器，支持多种执行器
type CommandBuilder struct {
	command     string
	args        []string
	timeout     time.Duration
	workingDir  string
	envVars     []string
	executorDir string // 执行器所在的目录，例如node_modules/.bin
}

// NewCommandBuilder 创建命令构建器，支持指定执行器类型
func NewCommandBuilder(command string) *CommandBuilder {
	builder := &CommandBuilder{
		command:     command,
		envVars:     os.Environ(),
		executorDir: installer.RuntimePath(),
		workingDir:  "",
	}
	return builder
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
	// 根据执行器类型构建命令
	var cmdExec string
	var cmdArgs []string
	var err error
	switch cb.command {
	case "npx":
		fallthrough
	case "bun":
		cmdExec, cmdArgs, err = cb.buildNpxCommand()
	case "uv":
		cmdExec, cmdArgs, err = cb.buildUVCommand()
	case "uvx":
		cmdExec, cmdArgs, err = cb.buildUVXCommand()
	default:
		return "", "", fmt.Errorf("不支持的执行器类型: %v", cb.command)
	}
	if err != nil {
		return "", "", fmt.Errorf("构建命令失败: %w", err)
	}
	// 创建命令对象
	cmd := exec.CommandContext(ctx, cmdExec, cmdArgs...)
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

// buildNpxCommand 构建npx命令
func (cb *CommandBuilder) buildNpxCommand() (string, []string, error) {
	platform := GetPlatform()
	basePath := cb.executorDir
	if basePath == "" {
		basePath = "node_modules/.bin"
	}

	// 处理'-y'参数
	prependIfNotContains(&cb.args, "-y")

	// 处理'x'参数
	prependIfNotContains(&cb.args, "x")
	switch platform {
	case Windows:
		npxPath := fmt.Sprintf("%s\\bun.exe", basePath)
		return npxPath, cb.args, nil
	case Darwin, Linux:
		npxPath := fmt.Sprintf("%s/bun", basePath)
		return npxPath, cb.args, nil
	default:
		return "", nil, fmt.Errorf("不支持的操作系统: %v", platform)
	}
}

// buildUVCommand 构建uv命令
func (cb *CommandBuilder) buildUVCommand() (string, []string, error) {
	platform := GetPlatform()
	basePath := cb.executorDir
	if basePath == "" {
		basePath = "venv/bin"
	}
	switch platform {
	case Windows:
		// 假设uv在Windows中的可执行路径为uv.exe
		uvPath := fmt.Sprintf("%s\\uv.exe", basePath)
		return uvPath, cb.args, nil
	case Darwin, Linux:
		// 假设uv在Unix-like系统中的可执行路径为uv
		uvPath := fmt.Sprintf("%s/uv", basePath)
		return uvPath, cb.args, nil
	default:
		return "", nil, fmt.Errorf("不支持的操作系统: %v", platform)
	}
}

// buildUVCommand 构建uv命令
func (cb *CommandBuilder) buildUVXCommand() (string, []string, error) {
	platform := GetPlatform()
	basePath := cb.executorDir
	if basePath == "" {
		basePath = "venv/bin"
	}
	switch platform {
	case Windows:
		// 假设uv在Windows中的可执行路径为uv.exe
		uvPath := fmt.Sprintf("%s\\uvx.exe", basePath)
		return uvPath, cb.args, nil
	case Darwin, Linux:
		// 假设uv在Unix-like系统中的可执行路径为uv
		uvPath := fmt.Sprintf("%s/uvx", basePath)
		return uvPath, cb.args, nil
	default:
		return "", nil, fmt.Errorf("不支持的操作系统: %v", platform)
	}
}

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

// 定义一个辅助函数，检查切片中是否包含指定元素
func contains(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// 在切片开头添加指定元素，如果元素不存在
func prependIfNotContains(args *[]string, element string) {
	if args != nil && len(*args) > 0 {
		if !contains(*args, element) {
			// 使用append在开头添加元素
			*args = append([]string{element}, *args...)
		}
	}
}
