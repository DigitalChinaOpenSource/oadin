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
	fmt.Println("=== Oadin Tray Starting ===")

	// 调试：列出嵌入的图标文件
	trayTemplate.DebugListFiles()

	// 检查服务器状态
	m.serverRunning = utils.IsServerRunning()
	fmt.Printf("Initial server status: %v\n", m.serverRunning)

	// 启动时如果服务器没运行，自动启动并打开浏览器
	if !m.serverRunning {
		fmt.Println("Server not running, attempting to start...")
		err := utils.StartOADINServer(m.logPath, m.pidPath)
		if err == nil {
			m.serverRunning = true
			fmt.Println("Server started successfully")
			// 启动成功后自动打开浏览器
			go m.waitAndOpenBrowser()
		} else {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	} else {
		fmt.Println("Server is already running")
		// 如果服务器已经运行，直接打开浏览器
		go m.waitAndOpenBrowser()
	}

	fmt.Println("Starting system tray...")
	systray.Run(m.onReady, m.onExit)
}

// waitAndOpenBrowser 等待服务器完全启动后打开浏览器
func (m *Manager) waitAndOpenBrowser() {
	fmt.Println("Waiting for server to be fully ready...")

	// 最多等待30秒，每秒检查一次端口
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		if isPortInUse("16699") {
			fmt.Printf("Server is ready on port 16699 after %d seconds\n", i+1)
			// 再等待1秒确保服务完全就绪
			time.Sleep(1 * time.Second)

			if err := m.openControlPanel(); err != nil {
				fmt.Printf("Failed to open web console: %v\n", err)
			} else {
				fmt.Println("Successfully opened web console")
			}
			return
		}
		fmt.Printf("Waiting for server... (%d/30)\n", i+1)
	}

	fmt.Println("Timeout waiting for server to start")
}

// 启动 Web Console 前端服务
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
	fmt.Println("System tray ready, setting up...")

	// Set icon - 针对 Windows 优先使用 ICO 格式
	var data []byte
	customData, err := getIcon()
	if err != nil {
		fmt.Printf("Failed to load custom icon, using default: %v\n", err)
		data = icon.Data
	} else {
		fmt.Printf("Successfully loaded custom icon, size: %d bytes\n", len(customData))
		data = customData
	}

	systray.SetIcon(data)
	systray.SetTitle("Oadin")
	systray.SetTooltip("Oadin AI Service Manager")
	fmt.Println("Tray icon and menu set up successfully")

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
				dialog.Message("Stop server failed: %v", err).Title("Error").Error()
			}
		}
	} else {
		// 启动服务器
		err := utils.StartOADINServer(m.logPath, m.pidPath)
		if err == nil {
			m.serverRunning = true
			// 启动成功后打开浏览器
			go m.waitAndOpenBrowser()
		} else {
			dialog.Message("Start server failed: %v", err).Title("Error").Error()
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
				dialog.Message("Failed to start server: %v", err).Title("Error").Error()
				return
			}
			m.serverRunning = true
			// 启动后等待并打开浏览器
			go m.waitAndOpenBrowser()
		} else {
			return
		}
	} else {
		// 服务器已运行，直接打开浏览器
		go m.waitAndOpenBrowser()
	}
}

func (m *Manager) onExit() {
	fmt.Println("=== Oadin Tray Exiting ===")
	// 确保完全退出
	os.Exit(0)
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
	fmt.Printf("Attempting to open URL: %s\n", url)

	err := browser.OpenURL(url)
	if err != nil {
		fmt.Printf("browser.OpenURL failed: %v\n", err)
		return fmt.Errorf("failed to open browser: %v", err)
	}

	fmt.Println("Browser.OpenURL called successfully")
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
	var files []string

	if runtime.GOOS == "darwin" {
		if isMacDarkMode() {
			files = []string{"oadin-icon-white.ico", "oadin-icon-white.png"}
		} else {
			files = []string{"OADIN-图标-单黑.ico", "OADIN-图标-单黑.png"}
		}
	} else if runtime.GOOS == "windows" {
		files = []string{"oadin-icon.ico", "oadin-icon.png"}
	} else {
		files = []string{"oadin-icon.ico", "oadin-icon.png"}
	}

	for _, file := range files {
		fmt.Printf("Trying to load icon: %s\n", file)
		data, err := trayTemplate.TrayIconFS.ReadFile(file)
		if err == nil {
			fmt.Printf("Successfully loaded icon: %s, size: %d bytes\n", file, len(data))
			return data, nil
		}
		fmt.Printf("Failed to load %s: %v\n", file, err)
	}

	fmt.Println("Trying fallback icons...")
	fallbackFiles := []string{
		"oadin-icon.ico", "OADIN-图标-单黑.ico", "oadin-icon-white.ico",
		"oadin-icon.png", "OADIN-图标-单黑.png", "oadin-icon-white.png",
	}

	for _, fallback := range fallbackFiles {
		data, err := trayTemplate.TrayIconFS.ReadFile(fallback)
		if err == nil {
			fmt.Printf("Successfully loaded fallback icon: %s, size: %d bytes\n", fallback, len(data))
			return data, nil
		}
		fmt.Printf("Failed to load fallback %s: %v\n", fallback, err)
	}

	return nil, fmt.Errorf("no icon files found")
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
