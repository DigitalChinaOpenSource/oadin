package engine

import (
	"byze/internal/cache"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"byze/internal/types"
	"byze/internal/utils"
	"byze/internal/utils/client"
	"byze/internal/utils/progress"

	"github.com/spf13/cobra"
)

type OllamaProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func NewOllamaProvider(config *types.EngineRecommendConfig) *OllamaProvider {
	if config != nil {
		return &OllamaProvider{
			EngineConfig: config,
		}
	}

	ByzeDir, err := utils.GetByzeDataDir()
	if err != nil {
		slog.Error("Get Byze data dir failed: ", err.Error())
		return nil
	}

	downloadPath := fmt.Sprintf("%s/%s", ByzeDir, "download")
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o750)
		if err != nil {
			return nil
		}
	}

	ollamaProvider := new(OllamaProvider)
	ollamaProvider.EngineConfig = ollamaProvider.GetConfig()

	return ollamaProvider
}

func (o *OllamaProvider) GetDefaultClient() *client.Client {
	// default host
	host := "127.0.0.1:16677"
	if o.EngineConfig.Host != "" {
		host = o.EngineConfig.Host
	}

	// default scheme
	scheme := "http"
	if o.EngineConfig.Scheme == "https" {
		scheme = "https"
	}

	return client.NewClient(&url.URL{
		Scheme: scheme,
		Host:   host,
	}, http.DefaultClient)
}

func (o *OllamaProvider) StartEngine() error {
	execFile := "ollama"
	switch runtime.GOOS {
	case "windows":
		if utils.IpexOllamaSupportGPUStatus() {
			slog.Info("start ipex-llm-ollama...")
			execFile = o.EngineConfig.ExecPath + "/" + o.EngineConfig.ExecFile
			slog.Info("exec file path: " + execFile)
		} else {
			execFile = "ollama.exe"
		}
	case "darwin":
		execFile = "/Applications/Ollama.app/Contents/Resources/ollama"
	case "linux":
		execFile = "ollama"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	cmd := exec.Command(execFile, "serve")
	// 通過啟動命令設置代理
	proxyHttp, proxyHttps := o.StartEngineWithProxy()
	if proxyHttp != "" {
		cmd.Env = append(os.Environ(),
			proxyHttp,
			proxyHttps,
		)
	}

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ollama: %v", err)
	}

	rootPath, err := utils.GetByzeDataDir()
	if err != nil {
		return fmt.Errorf("failed get byze dir: %v", err)
	}
	pidFile := fmt.Sprintf("%s/ollama.pid", rootPath)
	err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0o644)
	if err != nil {
		return fmt.Errorf("failed to write pid file: %v", err)
	}

	go func() {
		cmd.Wait()
	}()

	return nil
}

func (o *OllamaProvider) StopEngine() error {
	rootPath, err := utils.GetByzeDataDir()
	if err != nil {
		return fmt.Errorf("failed get byze dir: %v", err)
	}
	pidFile := fmt.Sprintf("%s/ollama.pid", rootPath)

	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %v", err)
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		return fmt.Errorf("invalid pid format: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}

	if err := os.Remove(pidFile); err != nil {
		return fmt.Errorf("failed to remove pid file: %v", err)
	}
	if runtime.GOOS == "windows" && utils.IpexOllamaSupportGPUStatus() {
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

func (o *OllamaProvider) GetConfig() *types.EngineRecommendConfig {
	if o.EngineConfig != nil {
		return o.EngineConfig
	}
	userDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Get user home dir failed: ", err.Error())
		return nil
	}

	downloadPath, err := utils.GetDownloadDir()
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o755)
		if err != nil {
			return nil
		}
	}

	dataDir, err := utils.GetByzeDataDir()
	if err != nil {
		slog.Error("Get Byze data dir failed: " + err.Error())
		return nil
	}

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := fmt.Sprintf("%s/%s", dataDir, "engine/ollama")
	switch runtime.GOOS {
	case "windows":
		execFile = "ollama.exe"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")

		switch utils.DetectGpuModel() {
		case types.GPUTypeNvidia + "," + types.GPUTypeAmd:
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/ollama-windows-amd64-all.zip"
		case types.GPUTypeNvidia:
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/ollama-windows-amd64.zip"
		case types.GPUTypeAmd:
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/ollama-windows-amd64-rocm.zip"
		case types.GPUTypeIntelArc:
			execPath = fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/ipex-llm-ollama.zip"
		default:
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/ollama-windows-amd64-base.zip"
		}

	case "linux":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
		downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/linux/OllamaSetup.exe"
	case "darwin":
		execFile = "ollama"
		execPath = fmt.Sprintf("/%s/%s/%s/%s/%s", "Applications", "Ollama.app", "Contents", "Resources", "ollama")
		if runtime.GOARCH == "amd64" {
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/macos/Ollama-darwin.zip"
		} else {
			downloadUrl = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/macos/Ollama-arm64.zip"
		}
	default:
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           "127.0.0.1:16677",
		Origin:         "127.0.0.1",
		Scheme:         "http",
		EnginePath:     enginePath,
		RecommendModel: "deepseek-r1:7b",
		DownloadUrl:    downloadUrl,
		DownloadPath:   downloadPath,
		ExecPath:       execPath,
		ExecFile:       execFile,
	}
}

