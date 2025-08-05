package tray

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"oadin/config"
	"oadin/internal/utils"
	trayTemplate "oadin/tray/icon"
	tray "oadin/tray/utils"

	"github.com/getlantern/systray/example/icon"
	"github.com/pkg/browser"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

// Manager handles the system tray functionality
type Manager struct {
	serverRunning   bool
	updateAvailable bool
	mRestartUpdate  *systray.MenuItem
	execPath        string
	logPath         string
	pidPath         string
}

// NewManager creates a new system tray manager
func NewManager(debug bool, logPath, pidPath string) *Manager {
	execPath, _ := os.Executable()
	return &Manager{
		serverRunning:   false,
		updateAvailable: false,
		execPath:        execPath,
		logPath:         logPath,
		pidPath:         pidPath,
	}
}

// Start initializes the system tray
func (m *Manager) Start() {
	// 检查服务器状态
	m.serverRunning = utils.IsServerRunning()

	// 不自动启动服务，让用户手动点击
	systray.Run(m.onReady, m.onExit)
}

// 启动 Web Console 前端服务（如有需要）
func startWebConsoleFrontend() {
	// 检查 16699 端口是否已被占用，避免重复启动
	if isPortInUse("16699") {
		return
	}
	// 假设前端服务在 frontend/app/webConsole 目录下，使用 pnpm 启动
	frontendPath := "frontend/app/webConsole"
	cmd := exec.Command("pnpm", "dev")
	cmd.Dir = frontendPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
}

