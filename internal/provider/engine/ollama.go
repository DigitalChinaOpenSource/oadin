package engine

import (
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
	"time"

	"intel.com/aog/internal/client"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
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

	AOGDir, err := utils.GetAOGDataDir()
	if err != nil {
		slog.Error("Get AOG data dir failed: ", err.Error())
		logger.EngineLogger.Error("[Ollama] Get AOG data dir failed: " + err.Error())
		return nil
	}

	downloadPath := fmt.Sprintf("%s/%s/%s", AOGDir, "engine", "ollama")
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o750)
		if err != nil {
			logger.EngineLogger.Error("[Ollama] Create download dir failed: " + err.Error())
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

func (o *OllamaProvider) StartEngine(mode string) error {
	logger.EngineLogger.Info("[Ollama] Start engine mode: " + mode)
	execFile := "ollama"
	switch runtime.GOOS {
	case "windows":
		if utils.IpexOllamaSupportGPUStatus() {
			logger.EngineLogger.Info("[Ollama] start ipex-llm-ollama...")
			execFile = o.EngineConfig.ExecPath + "/" + o.EngineConfig.ExecFile
			logger.EngineLogger.Info("[Ollama] exec file path: " + execFile)
		} else {
			execFile = "ollama.exe"
		}
	case "darwin":
		execFile = "/Applications/Ollama.app/Contents/Resources/ollama"
	case "linux":
		execFile = "ollama"
	default:
		logger.EngineLogger.Error("[Ollama] unsupported operating system: " + runtime.GOOS)
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if mode == types.EngineStartModeDaemon {
		cmd := exec.Command(execFile, "serve")
		err := cmd.Start()
		if err != nil {
			logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
			return fmt.Errorf("failed to start ollama: %v", err)
		}

		rootPath, err := utils.GetAOGDataDir()
		if err != nil {
			logger.EngineLogger.Error("[Ollama] failed get aog dir: " + err.Error())
			return fmt.Errorf("failed get aog dir: %v", err)
		}
		pidFile := fmt.Sprintf("%s/ollama.pid", rootPath)
		err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0o644)
		if err != nil {
			logger.EngineLogger.Error("[Ollama] failed to write pid file: " + err.Error())
			return fmt.Errorf("failed to write pid file: %v", err)
		}

		go func() {
			cmd.Wait()
		}()
	} else {
		if utils.IpexOllamaSupportGPUStatus() {
			cmd := exec.Command(o.EngineConfig.ExecPath + "/ollama-serve.bat")
			err := cmd.Start()
			if err != nil {
				logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
				return fmt.Errorf("failed to start ollama: %v", err)
			}
		} else {
			cmd := exec.Command(execFile, "serve")
			err := cmd.Start()
			if err != nil {
				logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
				return fmt.Errorf("failed to start ollama: %v", err)
			}
		}
	}

	return nil
}

func (o *OllamaProvider) StopEngine() error {
	rootPath, err := utils.GetAOGDataDir()
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed get aog dir: " + err.Error())
		return fmt.Errorf("failed get aog dir: %v", err)
	}
	pidFile := fmt.Sprintf("%s/ollama.pid", rootPath)

	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed to read pid file: " + err.Error())
		return fmt.Errorf("failed to read pid file: %v", err)
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		logger.EngineLogger.Error("[Ollama] invalid pid format: " + err.Error())
		return fmt.Errorf("invalid pid format: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed to find process: " + err.Error())
		return fmt.Errorf("failed to find process: %v", err)
	}

	if err := process.Kill(); err != nil {
		logger.EngineLogger.Error("[Ollama] failed to kill process: " + err.Error())
		return fmt.Errorf("failed to kill process: %v", err)
	}

	if err := os.Remove(pidFile); err != nil {
		logger.EngineLogger.Error("[Ollama] failed to remove pid file: " + err.Error())
		return fmt.Errorf("failed to remove pid file: %v", err)
	}

	return nil
}