func (o *OllamaProvider) HealthCheck() error {
	c := o.GetDefaultClient()
	if err := c.Do(context.Background(), http.MethodHead, "/", nil, nil); err != nil {
		return err
	}
	return nil
}

func (o *OllamaProvider) GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error) {
	c := o.GetDefaultClient()
	if err := c.Do(ctx, http.MethodGet, "/api/version", nil, resp); err != nil {
		slog.Error("Get engine version : " + err.Error())
		return nil, err
	}
	return resp, nil
}

func (o *OllamaProvider) InstallEngine() error {
	file, err := utils.DownloadFile(o.EngineConfig.DownloadUrl, o.EngineConfig.DownloadPath)
	if err != nil {
		return fmt.Errorf("failed to download ollama: %v, url: %v", err, o.EngineConfig.DownloadUrl)
	}

	slog.Info("[Install Engine] start install...")
	if runtime.GOOS == "darwin" {
		files, err := os.ReadDir(o.EngineConfig.DownloadPath)
		if err != nil {
			slog.Error("[Install Engine] read dir failed: ", o.EngineConfig.DownloadPath)
		}
		for _, f := range files {
			if f.IsDir() && f.Name() == "__MACOSX" {
				fPath := filepath.Join(o.EngineConfig.DownloadPath, f.Name())
				os.RemoveAll(fPath)
			}
		}
		appPath := filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		if _, err = os.Stat(appPath); os.IsNotExist(err) {
			unzipCmd := exec.Command("unzip", file, "-d", o.EngineConfig.DownloadPath)
			if err := unzipCmd.Run(); err != nil {
				return fmt.Errorf("failed to unzip file: %v", err)
			}
			appPath = filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		}

		// move it to Applications
		applicationPath := filepath.Join("/Applications/", "Ollama.app")
		if _, err = os.Stat(applicationPath); os.IsNotExist(err) {
			mvCmd := exec.Command("mv", appPath, "/Applications/")
			if err := mvCmd.Run(); err != nil {
				return fmt.Errorf("failed to move ollama to Applications: %v", err)
			}
		}

	} else {
		if utils.IpexOllamaSupportGPUStatus() {
			// 解压文件
			userDir, err := os.UserHomeDir()
			if err != nil {
				slog.Error("Get user home dir failed: ", err.Error())
				return err
			}
			ipexPath := fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				if runtime.GOOS == "windows" {
					unzipCmd := exec.Command("tar", "-xf", file, "-C", ipexPath)
					if err := unzipCmd.Run(); err != nil {
						return fmt.Errorf("failed to unzip file: %v", err)
					}
				}
			}
		} else if runtime.GOOS == "windows" {
			ipexPath := o.EngineConfig.ExecPath
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				unzipCmd := exec.Command("tar", "-xf", file, "-C", ipexPath)
				if err := unzipCmd.Run(); err != nil {
					return fmt.Errorf("failed to unzip file: %v", err)
				}
			}
		} else {
			return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
		}
	}

	slog.Info("[Install Engine] model engine install completed")
	return nil
}

func (o *OllamaProvider) InitEnv() error {
	err := os.Setenv("OLLAMA_HOST", o.EngineConfig.Host)
	if err != nil {
		return fmt.Errorf("failed to set OLLAMA_HOST: %w", err)
	}

	err = os.Setenv("OLLAMA_ORIGIN", o.EngineConfig.Origin)
	if err != nil {
		return fmt.Errorf("failed to set OLLAMA_ORIGIN: %w", err)
	}

	modelPath := fmt.Sprintf("%s/%s", o.EngineConfig.EnginePath, "models")
	currModelPath := os.Getenv("BYZE_OLLAMA_MODELS")
	if currModelPath != "" {
		modelPath = currModelPath
	}
	err = os.Setenv("OLLAMA_MODELS", modelPath)
	if err != nil {
		return fmt.Errorf("failed to set OLLAMA_MODELS: %w", err)
	}

	return nil
}

func (o *OllamaProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	c := o.GetDefaultClient()

	var resp types.ProgressResponse
	// Abortable HTTP request
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[strings.ToLower(req.Model)], cancel)
	client.ModelClientMap[strings.ToLower(req.Model)] = modelArray

	// 如果用戶設置了Registry, 則傳遞 insecure為true, 然後從私服地址拉取模型
	privateRegistryHandle(req)

	if err := c.Do(ctx, http.MethodPost, "/api/pull", req, &resp); err != nil {
		slog.Error("Pull model failed : " + err.Error())
		return &resp, err
	}
	return &resp, nil
}

