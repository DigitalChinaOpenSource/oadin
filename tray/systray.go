package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	onServerStart   func() error
	onServerStop    func() error
	onAllSeverStop  func() error
	onKillProcess   func() error
	updateAvailable bool
	mRestartUpdate  *systray.MenuItem
	execPath        string
}

// NewManager creates a new system tray manager
func NewManager(onStart, onStop, onAllStop, onKillProcess func() error, serverRunningStatus bool) *Manager {
	execPath, _ := os.Executable()
	return &Manager{
		serverRunning:   serverRunningStatus,
		onServerStart:   onStart,
		onServerStop:    onStop,
		onAllSeverStop:  onAllStop,
		onKillProcess:   onKillProcess,
		updateAvailable: false,
		execPath:        execPath,
	}
}

// Start initializes the system tray
func (m *Manager) Start() {
	systray.Run(m.onReady, m.onExit)
}

func (m *Manager) onReady() {
	// Set icon
	data, err := getIcon()
	if err != nil {
		data = icon.Data
	}
	systray.SetIcon(data)
	// systray.SetTitle("Oadin")
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
				oadinServerStatus := utils.IsServerRunning()
				if oadinServerStatus != m.serverRunning {
					m.serverRunning = oadinServerStatus
					m.updateStartStopMenuItem(mStartStop)
				} else {
					if m.serverRunning {
						// Show confirmation dialog before stopping
						if confirmed := dialog.Message("Are you sure you want to stop the Oadin server?").Title("Confirm Stop Server").YesNo(); confirmed {
							if err := m.onServerStop(); err == nil {
								m.serverRunning = false
								m.updateStartStopMenuItem(mStartStop)
							} else {
								dialog.Message("Failed to stop server: %v", err).Title("Error").Error()
							}
						}
					} else {
						if err := m.onServerStart(); err == nil {
							m.serverRunning = true
							m.updateStartStopMenuItem(mStartStop)
						} else {
							dialog.Message("Failed to start server: %v", err).Title("Error").Error()
						}
					}
				}

			case <-mConsole.ClickedCh:
				err := m.openControlPanel()
				if err != nil {
					dialog.Message("Failed to open control panel: %v", err).Title("Error").Error()
				}
			case <-m.mRestartUpdate.ClickedCh:
				if confirmed := dialog.Message("This will stop all servers and install the update. Continue?").Title("Confirm Update").YesNo(); confirmed {
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
					// Show confirmation dialog before quitting
					if confirmed := dialog.Message("Server is still running. Do you want to stop it and quit?").Title("Confirm Quit").YesNo(); confirmed {
						if err := m.onKillProcess(); err != nil {
							dialog.Message("Failed to stop server: %v", err).Title("Error").Error()
						}
						systray.Quit()
						return
					}
				} else {
					if confirmed := dialog.Message("Are you sure you want to quit Oadin?").Title("Confirm Quit").YesNo(); confirmed {
						if err := m.onKillProcess(); err != nil {
							dialog.Message("Failed to stop server: %v", err).Title("Error").Error()
						}
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
	status := "Stopped"
	if m.serverRunning {
		status = "Running"
	}
	fmt.Printf("Oadin Server Status: %s\nOS: %s\n", status, runtime.GOOS)
}

func (m *Manager) openControlPanel() error {
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
	// 1. 停止所有服务器
	if err := m.onAllSeverStop(); err != nil {
		return fmt.Errorf("failed to stop servers: %v", err)
	}

	// 2. 执行更新
	if err := m.installUpdate(); err != nil {
		return fmt.Errorf("failed to install update: %v", err)
	}
	return nil
}

// installUpdate
func (m *Manager) installUpdate() error {
	emptyStatus := utils.IsDirEmpty(config.GlobalOadinEnvironment.UpdateDir)
	if emptyStatus {
		return fmt.Errorf("installation directory is empty")
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		installFilaPath := filepath.Join(config.GlobalOadinEnvironment.UpdateDir, "oadin-installer-latest.pkg")
		cmd = exec.Command("open", installFilaPath)
	case "linux":
		installFilaPath := ""
		fmt.Sprintf(installFilaPath)
	case "windows":
		installFilaPath := filepath.Join(config.GlobalOadinEnvironment.UpdateDir, "oadin-installer-latest.exe")
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
	// 尝试从多个位置获取图标
	// 根据系统类型获取不同的图片，Windows用彩色的，mac用黑白的
	// oadin.ico oadin-mac.ico  aiwenxue.ico aiwenxue-mac.ico
	file := "aiwenxue.ico"
	if runtime.GOOS == "darwin" {
		file = "aiwenxue-mac-dark.ico"
		isMacDark := isMacDarkMode()
		if isMacDark {
			file = "aiwenxue-mac-white.ico"
		}
	}
	data, err := trayTemplate.TrayIconFS.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// isMacDarkMode 判断 macOS 是否为暗色模式 白色主题用黑的，黑色主题用白的
func isMacDarkMode() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		return false // 未设置暗色模式时会报错
	}
	return strings.Contains(string(out), "Dark")
}
