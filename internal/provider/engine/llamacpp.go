package engine

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"oadin/internal/client"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/models/modelscope"
	"oadin/internal/types"
	"oadin/internal/utils"
)

const (
	// Default configuration for llamacpp
	llamacppDefaultPort = "16697"
	llamacppDefaultHost = constants.DefaultHost + ":" + llamacppDefaultPort

	llamacppServerExec  = "llama-swap.exe"
	LlamaVulkanPath     = "llamacpp-windows-vulkan"
	LlamaSwapPath       = "llama-swap_137_windows_amd64"
	LlamaSwapConfigFile = "config.yaml"
	// Windows download URLs for llamacpp
	llamacppWindowsBaseURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/llamacpp-windows-vulkan.zip"
)

type llamacppProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func NewLlamacppProvider(config *types.EngineRecommendConfig) *llamacppProvider {
	if config != nil {
		return &llamacppProvider{
			EngineConfig: config,
		}
	}
	OADINDir, err := utils.GetOADINDataDir()
	if err != nil {
		slog.Error("Get OADIN data dir failed", "error", err.Error())
		logger.EngineLogger.Error("[llamacpp] Get OADIN data dir failed: " + err.Error())
		return nil
	}

	downloadPath := fmt.Sprintf("%s/%s/%s", OADINDir, "engine", "llamacpp")
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o750)
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] Create download dir failed: " + err.Error())
			return nil
		}
	}

	llamacppProvider := new(llamacppProvider)
	llamacppProvider.EngineConfig = llamacppProvider.GetConfig()

	return llamacppProvider
}

func (l *llamacppProvider) GetDefaultClient() *client.Client {
	// default host
	host := llamacppDefaultHost
	if l.EngineConfig.Host != "" {
		host = l.EngineConfig.Host
	}

	// default scheme
	scheme := types.ProtocolHTTP
	if l.EngineConfig.Scheme == types.ProtocolHTTPS {
		scheme = types.ProtocolHTTPS
	}
	logger.EngineLogger.Info("[llamacpp] Create client", "scheme", scheme, "host", host)
	return client.NewClient(&url.URL{
		Scheme: scheme,
		Host:   host,
	}, http.DefaultClient)
}

func (l *llamacppProvider) StartEngine(mode string) error {
	logger.EngineLogger.Info("[llamacpp] Start engine mode: " + mode)
	execFile := llamacppServerExec
	switch runtime.GOOS {
	case "windows":
		logger.EngineLogger.Info("[llamacpp] start llamacpp-server...")
		execFile = l.EngineConfig.ExecPath + "/" + l.EngineConfig.ExecFile
		logger.EngineLogger.Info("[llamacpp] exec file path: " + execFile)

	case "darwin":
		logger.EngineLogger.Warn("[llamacpp] Darwin system does not support llamacpp yet")

	case "linux":
		logger.EngineLogger.Warn("[llamacpp] linux system does not support llamacpp yet")

	default:
		logger.EngineLogger.Error("[llamacpp] unsupported operating system: " + runtime.GOOS)
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if mode == types.EngineStartModeDaemon {
		LlamaSwapConfigFilePath := fmt.Sprintf("%s/%s", l.EngineConfig.ExecPath, LlamaSwapConfigFile)
		cmd := exec.Command(execFile, "-listen", l.EngineConfig.Host, "-config", LlamaSwapConfigFilePath)
		err := cmd.Start()
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] failed to start llamacpp-server: " + err.Error())
			return fmt.Errorf("failed to start llamacpp-server: %v", err)
		}

		rootPath, err := utils.GetOADINDataDir()
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] failed get oadin dir: " + err.Error())
			return fmt.Errorf("failed get oadin dir: %v", err)
		}
		pidFile := fmt.Sprintf("%s/llamacpp.pid", rootPath)
		err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0o644)
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] failed to write pid file: " + err.Error())
			return fmt.Errorf("failed to write pid file: %v", err)
		}

		go func() {
			cmd.Wait()
		}()
	}

	return nil
}