func (o *OllamaProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	c := o.GetDefaultClient()
	// Abortable HTTP request
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[strings.ToLower(req.Model)], cancel)
	client.ModelClientMap[strings.ToLower(req.Model)] = modelArray

	// 如果用戶設置了Registry, 則傳遞 insecure為true, 然後從私服地址拉取模型
	privateRegistryHandle(req)

	dataCh, errCh := c.StreamResponse(ctx, http.MethodPost, "/api/pull", req)
	return dataCh, errCh
}

func (o *OllamaProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	fmt.Printf("Ollama: Deleting model %s\n", req.Model)
	c := o.GetDefaultClient()

	if err := c.Do(ctx, http.MethodDelete, "/api/delete", req, nil); err != nil {
		slog.Error("Delete model failed : " + err.Error())
		return err
	}

	return nil
}

func (o *OllamaProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	c := o.GetDefaultClient()
	var lr types.ListResponse
	if err := c.Do(ctx, http.MethodGet, "/api/tags", nil, &lr); err != nil {
		slog.Error("[Service] Get model list failed :" + err.Error())
		return nil, err
	}
	return &lr, nil
}

func (o *OllamaProvider) PullHandler(cmd *cobra.Command, args []string) error {
	insecure, err := cmd.Flags().GetBool("insecure")
	if err != nil {
		return err
	}

	p := progress.NewProgress(os.Stderr)
	defer p.Stop()

	bars := make(map[string]*progress.Bar)

	var status string
	var spinner *progress.Spinner

	fn := func(resp types.ProgressResponse) error {
		if resp.Digest != "" {
			if spinner != nil {
				spinner.Stop()
			}

			bar, ok := bars[resp.Digest]
			if !ok {
				bar = progress.NewBar(fmt.Sprintf("pulling %s...", resp.Digest[7:19]), resp.Total, resp.Completed)
				bars[resp.Digest] = bar
				p.Add(resp.Digest, bar)
			}

			bar.Set(resp.Completed)
		} else if status != resp.Status {
			if spinner != nil {
				spinner.Stop()
			}

			status = resp.Status
			spinner = progress.NewSpinner(status)
			p.Add(status, spinner)
		}

		return nil
	}

	request := types.PullModelRequest{Name: args[0], Insecure: insecure}
	if _, err := o.PullModel(context.Background(), &request, fn); err != nil {
		return err
	}

	return nil
}

// CopyModel 复制模型
func (o *OllamaProvider) CopyModel(ctx context.Context, req *types.CopyModelRequest) error {
	fmt.Println("Ollama CopyModel: " + req.Source + " to " + req.Destination)
	c := o.GetDefaultClient()

	if err := c.Do(ctx, http.MethodDelete, "/api/copy", req, nil); err != nil {
		fmt.Println("[Service] copy model failed : " + err.Error())
		slog.Error("copye model failed : " + err.Error())
		// todo: 貌似复制不成功,后续处理
		return nil
	}

	return nil
}

// 替換為私倉拉取模型, 為防止出現中斷, 不做異常處理
func privateRegistryHandle(req *types.PullModelRequest) {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("获取Ollama仓库地址失败", "error", err)
		return
	}

	// 如果用户设置了Ollama仓库地址，则将其添加到请求中
	if settings.OllamaRegistry != "" {
		req.Insecure = true // 设置为true以允许不安全的连接
		req.Model = settings.OllamaRegistry + "/library/" + req.Model
		fmt.Println("[PullModel] Using private registry:", req.Model)

	}
}

// StartEngineWithProxy 通過代理網絡啟動ollama服務
func (o *OllamaProvider) StartEngineWithProxy() (string, string) {
	var proxyHttp, proxyHttps string
	// 读取用户配置文件中的系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("获取系统设置失败", "error", err)
		return proxyHttp, proxyHttps
	}

	// 通過啟動環境變量設置代理
	if settings.SystemProxy.Enabled {
		if settings.SystemProxy.Endpoint == "" {
			slog.Error("Ollama engine start with proxy, but proxy endpoint is empty")
			return proxyHttp, proxyHttps

		}
		if settings.SystemProxy.Username != "" && settings.SystemProxy.Password != "" {
			proxyHttp = fmt.Sprintf("HTTP_PROXY=http://%s:%s@%s", settings.SystemProxy.Username, settings.SystemProxy.Password, settings.SystemProxy.Endpoint)
			proxyHttps = fmt.Sprintf("HTTPS_PROXY=http://%s:%s@%s", settings.SystemProxy.Username, settings.SystemProxy.Password, settings.SystemProxy.Endpoint)
		} else {
			proxyHttp = fmt.Sprintf("HTTP_PROXY=http://%s", settings.SystemProxy.Endpoint)
			proxyHttps = fmt.Sprintf("HTTPS_PROXY=http://%s", settings.SystemProxy.Endpoint)
		}
		fmt.Println("Ollama engine start with proxy: " + settings.SystemProxy.Endpoint)
		slog.Info("Ollama engine start with proxy: " + settings.SystemProxy.Endpoint)
	} else {
		slog.Info("Ollama engine start without proxy")
	}

	return proxyHttp, proxyHttps
}
