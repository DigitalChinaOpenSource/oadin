package installer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// DownloadOptions 配置下载行为的选项
type DownloadOptions struct {
	// 超时时间，默认30秒
	Timeout time.Duration
	// 最大重试次数，默认3次
	MaxRetries int
	// 重试间隔，默认2秒
	RetryInterval time.Duration
	// 进度回调函数
	ProgressCallback func(bytesRead, totalBytes int64)
	// 验证下载文件的期望SHA256哈希值（可选）
	ExpectedSHA256 string
}

// DefaultDownloadOptions 返回默认下载选项
func DefaultDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryInterval: 2 * time.Second,
	}
}

// DownloadWithRedirects 从URL下载文件到指定路径，支持重定向处理
func DownloadWithRedirects(url string, destinationPath string, opts *DownloadOptions) error {
	if opts == nil {
		opts = DefaultDownloadOptions()
	}

	// 创建目标目录
	destDir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("创建目录失败 %s: %w", destDir, err)
	}

	var lastErr error
	// 实现重试逻辑
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(opts.RetryInterval)
			fmt.Printf("重试下载 %s (尝试 %d/%d)\n", url, attempt, opts.MaxRetries)
		}

		err := downloadFile(url, destinationPath, opts)
		if err == nil {
			// 如果需要验证文件
			if opts.ExpectedSHA256 != "" {
				if err := verifyFileSHA256(destinationPath, opts.ExpectedSHA256); err != nil {
					lastErr = err
					// 文件校验失败，删除损坏的文件
					_ = os.Remove(destinationPath)
					continue
				}
			}
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("下载失败，已达到最大重试次数: %w", lastErr)
}

// downloadFile 执行实际的文件下载
func downloadFile(url, destinationPath string, opts *DownloadOptions) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置用户代理
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 创建HTTP客户端
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	// 创建临时文件
	tempFile := destinationPath + ".download"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	// 确保在函数返回时关闭并删除临时文件(如果需要)
	defer func() {
		file.Close()
		if err != nil {
			os.Remove(tempFile)
		}
	}()

	// 准备读取响应体
	var reader io.Reader = resp.Body
	totalBytes := resp.ContentLength

	// 如果提供了进度回调，则使用进度读取器
	if opts.ProgressCallback != nil && totalBytes > 0 {
		reader = &progressReader{
			reader:           resp.Body,
			totalBytes:       totalBytes,
			bytesRead:        0,
			progressCallback: opts.ProgressCallback,
		}
	}

	// 写入文件
	if _, err = io.Copy(file, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 关闭文件以确保所有数据已写入
	if err = file.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}

	// 重命名临时文件为最终文件
	if err = os.Rename(tempFile, destinationPath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// 进度读取器结构体
type progressReader struct {
	reader           io.Reader
	totalBytes       int64
	bytesRead        int64
	progressCallback func(bytesRead, totalBytes int64)
}

// Read 实现io.Reader接口
func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	if n > 0 {
		pr.bytesRead += int64(n)
		pr.progressCallback(pr.bytesRead, pr.totalBytes)
	}
	return
}

// verifyFileSHA256 验证文件的SHA256哈希值
func verifyFileSHA256(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法打开文件进行验证: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("计算文件哈希值失败: %w", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("文件哈希值不匹配: 期望 %s, 实际 %s", expectedHash, actualHash)
	}

	return nil
}