func (l *llamacppProvider) StopEngine(ctx context.Context) error {
	rootPath, err := utils.GetOADINDataDir()
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] failed get oadin dir: " + err.Error())
		return fmt.Errorf("failed get oadin dir: %v", err)
	}
	pidFile := fmt.Sprintf("%s/llamacpp.pid", rootPath)

	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to read pid file: " + err.Error())
		return fmt.Errorf("failed to read pid file: %v", err)
	}

	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] invalid pid format: " + err.Error())
		return fmt.Errorf("invalid pid format: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to find process: " + err.Error())
		return fmt.Errorf("failed to find process: %v", err)
	}

	if err := process.Kill(); err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to kill process: " + err.Error())
		return fmt.Errorf("failed to kill process: %v", err)
	}

	if err := os.Remove(pidFile); err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to remove pid file: " + err.Error())
		return fmt.Errorf("failed to remove pid file: %v", err)
	}

	return nil
}

func (l *llamacppProvider) GetConfig() *types.EngineRecommendConfig {
	if l.EngineConfig != nil {
		return l.EngineConfig
	}

	downloadPath, _ := utils.GetDownloadDir()
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o755)
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] Create download dir failed: " + err.Error())
			return nil
		}
	}
	dataDir, err := utils.GetOADINDataDir()
	if err != nil {
		slog.Error("Get OADIN data dir failed", "error", err)
		return nil
	}

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := fmt.Sprintf("%s/%s", dataDir, "engine/llamacpp")
	switch runtime.GOOS {
	case "windows":
		execFile = llamacppServerExec
		execPath = fmt.Sprintf("%s/%s/%s", enginePath, LlamaVulkanPath, LlamaSwapPath)

		switch utils.DetectGpuModel() {
		case types.GPUTypeAmd:
			downloadUrl = llamacppWindowsBaseURL
		default:
			downloadUrl = llamacppWindowsBaseURL
		}

	case "linux":
		slog.Warn("[llamacpp] linux system does not support llamacpp yet")

	case "darwin":
		slog.Warn("[llamacpp] Darwin system does not support llamacpp yet")

	default:
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           llamacppDefaultHost,
		Origin:         constants.DefaultHost,
		Scheme:         types.ProtocolHTTP,
		EnginePath:     enginePath,
		RecommendModel: constants.RecommendModel,
		DownloadUrl:    downloadUrl,
		DownloadPath:   downloadPath,
		ExecFile:       execFile,
		ExecPath:       execPath,
	}
}

func (l *llamacppProvider) HealthCheck() error {
	c := l.GetDefaultClient()
	if err := c.Do(context.Background(), http.MethodGet, "/running", nil, nil); err != nil {
		logger.EngineLogger.Error("[llamacpp] Health check failed: ", err.Error(), "client", c)
		return err
	}

	logger.EngineLogger.Info("[llamacpp] llamacpp server health")
	return nil
}

func (l *llamacppProvider) GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error) {
	c := l.GetDefaultClient()
	if err := c.Do(ctx, http.MethodGet, "/version", nil, resp); err != nil {
		slog.Error("Get engine version : " + err.Error())
		return nil, err
	}
	return resp, nil
}

func (l *llamacppProvider) InstallEngine() error {
	file, err := utils.DownloadFile(l.EngineConfig.DownloadUrl, l.EngineConfig.DownloadPath)
	if err != nil {
		return fmt.Errorf("failed to download llamacpp-server: %v, url: %v", err, l.EngineConfig.DownloadUrl)
	}

	logger.EngineLogger.Info("[Install Engine] start install llamacpp...")

	switch runtime.GOOS {
	case "windows":
		// On Windows: Extract to the engine directory
		engineDir := l.EngineConfig.EnginePath
		if _, err := os.Stat(engineDir); os.IsNotExist(err) {
			err := os.MkdirAll(engineDir, 0o755)
			if err != nil {
				return fmt.Errorf("failed to create engine directory: %v", err)
			}
		}

		// Use PowerShell's Expand-Archive command to extract the ZIP file
		unzipCmd := exec.Command("powershell", "-Command", fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", file, engineDir))
		if err := unzipCmd.Run(); err != nil {
			return fmt.Errorf("failed to unzip file to engine directory: %v", err)
		}
		logger.EngineLogger.Info("[Install Engine] llamacpp extracted to engine directory", "path", engineDir)

	case "darwin":
		logger.EngineLogger.Warn("[Install Engine] darwin installation not implemented yet")

	case "linux":
		logger.EngineLogger.Warn("[Install Engine] Linux installation not implemented yet")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	logger.EngineLogger.Info("[Install Engine] llamacpp model engine install completed")
	return nil
}

// todo: Need to add logic here to initialize the llama-swap YAML file
func (l *llamacppProvider) InitEnv() error {
	err := os.Setenv("llamacpp_HOST", l.EngineConfig.Host)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to set llamacpp_HOST: " + err.Error())
		return fmt.Errorf("failed to set llamacpp_HOST: %w", err)
	}

	err = os.Setenv("llamacpp_ORIGIN", l.EngineConfig.Origin)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] failed to set llamacpp_ORIGIN: " + err.Error())
		return fmt.Errorf("failed to set llamacpp_ORIGIN: %w", err)
	}

	return nil
}

