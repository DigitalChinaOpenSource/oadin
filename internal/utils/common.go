package utils

import (
	"bufio"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"gorm.io/gorm/utils"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"oadin/internal/types"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/disk"
)

var textContentTypes = []string{"text/", "application/json", "application/xml", "application/javascript", "application/x-ndjson"}

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

func GetOadinDataDir() (string, error) {
	var dir string
	userDir, err := GetUserDataDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(userDir, "Oadin")
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
	src := rand.NewSource(time.Now().UnixNano())
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

func DetectGpuModel() string {
	gpu, err := ghw.GPU()
	if err != nil {
		return types.GPUTypeNone
	}

	hasNvidia := false
	hasAMD := false
	hasIntel := false

	for _, card := range gpu.GraphicsCards {
		// 转为小写
		productName := strings.ToLower(card.DeviceInfo.Product.Name)
		if strings.Contains(productName, "nvidia") {
			hasNvidia = true
		} else if strings.Contains(productName, "amd") {
			hasAMD = true
		} else if strings.Contains(productName, "intel") && (strings.Contains(productName, "arc") || strings.Contains(productName, "core")) {
			hasIntel = true
		}
	}

	if hasNvidia && hasAMD {
		return types.GPUTypeNvidia + "," + types.GPUTypeAmd
	} else if hasNvidia {
		return types.GPUTypeNvidia
	} else if hasAMD {
		return types.GPUTypeAmd
	} else if hasIntel {
		return types.GPUTypeIntelArc
	} else {
		return types.GPUTypeNone
	}
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
	gpuInfo = strings.TrimSpace(gpuInfo)
	gpuUtilization, err := strconv.Atoi(gpuInfo)
	if err != nil {
		return 0, err
	}
	return gpuUtilization, nil
}

// oadin
type SmartVisionUrlInfo struct {
	Url             string `json:"url"`
	AccessToken     string `json:"access_token"`
	ChatEnterPoint  string `json:"chat_enter_point"`
	EmbedEnterPoint string `json:"embed_enter_point"`
}

func GetSmartVisionUrl() map[string]SmartVisionUrlInfo {
	SmartVisionUrlMap := make(map[string]SmartVisionUrlInfo)
	SmartVisionUrlMap["product"] = SmartVisionUrlInfo{
		Url:             "https://smartvision.dcclouds.com",
		AccessToken:     "private-CA004JS8Xkzblqv1gp8M6iBS",
		ChatEnterPoint:  "/api/v1/aipc/chat/completions",
		EmbedEnterPoint: "/api/v1/aipc/chat/embedding",
	}
	SmartVisionUrlMap["dev"] = SmartVisionUrlInfo{
		Url:             "https://smartvision-dev.digitalchina.com",
		AccessToken:     "private-doBMjUAikf2ErGqVUGzs4yGe",
		ChatEnterPoint:  "/api/v1/aipc/chat/completions",
		EmbedEnterPoint: "/api/v1/aipc/chat/embedding",
	}
	return SmartVisionUrlMap
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

func StartOadinServer(logPath string, pidFilePath string) error {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()
	execCmd := "oadin.exe"
	if runtime.GOOS != "windows" {
		execCmd = "oadin"
	}
	cmd := exec.Command(execCmd, "server", "start")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if runtime.GOOS == "windows" {
		SetCmdSysProcAttr(cmd)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Oadin server: %v", err)
	}
	// Save PID to file.
	pid := cmd.Process.Pid
	pidFile := filepath.Join(pidFilePath, "oadin.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0o644); err != nil {
		return fmt.Errorf("failed to save PID to file: %v", err)
	}

	fmt.Printf("\rOadin server started with PID: %d\n", cmd.Process.Pid)
	return nil
}

func StopOadinServer(pidFilePath string) error {
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
	if runtime.GOOS == "windows" && IpexOllamaSupportGPUStatus() {
		extraProcessName := "ollama-lib.exe"
		extraCmd := exec.Command("taskkill", "/IM", extraProcessName, "/F")
		_, err := extraCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("failed to kill process: %s", extraProcessName)
			return nil
		}

		fmt.Printf("Successfully killed process: %s\n", extraProcessName)
	}
	return nil
}

func SystemDiskSize(path string) (*PathDiskSizeInfo, error) {
	if runtime.GOOS == "windows" {
		path = filepath.VolumeName(path)
	}
	usage, err := disk.Usage(path)
	if err != nil {
		return &PathDiskSizeInfo{}, err
	}
	res := &PathDiskSizeInfo{}
	res.TotalSize = int(usage.Total / 1024 / 1024 / 1024)
	res.FreeSize = int(usage.Free / 1024 / 1024 / 1024)
	res.UsageSize = int(usage.Used / 1024 / 1024 / 1024)
	return res, nil
}

// PathPermissionResult holds the result of a permission check
type PathPermissionResult struct {
	HasPermission bool     // Whether all paths have proper permissions
	FailedPaths   []string // List of paths that failed permission checks
}

// CheckPathPermission Recursively check the read and write permissions of all files and directories under the given path.
// As soon as any file or directory does not meet the required permissions, immediately return false.
func CheckPathPermission(path string) bool {
	// 获取路径信息
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	// 根据路径类型检查权限
	if info.IsDir() {
		return checkDirPermission(path)
	} else {
		return checkFilePermission(path)
	}
}

// checkDirPermission 检查目录的权限
func checkDirPermission(path string) bool {
	// check dictionary read permission
	_, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	// check dictionary write permission
	testFile := filepath.Join(path, ".permission_test_file")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		return false
	}
	_ = os.Remove(testFile)

	// Traverse all files and subdirectories under the directory.
	err = filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory since it has already been checked.
		if currentPath == path {
			return nil
		}

		// Check permissions based on the type of the path.
		if info.IsDir() {
			if !checkDirPermission(currentPath) {
				return filepath.SkipAll
			}
		} else {
			if !checkFilePermission(currentPath) {
				return filepath.SkipAll
			}
		}
		return nil
	})

	return err == nil
}