// 检查端口是否已被占用
func isPortInUse(port string) bool {
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func (m *Manager) onReady() {
	// Set icon
	data, err := getIcon()
	if err != nil {
		data = icon.Data
	}
	systray.SetIcon(data)
	systray.SetTitle("Oadin")
	systray.SetTooltip("Oadin AI Service Manager")

	// Add menu items
	mStartStop := systray.AddMenuItem("Start Server", "Start/Stop Oadin server")
	systray.AddSeparator()
	mConsole := systray.AddMenuItem("Web Console", "Open Control Panel")
	m.mRestartUpdate = systray.AddMenuItem("Restart and Update", "Restart and install update")
	if !m.updateAvailable {
		m.mRestartUpdate.Hide()
	}
	mViewLogs := systray.AddMenuItem("View Logs", "View Logs")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Oadin")

	// Update initial server status
	m.updateStartStopMenuItem(mStartStop)

	// Handle menu item clicks
	go func() {
		for {
			select {
			case <-mStartStop.ClickedCh:
				m.handleStartStop()
				m.updateStartStopMenuItem(mStartStop)

			case <-mConsole.ClickedCh:
				m.handleOpenConsole()

			case <-m.mRestartUpdate.ClickedCh:
				if confirmed := dialog.Message("This will stop all servers and install the update. Continue?").Title("Confirm Update").YesNo(); confirmed {
					err := utils.StopOADINServer(filepath.Join(m.pidPath, "oadin.pid"))
					if err != nil {
						dialog.Message("Failed to stop server: %v", err).Title("Error").Error()
					}
					if err := m.performUpdate(); err != nil {
						dialog.Message("Failed to perform update: %v", err).Title("Error").Error()
					} else {
						systray.Quit()
						os.Exit(0)
					}
				}
			case <-mViewLogs.ClickedCh:
				err := m.viewLogs()
				if err != nil {
					dialog.Message("Failed to view logs: %v", err).Title("Error").Error()
				}
			case <-mQuit.ClickedCh:
				if m.serverRunning {
					if confirmed := dialog.Message("Server is still running. Do you want to stop it and quit?").Title("Confirm Quit").YesNo(); confirmed {
						err := utils.StopOADINServer(filepath.Join(m.pidPath, "oadin.pid"))
						if err != nil {
							dialog.Message("Failed to stop server: %v", err).Title("Error").Error()
						}
						systray.Quit()
						return
					}
				} else {
					if confirmed := dialog.Message("Are you sure you want to quit Oadin?").Title("Confirm Quit").YesNo(); confirmed {
						systray.Quit()
						return
					}
				}
			}
		}
	}()
}

// 处理启动/停止服务
func (m *Manager) handleStartStop() {
	// 检查实际服务器状态
	actualStatus := utils.IsServerRunning()
	m.serverRunning = actualStatus

	if m.serverRunning {
		// 停止服务器
		if confirmed := dialog.Message("Are you sure you want to stop the Oadin server?").Title("Confirm Stop Server").YesNo(); confirmed {
			err := utils.StopOADINServer(filepath.Join(m.pidPath, "oadin.pid"))
			if err == nil {
				m.serverRunning = false
			} else {
				dialog.Message(fmt.Sprintf("Stop server failed: %v", err)).Title("Error").Error()
			}
		}
	} else {
		// 启动服务器
		err := utils.StartOADINServer(m.logPath, m.pidPath)
		if err == nil {
			m.serverRunning = true

			// 启动成功后自动打开浏览器
			go func() {
				time.Sleep(3 * time.Second) // 等待3秒让服务器完全启动
				if err := m.openControlPanel(); err != nil {
					fmt.Printf("Failed to open web console: %v\n", err)
				}
			}()
		} else {
			dialog.Message(fmt.Sprintf("Start server failed: %v", err)).Title("Error").Error()
		}
	}
}

// 处理打开控制台
func (m *Manager) handleOpenConsole() {
	// 检查服务器是否运行
	if !utils.IsServerRunning() {
		// 如果服务器没运行，询问是否启动
		if confirmed := dialog.Message("Oadin server is not running. Start it now?").Title("Start Server").YesNo(); confirmed {
			err := utils.StartOADINServer(m.logPath, m.pidPath)
			if err != nil {
				dialog.Message(fmt.Sprintf("Failed to start server: %v", err)).Title("Error").Error()
				return
			}
			m.serverRunning = true

			// 等待服务器启动
			time.Sleep(3 * time.Second)
		} else {
			return
		}
	}

	// 打开控制台
	err := m.openControlPanel()
	if err != nil {
		dialog.Message(fmt.Sprintf("Failed to open control panel: %v", err)).Title("Error").Error()
	}
}

func (m *Manager) onExit() {
	// Cleanup
}

func (m *Manager) updateStartStopMenuItem(item *systray.MenuItem) {
	if m.serverRunning {
		item.SetTitle("Stop Server")
		item.SetTooltip("Stop Oadin server")
	} else {
		item.SetTitle("Start Server")
		item.SetTooltip("Start Oadin server")
	}
}

func (m *Manager) showStatus() {
	if m.serverRunning {
		fmt.Println("Oadin Server: Running")
	} else {
		fmt.Println("Oadin Server: Stopped")
	}
}

func (m *Manager) openControlPanel() error {
	url := "http://127.0.0.1:16699/"
	err := browser.OpenURL(url)
	if err != nil {
		return fmt.Errorf("failed to open browser: %v", err)
	}
	return nil
}

// SetUpdateAvailable set update label
func (m *Manager) SetUpdateAvailable(available bool) {
	m.updateAvailable = available
	if available {
		m.mRestartUpdate.Show()
	} else {
		m.mRestartUpdate.Hide()
	}
}

// performUpdate
func (m *Manager) performUpdate() error {
	// 1. 停止服务（已在菜单逻辑中处理）
	// 2. 执行更新
	if err := m.installUpdate(); err != nil {
		return fmt.Errorf("failed to install update: %v", err)
	}
	return nil
}

// installUpdate
func (m *Manager) installUpdate() error {
	emptyStatus := utils.IsDirEmpty(config.GlobalOADINEnvironment.UpdateDir)
	if emptyStatus {
		return fmt.Errorf("installation directory is empty")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		installFilaPath := filepath.Join(config.GlobalOADINEnvironment.UpdateDir, "oadin-installer-latest.pkg")
		cmd = exec.Command("open", installFilaPath)
	case "linux":
		return fmt.Errorf("auto update not supported on Linux yet")
	case "windows":
		installFilaPath := filepath.Join(config.GlobalOADINEnvironment.UpdateDir, "oadin-installer-latest.exe")
		cmd = exec.Command(installFilaPath, "/S")
	}
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install update: %v", err)
	}

	return nil
}

// restartApplication 重启应用
func (m *Manager) restartApplication() error {
	// 获取当前可执行文件路径
	dir := filepath.Dir(m.execPath)

	// 创建新进程
	cmd := exec.Command(m.execPath)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	// 启动新进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %v", err)
	}

	return nil
}

func (m *Manager) viewLogs() error {
	err := tray.ShowLogs()
	return err
}

// getIcon returns the icon data
func getIcon() ([]byte, error) {
	// 根据系统类型和主题获取不同的图片
	var file string
	if runtime.GOOS == "darwin" {
		if isMacDarkMode() {
			file = "OADIN-图标-单白.png"
		} else {
			file = "OADIN-图标-单黑.png" 
		}
	} else if runtime.GOOS == "windows" {
		file = "OADIN-图标-彩色.png"
	} else {
		file = "OADIN-图标-彩色.png" 
	}

	data, err := trayTemplate.TrayIconFS.ReadFile(file)
	if err != nil {
		data, err = trayTemplate.TrayIconFS.ReadFile("OADIN-图标-彩色.png")
		if err != nil {
			return nil, err
		}
	}
	return data, nil
}

// 判断 macOS 是否为暗色模式
func isMacDarkMode() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		return false // 未设置暗色模式时会报错
	}
	return string(out) == "Dark\n"
}
