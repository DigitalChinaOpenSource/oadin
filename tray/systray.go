package tray

import (
	trayTemplate "byze/tray/icon"
	tray "byze/tray/utils"
	"byze/version"
	"context"
	"fmt"
	"github.com/getlantern/systray/example/icon"
	"github.com/pkg/browser"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
	systray.SetTitle("Byze")
	systray.SetTooltip("Byze AI Service Manager")

	// Add menu items
	mStartStop := systray.AddMenuItem("Start Server", "Start/Stop Byze server")
	mStatus := systray.AddMenuItem("Status", "Show server status")
	mStatus.Disabled()
	systray.AddSeparator()
	mConsole := systray.AddMenuItem("Web Console", "Open Control Panel")
	mUpdate := systray.AddMenuItem("Check Update", "Check for updates")
	m.mRestartUpdate = systray.AddMenuItem("Restart and Update", "Restart and install update")
	m.mRestartUpdate.Hide() // 默认隐藏更新按钮
	mViewLogs := systray.AddMenuItem("View Logs", "View Logs")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit Byze")

	// Update initial server status
	m.updateStartStopMenuItem(mStartStop)

	// Handle menu item clicks
	go func() {
		for {
			select {
			case <-mStartStop.ClickedCh:
				if m.serverRunning {
					// Show confirmation dialog before stopping
					if confirmed := dialog.Message("Are you sure you want to stop the Byze server?").Title("Confirm Stop Server").YesNo(); confirmed {
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
			case <-mStatus.ClickedCh:
				m.showStatus()
			case <-mConsole.ClickedCh:
				err := m.openControlPanel()
				if err != nil {
					dialog.Message("Failed to open control panel: %v", err).Title("Error").Error()
				}
			case <-mUpdate.ClickedCh:
				m.checkUpdate()
			case <-m.mRestartUpdate.ClickedCh:
				if confirmed := dialog.Message("This will stop all servers and install the update. Continue?").Title("Confirm Update").YesNo(); confirmed {
					if err := m.performUpdate(); err != nil {
						dialog.Message("Failed to perform update: %v", err).Title("Error").Error()
					} else {
						// 成功执行更新后，退出系统托盘
						systray.Quit()
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
					if confirmed := dialog.Message("Are you sure you want to quit Byze?").Title("Confirm Quit").YesNo(); confirmed {
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
		item.SetTooltip("Stop Byze server")
	} else {
		item.SetTitle("Start Server")
		item.SetTooltip("Start Byze server")
	}
}

func (m *Manager) showStatus() {
	status := "Stopped"
	if m.serverRunning {
		status = "Running"
	}
	fmt.Printf("Byze Server Status: %s\nOS: %s\n", status, runtime.GOOS)
}

func (m *Manager) openControlPanel() error {
	url := "http://127.0.0.1:16699/"
	err := browser.OpenURL(url)
	if err != nil {
		return err
	}
	return nil
}

// SetUpdateAvailable 设置是否有更新可用
func (m *Manager) SetUpdateAvailable(available bool) {
	m.updateAvailable = available
	if available {
		m.mRestartUpdate.Show()
	} else {
		m.mRestartUpdate.Hide()
	}
}

// checkUpdate 检查更新
func (m *Manager) checkUpdate() {
	status, _ := version.IsNewVersionAvailable(context.Background())
	m.updateAvailable = status
	m.SetUpdateAvailable(m.updateAvailable)
}

// performUpdate 执行更新
func (m *Manager) performUpdate() error {
	// 1. 停止所有服务器
	if err := m.onAllSeverStop(); err != nil {
		return fmt.Errorf("failed to stop servers: %v", err)
	}

	// 2. 执行更新
	if err := m.installUpdate(); err != nil {
		return fmt.Errorf("failed to install update: %v", err)
	}

	// 3. 重启应用
	if err := m.restartApplication(); err != nil {
		return fmt.Errorf("failed to restart application: %v", err)
	}

	return nil
}

// installUpdate 安装更新
func (m *Manager) installUpdate() error {
	// TODO: 实现安装更新的逻辑
	// 1. 备份当前版本

	// 2. 安装新版本
	// 3. 更新配置文件等
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
	data, err := trayTemplate.TrayIconFS.ReadFile("byze.ico")
	if err != nil {
		return nil, err
	}
	return data, nil
}
