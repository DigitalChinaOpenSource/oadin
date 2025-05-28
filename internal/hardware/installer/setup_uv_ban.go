package installer

import (
	"fmt"
	"sync"
	"time"
)

// InstallNpxUvProps 定义了InstallNpxUv的属性
type InstallNpxUvProps struct {
	Mini bool
}

// InstallNpxUv 的状态和方法
type InstallNpxUv struct {
	isUvInstalled   bool
	isBunInstalled  bool
	isInstallingUv  bool
	isInstallingBun bool
	uvPath          string
	bunPath         string
	binariesDir     string
	translationFunc func(string) string
	mutex           sync.RWMutex
}

// NewInstallNpxUv 创建一个新的安装组件
func NewInstallNpxUv() *InstallNpxUv {
	return &InstallNpxUv{
		isUvInstalled:  true,
		isBunInstalled: true,
	}
}

// CheckBinaries 检查二进制文件是否存在
func (i *InstallNpxUv) CheckBinaries() error {
	// 模拟检查二进制是否存在
	uvExists, bunExists := i.isBinaryExist()

	i.mutex.Lock()
	i.isUvInstalled = uvExists
	i.isBunInstalled = bunExists
	i.mutex.Unlock()
	return nil
}

// InstallUV 安装 UV 二进制
func (i *InstallNpxUv) InstallUV() error {
	i.mutex.Lock()
	i.isInstallingUv = true
	i.mutex.Unlock()
	defer func() {
		i.mutex.Lock()
		i.isInstallingUv = false
		i.mutex.Unlock()
	}()
	// 在安装UV二进制
	go func() {
		// 模拟安装过程
		fmt.Println("Installing UV...")
		err := InstallUv()
		if err != nil {
			// 处理错误
			fmt.Printf("Error installing UV: %v\n", err)
			return
		}
		fmt.Println("UV installed successfully.")
	}()
	return nil
}

// InstallBun 安装 Bun 二进制
func (i *InstallNpxUv) InstallBun() error {
	i.mutex.Lock()
	i.isInstallingBun = true
	i.mutex.Unlock()
	defer func() {
		i.mutex.Lock()
		i.isInstallingBun = false
		i.mutex.Unlock()
	}()
	// 在安装Bun二进制
	go func() {
		// 模拟安装过程
		fmt.Println("Installing Bun...")
		err := InstallBun()
		if err != nil {
			// 处理错误
			fmt.Printf("Error installing Bun: %v\n", err)
			return
		}
		fmt.Println("Bun installed successfully.")
	}()
	return nil
}

// GetInstallInfo 获取安装信息
func (i *InstallNpxUv) GetInstallInfo() (*InstallInfo, error) {
	// 模拟获取安装信息
	return &InstallInfo{
		UvPath:      i.uvPath,
		BunPath:     i.bunPath,
		BinariesDir: i.binariesDir,
	}, nil
}

// InstallInfo 安装信息
type InstallInfo struct {
	UvPath      string
	BunPath     string
	BinariesDir string
}

// 在另一个 Goroutine 中运行检查
func (i *InstallNpxUv) RunCheckBinaries() {
	go func() {
		if err := i.CheckBinaries(); err != nil {
			// 处理错误
			fmt.Printf("Error checking binaries: %v\n", err)
		}
	}()
}

// isBinaryExist 检查二进制是否存在
func (i *InstallNpxUv) isBinaryExist() (uvExist, bunExist bool) {
	// 检查本地是否存在文件
	return isBinaryExists("uv"), isBinaryExists("bun")
}

// 在安装完成后调用
func (i *InstallNpxUv) setTimeout(f func(), delay time.Duration) {
	time.Sleep(delay)
	f()
}

func SetupBunAndUv() {
	// 创建一个新的安装组件
	installNpxUv := NewInstallNpxUv()
	// 执行检查
	installNpxUv.RunCheckBinaries()
	// 模拟安装UV
	installNpxUv.InstallUV()
	// 模拟安装Bun
	installNpxUv.InstallBun()
}
