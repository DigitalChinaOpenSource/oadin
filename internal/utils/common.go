package utils

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
)

var textContentTypes = []string{"text/", "application/json", "application/xml", "application/javascript", "application/x-ndjson"}

type MemoryInfo struct {
	Size       int
	MemoryType string
}

func IsHTTPText(header http.Header) bool {
	if contentType := header.Get("Content-Type"); contentType != "" {
		ct := strings.ToLower(contentType)
		for _, t := range textContentTypes {
			if strings.Contains(ct, t) {
				return true
			}
		}
	}
	return false
}

func BodyToString(header http.Header, body []byte) string {
	if IsHTTPText(header) {
		return string(body)
	}
	return fmt.Sprintf("<Binary Data: %d bytes>", len(body))
}

// GetAbsolutePath Convert relative path to absolute path from the passed in base directory
// No change if the passed in path is already an absolute path
func GetAbsolutePath(p string, base string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(base, p))
}

func GetUserDataDir() (string, error) {
	var dir string
	switch sys := runtime.GOOS; sys {
	case "darwin":
		dir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support")
	case "windows":
		dir = filepath.Join(os.Getenv("APPDATA"))
	case "linux":
		dir = filepath.Join(os.Getenv("HOME"), ".config")
	default:
		return "", fmt.Errorf("unsupported operating system")
	}

	return dir, nil
}

func GetAOGDataDir() (string, error) {
	var dir string
	userDir, err := GetUserDataDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(userDir, "AOG")
	if err = os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
	}
	return dir, nil
}

// CheckServiceIsExistInEnv  检查服务是否存在于环境变量
func CheckServiceIsExistInEnv(serviceName string) bool {
	_, err := exec.LookPath(serviceName)
	return err != nil
}

func Contains(slice []string, target string) bool {
	for _, str := range slice {
		if str == target {
			return true
		}
	}
	return false
}

func DownloadFile(downloadURL string, saveDir string) (string, error) {
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	fileName := filepath.Base(parsedURL.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		return "", fmt.Errorf("could not determine file name from URL: %s", downloadURL)
	}

	savePath := filepath.Join(saveDir, fileName)

	if _, err := os.Stat(savePath); err == nil {
		fmt.Printf("%s already exists, skip download.\n", savePath)
		return savePath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check file %s: %v", savePath, err)
	}

	proxyURL, err := http.ProxyFromEnvironment(&http.Request{URL: parsedURL})
	if err != nil {
		return "", fmt.Errorf("failed to get proxy URL: %v", err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	resp, err := client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: HTTP status %s", resp.Status)
	}

	file, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return savePath, nil
}

func Sha256hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func HmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

func HmacSha256String(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	hmacResult := hashed.Sum(nil)
	signature := hex.EncodeToString(hmacResult)
	return signature
}

// generate nonce str
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateNonceString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func GetDownloadDir() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		downloadsPath := os.Getenv("USERPROFILE")
		if downloadsPath == "" {
			return "", fmt.Errorf("unable to get user profile directory on Windows")
		}
		return filepath.Join(downloadsPath, "Downloads"), nil
	case "darwin":
		return filepath.Join(currentUser.HomeDir, "Downloads"), nil
	case "linux":
		xdgDownload := os.Getenv("XDG_DOWNLOAD_DIR")
		if xdgDownload != "" {
			return xdgDownload, nil
		}
		return filepath.Join(currentUser.HomeDir, "Downloads"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func IpexOllamaSupportGPUStatus() bool {
	gpu, err := ghw.GPU()
	if err != nil {
		return false
	}

	for _, card := range gpu.GraphicsCards {
		if strings.Contains(card.DeviceInfo.Product.Name, "Intel") {
			if strings.Contains(card.DeviceInfo.Product.Name, "Arc") || strings.Contains(card.DeviceInfo.Product.Name, "Core") {
				return true
			}
		}
	}
	return false
}

// +-----------------------------+--------------------------------------------------------------------+
// | Device ID                   | 0                                                                  |
// +-----------------------------+--------------------------------------------------------------------+
// | GPU Utilization (%)         | 0                                                                  |
// | EU Array Active (%)         |                                                                    |
// Analyze the output table content of the above terminal command
func ParseTableOutput(output string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				key := strings.TrimSpace(parts[1])
				value := strings.TrimSpace(parts[2])
				if key != "" && value != "" {
					result[key] = value
				}
			}
		}
	}
	return result
}