func (l *llamacppProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	logger.EngineLogger.Info("[llamacpp] Pull model: " + req.Name)

	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray

	localModelPath := fmt.Sprintf("%s/models/%s", l.EngineConfig.EnginePath, req.Model)

	// Create models directory if it doesn't exist
	modelsDir := fmt.Sprintf("%s/models", l.EngineConfig.EnginePath)
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(modelsDir, 0o750); err != nil {
			logger.EngineLogger.Error("[llamacpp] Failed to create models directory: " + err.Error())
			return nil, err
		}
	}

	// Use ModelScope downloader
	return modelscope.PullModel(ctx, req, localModelPath, constants.GGUFModelType, fn, nil)
}

func (l *llamacppProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	logger.EngineLogger.Info("[llamacpp] Pull model: " + req.Name + " , mode: stream")

	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray

	localModelPath := fmt.Sprintf("%s/models/%s", l.EngineConfig.EnginePath, req.Model)

	// Create models directory if it doesn't exist
	modelsDir := fmt.Sprintf("%s/models", l.EngineConfig.EnginePath)
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(modelsDir, 0o750); err != nil {
			logger.EngineLogger.Error("[llamacpp] Failed to create models directory: " + err.Error())
			dataCh := make(chan []byte)
			errCh := make(chan error, 1)
			close(dataCh)
			errCh <- err
			close(errCh)
			return dataCh, errCh
		}
	}

	// Use ModelScope downloader
	return modelscope.PullModelStream(ctx, req, localModelPath, constants.GGUFModelType, nil)
}

// DeleteModel removes a model from local storage
func (l *llamacppProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	logger.EngineLogger.Info("[llamacpp] Delete model: " + req.Model)

	c := l.GetDefaultClient()
	if err := c.Do(ctx, http.MethodDelete, "/models/"+req.Model, nil, nil); err != nil {
		logger.EngineLogger.Error("[llamacpp] Delete model failed : " + err.Error())
		return err
	}
	logger.EngineLogger.Info("[llamacpp] Delete model success: " + req.Model)

	return nil
}

func (l *llamacppProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	c := l.GetDefaultClient()
	var lr types.ListResponse
	if err := c.Do(ctx, http.MethodGet, "/v1/models", nil, &lr); err != nil {
		logger.EngineLogger.Error("[llamacpp] Get model list failed :" + err.Error())
		return nil, err
	}

	return &lr, nil
}


func (l *llamacppProvider) CopyModel(ctx context.Context, req *types.CopyModelRequest) error {
	return nil
}

func (l *llamacppProvider) LoadModel(ctx context.Context, req *types.LoadRequest) error {
	return nil
}

func (l *llamacppProvider) UnloadModel(ctx context.Context, req *types.UnloadModelRequest) error {
	return nil
}

func (l *llamacppProvider) GetRunningModels(ctx context.Context) (*types.ListResponse, error) {
	return nil, nil
}

func (l *llamacppProvider) GetOperateStatus() int {
	return 0
}

func (l *llamacppProvider) SetOperateStatus(status int) {

}