func (o *OllamaProvider) GetConfig() *types.EngineRecommendConfig {
	if o.EngineConfig != nil {
		return o.EngineConfig
	}

	userDir, err := os.UserHomeDir()
	if err != nil {
		logger.EngineLogger.Error("[Ollama] Get user home dir failed: ", err.Error())
		return nil
	}

	downloadPath, _ := utils.GetDownloadDir()
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o755)
		if err != nil {
			logger.EngineLogger.Error("[Ollama] Create download dir failed: ", err.Error())
			return nil
		}
	}

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := ""
	switch runtime.GOOS {
	case "windows":
		if utils.IpexOllamaSupportGPUStatus() {
			execPath = fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
			execFile = "ollama.exe"
			downloadUrl = "http://120.232.136.73:31619/aogdev/ollama-intel-2.3.0b20250429-win.zip"
			enginePath = fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
		} else {
			execFile = "ollama.exe"
			execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
			downloadUrl = "http://120.232.136.73:31619/aogdev/OllamaSetup.exe"
			enginePath = fmt.Sprintf("%s/%s", userDir, "ollama")
		}
	case "linux":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
		downloadUrl = "http://120.232.136.73:31619/aogdev/OllamaSetup.exe"
	case "darwin":
		execFile = "ollama"
		execPath = fmt.Sprintf("/%s/%s/%s/%s/%s", "Applications", "Ollama.app", "Contents", "Resources", "ollama")
		if runtime.GOARCH == "amd64" {
			downloadUrl = "http://120.232.136.73:31619/aogdev/Ollama-darwin.zip"
		} else {
			downloadUrl = "http://120.232.136.73:31619/aogdev/Ollama-arm64.zip"
		}
	default:
		logger.EngineLogger.Error("[Ollama] unsupported operating system: " + runtime.GOOS)
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           "127.0.0.1:16677",
		Origin:         "127.0.0.1",
		Scheme:         "http",
		RecommendModel: "deepseek-r1:7b",
		DownloadUrl:    downloadUrl,
		DownloadPath:   downloadPath,
		EnginePath:     enginePath,
		ExecPath:       execPath,
		ExecFile:       execFile,
	}
}

func (o *OllamaProvider) HealthCheck() error {
	c := o.GetDefaultClient()
	if err := c.Do(context.Background(), http.MethodHead, "/", nil, nil); err != nil {
		logger.EngineLogger.Error("[Ollama] Health check failed: " + err.Error())
		return err
	}
	logger.EngineLogger.Info("[Ollama] Ollama server health")

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

		//cmd := exec.Command("open", appPath)
		//if err := cmd.Run(); err != nil {
		//	return fmt.Errorf("failed to open ollama installer: %v", err)
		//}
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

		} else { // Handle other operating systems
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			cmd := exec.CommandContext(ctx, file)
			_, err := cmd.CombinedOutput()
			if err != nil {
				// 如果是超时错误
				if ctx.Err() == context.DeadlineExceeded {
					fmt.Println("cmd execute timeout")
					return err
				}
				fmt.Printf("cmd execute error: %v\n", err)
				return err
			}
			return nil
		}
	}
	slog.Info("[Install Engine] model engine install completed")
	return nil
}

func (o *OllamaProvider) InitEnv() error {
	err := os.Setenv("OLLAMA_HOST", o.EngineConfig.Host)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed to set OLLAMA_HOST: " + err.Error())
		return fmt.Errorf("failed to set OLLAMA_HOST: %w", err)
	}

	err = os.Setenv("OLLAMA_ORIGIN", o.EngineConfig.Origin)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed to set OLLAMA_ORIGIN: " + err.Error())
		return fmt.Errorf("failed to set OLLAMA_ORIGIN: %w", err)
	}
	return nil
}

func (o *OllamaProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	logger.EngineLogger.Info("[Ollama] Pull model: " + req.Name)

	c := o.GetDefaultClient()
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray

	var resp types.ProgressResponse
	if err := c.Do(ctx, http.MethodPost, "/api/pull", req, &resp); err != nil {
		logger.EngineLogger.Error("[Ollama] Pull model failed : " + err.Error())
		return &resp, err
	}
	logger.EngineLogger.Info("[Ollama] Pull model success: " + req.Name)

	return &resp, nil
}

func (o *OllamaProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	logger.EngineLogger.Info("[Ollama] Pull model: " + req.Name + " , mode: stream")

	c := o.GetDefaultClient()
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray
	dataCh, errCh := c.StreamResponse(ctx, http.MethodPost, "/api/pull", req)
	logger.EngineLogger.Info("[Ollama] Pull model success: " + req.Name + " , mode: stream")

	return dataCh, errCh
}

func (o *OllamaProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	logger.EngineLogger.Info("[Ollama] Delete model: " + req.Model)

	c := o.GetDefaultClient()
	if err := c.Do(ctx, http.MethodDelete, "/api/delete", req, nil); err != nil {
		logger.EngineLogger.Error("[Ollama] Delete model failed : " + err.Error())
		return err
	}
	logger.EngineLogger.Info("[Ollama] Delete model success: " + req.Model)

	return nil
}

func (o *OllamaProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	c := o.GetDefaultClient()
	var lr types.ListResponse
	if err := c.Do(ctx, http.MethodGet, "/api/tags", nil, &lr); err != nil {
		logger.EngineLogger.Error("[Ollama] Get model list failed :" + err.Error())
		return nil, err
	}

	return &lr, nil
}