func GetGpuInfo() (int, error) {
	gpuInfo := "0"
	isIntelEngine := IpexOllamaSupportGPUStatus()
	if isIntelEngine {
		cmd := exec.Command("xpu-smi", "stats", "-d", "0")
		output, err := cmd.Output()
		if err != nil {
			return 0, err
		}
		result := ParseTableOutput(string(output))
		gpuInfo = result["GPU Utilization (%)"]
	} else {
		cmd := exec.Command("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits")
		output, err := cmd.Output()
		if err != nil {
			return 0, err
		}
		gpuInfo = string(output)
	}
	gpuUtilization, err := strconv.Atoi(gpuInfo)
	if err != nil {
		return 0, err
	}
	return gpuUtilization, nil
}

// 解压文件 区分系统 win/linux/macos
func UnzipFile(zipFile, destDir string) error {
	// 检查目标目录是否存在，如果不存在则创建
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		err := os.MkdirAll(destDir, 0o750)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", destDir, err)
		}
	}

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Expand-Archive -Path \"%s\" -DestinationPath \"%s\"", zipFile, destDir))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to unzip file %s: %v", zipFile, err)
		}
		return nil
	case "linux":
		cmd := exec.Command("unzip", zipFile, "-d", destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to unzip file %s: %v", zipFile, err)
		}
		return nil
	case "darwin":
		cmd := exec.Command("unzip", zipFile, "-d", destDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to unzip file %s: %v", zipFile, err)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return nil
}

func IsServerRunning() bool {
	serverUrl := "http://127.0.0.1:16688" + "/health"
	resp, err := http.Get(serverUrl)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func StartAOGServer(logPath string, pidFilePath string) error {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	execCmd := "aog.exe"
	if runtime.GOOS != "windows" {
		execCmd = "aog"
	}
	cmd := exec.Command(execCmd, "server", "start")
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start aog server: %v", err)
	}

	// Save PID to file.
	pid := cmd.Process.Pid
	pidFile := filepath.Join(pidFilePath, "aog.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0o644); err != nil {
		return fmt.Errorf("failed to save PID to file: %v", err)
	}

	fmt.Printf("\raog server started with PID: %d\n", cmd.Process.Pid)
	return nil
}

func StopAOGServer(pidFilePath string) error {
	files, err := filepath.Glob(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to list pid files: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No running processes found")
		return nil
	}

	// Traverse all pid files.
	for _, pidFile := range files {
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Printf("Failed to read PID file %s: %v\n", pidFile, err)
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if err != nil {
			fmt.Printf("Invalid PID in file %s: %v\n", pidFile, err)
			continue
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Failed to find process with PID %d: %v\n", pid, err)
			continue
		}

		if err := process.Kill(); err != nil {
			if strings.Contains(err.Error(), "process already finished") {
				fmt.Printf("Process with PID %d is already stopped\n", pid)
			} else {
				fmt.Printf("Failed to kill process with PID %d: %v\n", pid, err)
				continue
			}
		} else {
			fmt.Printf("Successfully stopped process with PID %d\n", pid)
		}

		// remove pid file
		if err := os.Remove(pidFile); err != nil {
			fmt.Printf("Failed to remove PID file %s: %v\n", pidFile, err)
		}
	}
	return nil
}

func ParseImageData(data []byte) ([][]byte, error) {
	// Create a reader
	r := bytes.NewReader(data)
	var images [][]byte
	fmt.Println(len(data))

	// First, read the number of images (4 bytes)
	var imageCount uint32
	err := binary.Read(r, binary.LittleEndian, &imageCount)
	if err != nil {
		return nil, fmt.Errorf("failed to read image count: %v", err)
	}

	// Read each image
	for i := 0; i < int(imageCount); i++ {
		// Read the image length (4 bytes)
		var imgLen uint32
		err := binary.Read(r, binary.LittleEndian, &imgLen)
		if err != nil {
			return nil, fmt.Errorf("failed to read image length for image %d: %v", i+1, err)
		}

		// Read image data
		imgData := make([]byte, imgLen)
		_, err = io.ReadFull(r, imgData)
		if err != nil {
			return nil, fmt.Errorf("failed to read image data for image %d: %v", i+1, err)
		}

		images = append(images, imgData)
	}

	return images, nil
}
