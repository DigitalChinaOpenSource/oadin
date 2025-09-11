//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

package utils

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm/utils"

	"oadin/internal/types"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/disk"
)

const (
	// Content types
	ContentTypeJSON   = "application/json"
	ContentTypeXML    = "application/xml"
	ContentTypeJS     = "application/javascript"
	ContentTypeNDJSON = "application/x-ndjson"
	ContentTypeText   = "text/"

	// Time conversion constants
	SecondsPerMinute      = 60
	SecondsPerHour        = 3600
	MillisecondsPerSecond = 1000

	// Default timestamp format
	DefaultSRTTime = "00:00:00,000"
)

var textContentTypes = []string{ContentTypeText, ContentTypeJSON, ContentTypeXML, ContentTypeJS, ContentTypeNDJSON}

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

func GetOADINDataDir() (string, error) {
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

/*
DownloadFileWithProgress 下载文件并支持进度回调
path, err := DownloadFileWithProgress(url, saveDir, func(downloaded, total int64) {
    fmt.Printf("下载进度: %.2f%%\n", float64(downloaded)*100/float64(total))
})
*/
func DownloadFileWithProgress(downloadURL string, saveDir string, onProgress func(downloaded, total int64)) (string, error) {
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
        return "", fmt.Errorf("failed to download file: %v, url: %s", err, downloadURL)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to download file: HTTP status %s, url: %s", resp.Status, downloadURL)
    }

    file, err := os.Create(savePath)
    if err != nil {
        return "", fmt.Errorf("failed to create file: %v, url: %s", err, downloadURL)
    }
    defer file.Close()

    var total int64 = 0
    if resp.ContentLength > 0 {
        total = resp.ContentLength
    }

    var downloaded int64 = 0
    var lastProgressTime time.Time
    var lastProgressBytes int64
    
    // 如果文件小于10MB 进度更新间隔：100ms 或者每1MB
	progressInterval := 100 * time.Millisecond
	progressThreshold := 1024 * 1024
	buf := make([]byte, 16*1024)

	if total > 100*1024*1024 {
		progressInterval = 500 * time.Millisecond
		progressThreshold = 5 * 1024 * 1024
		buf = make([]byte, 64*1024) // 增大缓冲区到64KB
	}

	if total > 1000*1024*1024 {
		progressInterval = 1000 * time.Millisecond
		progressThreshold = 10 * 1024 * 1024
		buf = make([]byte, 256*1024)
	}

    for {
        nr, er := resp.Body.Read(buf)
        if nr > 0 {
            nw, ew := file.Write(buf[0:nr])
            if nw > 0 {
                downloaded += int64(nw)
                
                // 优化进度回调：只在时间间隔或字节阈值满足时调用
                now := time.Now()
                bytesFromLastProgress := downloaded - lastProgressBytes
                
                if onProgress != nil && total > 0 && 
                   (now.Sub(lastProgressTime) >= progressInterval || 
                    bytesFromLastProgress >= int64(progressThreshold) || 
                    downloaded == total) {
                    onProgress(downloaded, total)
                    lastProgressTime = now
                    lastProgressBytes = downloaded
                }
            }
            if ew != nil {
                return "", fmt.Errorf("failed to write file: %v", ew)
            }
            if nr != nw {
                return "", fmt.Errorf("short write")
            }
        }
        if er != nil {
            if er == io.EOF {
                // 确保最终进度为100%
                if onProgress != nil && total > 0 {
                    onProgress(total, total)
                }
                break
            }
            return "", fmt.Errorf("failed to read response body: %v", er)
        }
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

func HmacSha1String(s, key string) string {
	hashed := hmac.New(sha1.New, []byte(key))
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

func ParseRequestBody(reqBody []byte) (map[string]interface{}, error) {
	var body map[string]interface{}
	if err := json.Unmarshal(reqBody, &body); err != nil {
		return nil, fmt.Errorf("unmarshal request body: %w", err)
	}
	return body, nil
}

func BuildGetRequestURL(baseURL string, body []byte) (string, error) {
	queryParams := make(map[string][]string)
	if err := json.Unmarshal(body, &queryParams); err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, values := range queryParams {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// GenerateUUID 生成一个唯一的UUID字符串
func GenerateUUID() string {
	// 使用时间和随机字符串组合生成唯一ID
	now := time.Now().UnixNano()
	random := GenerateNonceString(16)
	return fmt.Sprintf("%d-%s", now, random)
}

// DecodeBase64 解码Base64字符串为字节数组
func DecodeBase64(data string) ([]byte, error) {
	// 处理可能的URL安全Base64格式
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	// 处理不完整的Base64字符串
	missing := len(data) % 4
	if missing > 0 {
		data += strings.Repeat("=", 4-missing)
	}

	return base64.StdEncoding.DecodeString(data)
}

func ParseSRTTimestamps(srtContent string) (*int, *int) {
	var beginTime, endTime *int

	// 检查内容是否为空
	if srtContent == "" {
		return nil, nil
	}

	// 按行分割内容
	lines := strings.Split(srtContent, "\n")

	// 查找时间戳行 (格式: 00:00:00,000 --> 00:00:00,000)
	for _, line := range lines {
		if strings.Contains(line, " --> ") {
			parts := strings.Split(line, " --> ")
			if len(parts) == 2 {
				// 解析开始时间
				start := ParseTimestamp(parts[0])
				if start >= 0 {
					startMs := start
					if beginTime == nil || startMs < *beginTime {
						beginTime = &startMs
					}
				}

				// 解析结束时间
				end := ParseTimestamp(parts[1])
				if end >= 0 {
					endMs := end
					if endTime == nil || endMs > *endTime {
						endTime = &endMs
					}
				}

				// 由于我们只需要找到最早的开始时间和最晚的结束时间，可以继续搜索下一个时间戳行
			}
		}
	}

	return beginTime, endTime
}

// ParseTimestamp 将 SRT 格式的时间戳 (00:00:00,000) 转换为毫秒
func ParseTimestamp(timestamp string) int {
	// 去除可能的空白字符
	timestamp = strings.TrimSpace(timestamp)

	// 分离毫秒部分
	parts := strings.Split(timestamp, ",")
	if len(parts) != 2 {
		return -1
	}

	timeStr := parts[0] // 00:00:00
	msStr := parts[1]   // 000

	// 解析时间部分 (小时:分钟:秒)
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) != 3 {
		return -1
	}

	hours, errH := strconv.Atoi(timeParts[0])
	minutes, errM := strconv.Atoi(timeParts[1])
	seconds, errS := strconv.Atoi(timeParts[2])
	milliseconds, errMs := strconv.Atoi(msStr)

	if errH != nil || errM != nil || errS != nil || errMs != nil {
		return -1
	}

	// 转换为总毫秒数
	totalMs := hours*3600000 + minutes*60000 + seconds*1000 + milliseconds

	return totalMs
}

// Min 辅助函数，返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NowUnixMilli 获取当前Unix时间戳（毫秒）
func NowUnixMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func ReadImageFileToBase64(filePath string) (string, error) {
	imgData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("read image file failed: %w", err)
	}
	return base64.StdEncoding.EncodeToString(imgData), nil
}

func DownloadImageUrlToPath(url string) (string, error) {
	downLoadPath, err := GetDownloadDir()
	if err != nil {
		return "", fmt.Errorf("get download dir failed: %w", err)
	}
	savePath, err := DownloadFile(url, downLoadPath)
	if err != nil {
		return "", fmt.Errorf("download image file failed: %w", err)
	}
	return savePath, nil
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

func SystemDiskSize(path string) (*types.PathDiskSizeInfo, error) {
	if runtime.GOOS == "windows" {
		path = filepath.VolumeName(path)
	}
	usage, err := disk.Usage(path)
	if err != nil {
		return &types.PathDiskSizeInfo{}, err
	}
	res := &types.PathDiskSizeInfo{}
	res.TotalSize = int(usage.Total / 1024 / 1024 / 1024)
	res.FreeSize = int(usage.Free / 1024 / 1024 / 1024)
	res.UsageSize = int(usage.Used / 1024 / 1024 / 1024)
	return res, nil
}

func ParseSizeToGB(sizeStr string) float64 {
	s := strings.TrimSpace(strings.ToUpper(sizeStr))
	re := regexp.MustCompile(`([\d.]+)\s*(GB|MB)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		return 1
	}
	num, _ := strconv.ParseFloat(matches[1], 64)
	unit := matches[2]
	if unit == "GB" {
		return num
	} else if unit == "MB" {
		return num / 1024
	}
	return 1
}

// FormatSecondsToSRT 将秒数转换为SRT时间格式 (HH:MM:SS,mmm)
func FormatSecondsToSRT(secondsStr string) string {
	seconds, err := strconv.ParseFloat(secondsStr, 64)
	if err != nil {
		return DefaultSRTTime
	}

	hours := int(seconds) / SecondsPerHour
	minutes := (int(seconds) % SecondsPerHour) / SecondsPerMinute
	secs := int(seconds) % SecondsPerMinute
	milliseconds := int((seconds - float64(int(seconds))) * MillisecondsPerSecond)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, milliseconds)
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
	err = os.WriteFile(testFile, []byte("test"), 0o644)
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