// checkFilePermission check file permission
func checkFilePermission(path string) bool {
	// read and write permission
	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return false
	}
	file.Close()

	return true
}

func IsDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read just one entry from the directory
	fName, err := f.Readdirnames(3)
	if runtime.GOOS == "darwin" {
		fileHideCount := 0
		for _, name := range fName {
			if utils.Contains([]string{".", "..", ".DS_Store"}, name) {
				fileHideCount += 1
			}
		}
		if err == nil && fileHideCount == len(fName) {
			return true
		}
	}

	// If we got an EOF error, the directory is empty
	return err == io.EOF
}

// CopyDir Copy the source directory to the target location while preserving file attributes and structure.
// src: source path
// dest: target path
func CopyDir(src, dest string) error {
	// get source file info
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("get source file info failed: %v", err)
	}

	// handle symbolic link
	if srcInfo.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return fmt.Errorf("Read a symbolic link %s failed: %v", src, err)
		}
		return os.Symlink(linkTarget, dest)
	}

	if !srcInfo.IsDir() {
		return copyFile(src, dest)
	}

	err = os.MkdirAll(dest, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create target dictionary failed %s 失败: %v", dest, err)
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dest, relPath)

		if path == src {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, targetPath)
		}

		if info.IsDir() {
			err = os.MkdirAll(targetPath, info.Mode())
			if err != nil {
				return err
			}
			// 保持目录时间戳
			return os.Chtimes(targetPath, info.ModTime(), info.ModTime())
		}

		// 处理普通文件
		return copyFile(path, targetPath)
	})
}

// CopyFile Copy the source file to the target location
// src: source file
// dest: target file
func copyFile(src, dest string) error {
	srcInfo, err := os.Lstat(src)
	if err != nil {
		return fmt.Errorf("get source file info failed: %v", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create target file failed: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return err
	}

	return nil
}

func GetFilePathTotalSize(path string) (int64, error) {
	var totalSize int64 = 0

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	return totalSize / 1024 / 1024 / 1024, err
}

func ClearDir(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		err = os.RemoveAll(entryPath)
		if err != nil {
			return err
		}
	}

	return nil
}
