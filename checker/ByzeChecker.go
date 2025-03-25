package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var (
	userResponse   chan bool
	webServerMutex sync.Mutex
	webServerPort  = 5000
)

//export ByzeInit
func ByzeInit(byzeFilePath *C.char) {
	var configPath string
	if byzeFilePath != nil {
		configPath = C.GoString(byzeFilePath)
	} else {
		configPath = filepath.Join(getProjectRootDirectory(), ".byze")
	}

	if !isByzeAvailable() {
		userResponse = make(chan bool)
		startWebServer()

		if !openBrowser(fmt.Sprintf("http://localhost:%d/install-prompt", webServerPort)) {
			return
		}

		select {
		case choice := <-userResponse:
			if !choice {
				return
			}
			if !downloadByze() {
				return
			}
			if !installByze() {
				return
			}
			if !importByze(configPath) {
				return
			}
		case <-time.After(5 * time.Minute):
			return
		}
	}
}

func main() {} // 必须保留空的main函数

// ========== 核心逻辑实现 ==========

func isByzeAvailable() bool {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://localhost:16688")
	return err == nil && resp.StatusCode == http.StatusOK
}

func startWebServer() {
	webServerMutex.Lock()
	defer webServerMutex.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/install-prompt", handleInstallPrompt)
	mux.HandleFunc("/user-response", handleUserResponse)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", webServerPort))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		log.Fatal(http.Serve(listener, mux))
	}()
}

func handleInstallPrompt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, installPromptHTML)
}

func handleUserResponse(w http.ResponseWriter, r *http.Request) {
	choice := r.URL.Query().Get("choice") == "true"
	userResponse <- choice
	w.WriteHeader(http.StatusOK)
}

// ========== 系统交互部分 ==========

func openBrowser(url string) bool {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start() == nil
}

func downloadByze() bool {
	url := "http://120.232.136.73:31619/byzedev/byze.exe"
	userDir, _ := os.UserHomeDir()
	dest := filepath.Join(userDir, "Byze", "byze.exe")

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return false
	}

	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return false
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err == nil
}

func installByze() bool {
	userDir, _ := os.UserHomeDir()
	byzePath := filepath.Join(userDir, "Byze", "byze.exe")

	cmd := exec.Command(byzePath, "server", "start", "-d")
	return cmd.Start() == nil
}

func importByze(configPath string) bool {
	userDir, _ := os.UserHomeDir()
	byzePath := filepath.Join(userDir, "Byze", "byze.exe")

	if !fileExists(configPath) {
		log.Fatalf("配置文件未找到: %s", configPath)
	}

	cmd := exec.Command(byzePath, "import", "--file", configPath)
	return cmd.Run() == nil
}

// ========== 工具方法 ==========

func getProjectRootDirectory() string {
	return filepath.Dir(os.Args[0])
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

// ========== HTML模板 ==========

const installPromptHTML = `
<html><body style="padding:20px;font-family:Arial">
    <h2>安装确认</h2>
    <p>需要安装Byze组件才能继续，是否允许？</p>
    <button onclick="respond(true)">同意安装</button>
    <button onclick="respond(false)">取消</button>
    <script>
        function respond(choice) {
            fetch('/user-response?choice=' + choice)
                .then(() => window.close())
        }
    </script>
</body></html>`
