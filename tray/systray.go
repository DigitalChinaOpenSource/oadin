package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
func NewManager(serverRunningStatus bool, logPath, pidPath string) *Manager {
	execPath, _ := os.Executable()
	return &Manager{
		serverRunning:   serverRunningStatus,
		updateAvailable: false,
		execPath:        execPath,
		logPath:         logPath,
		pidPath:         pidPath,
	}
}

// Start initializes the system tray
func (m *Manager) Start() {
	// 启动时自动启动服务
	if !utils.IsServerRunning() {
		_ = utils.StartOADINServer(m.logPath, m.pidPath)
		m.serverRunning = true
	}
	// 启动 Web Console 前端服务（如有需要）
	startWebConsoleFrontend()
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
	conn, err := exec.Command("powershell", "-Command", "Test-NetConnection -ComputerName 127.0.0.1 -Port "+port+" | Select-Object -ExpandProperty TcpTestSucceeded").Output()
	if err == nil && string(conn) == "True\r\n" {
		return true
	}
	// Linux/Mac
	err2 := exec.Command("sh", "-c", "nc -z 127.0.0.1 "+port).Run()
	return err2 == nil
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
	mStartStop := systray.AddMenuItem("Start/Stop Server", "Start/Stop Oadin server")
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
				if m.serverRunning {
					if confirmed := dialog.Message("Are you sure you want to stop the Oadin server?").Title("Confirm Stop Server").YesNo(); confirmed {
						err := utils.StopOADINServer(filepath.Join(m.pidPath, "oadin.pid"))
						if err == nil {
							m.serverRunning = false
							m.updateStartStopMenuItem(mStartStop)
						} else {
							dialog.Message("Stop server failed.").Title("Error").Error()
						}
					}
				} else {
					err := utils.StartOADINServer(m.logPath, m.pidPath)
					if err == nil {
						m.serverRunning = true
						m.updateStartStopMenuItem(mStartStop)
					} else {
						dialog.Message("Start server failed.").Title("Error").Error()
					}
				}

			case <-mConsole.ClickedCh:
				err := m.openControlPanel()
				if err != nil {
					dialog.Message("Failed to open control panel: %v", err).Title("Error").Error()
				}
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
	// 确保前端服务已启动
	startWebConsoleFrontend()
	url := "http://127.0.0.1:16699/"
	err := browser.OpenURL(url)
	if err != nil {
		return err
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
		installFilaPath := ""
		fmt.Sprintf(installFilaPath)
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
	file := "oadin.ico" // 默认彩色
	if runtime.GOOS == "darwin" {
		file = "oadin-mac-dark.ico"
		if isMacDarkMode() {
			file = "oadin-mac-white.ico"
		}
	}
	data, err := trayTemplate.TrayIconFS.ReadFile(file)
	if err != nil {
		return nil, err
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
