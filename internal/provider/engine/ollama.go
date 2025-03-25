package engine

import (
	"aipc/byze/internal/types"
	"aipc/byze/internal/utils"
	"aipc/byze/internal/utils/client"
	"aipc/byze/internal/utils/progress"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"

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
		execFile = "ollama"
	case "linux":
		execFile = "ollama"
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	cmd := exec.Command(execFile, "serve")
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

	return nil
}

func (o *OllamaProvider) GetConfig() *types.EngineRecommendConfig {
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

	execFile := ""
	execPath := ""
	downloadUrl := ""
	switch runtime.GOOS {
	case "windows":
		if utils.IpexOllamaSupportGPUStatus() {
			execPath = fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
			execFile = "ollama.exe"
			downloadUrl = "http://120.232.136.73:31619/byzedev/ipex-llm-ollama-Installer-20250122.exe"
		} else {
			execFile = "ollama.exe"
			execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
			downloadUrl = "http://120.232.136.73:31619/byzedev/OllamaSetup.exe"
		}
	case "linux":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
		downloadUrl = "http://120.232.136.73:31619/byzedev/OllamaSetup.exe"
	case "darwin":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
		if runtime.GOARCH == "amd64" {
			downloadUrl = "http://120.232.136.73:31619/byzedev/Ollama-darwin.zip"
		} else {
			downloadUrl = "http://120.232.136.73:31619/byzedev/Ollama-arm64.zip"
		}
	default:
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           "127.0.0.1:16677",
		Origin:         "127.0.0.1",
		Scheme:         "http",
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

func (o *OllamaProvider) InstallEngine() error {
	file, err := utils.DownloadFile(o.EngineConfig.DownloadUrl, o.EngineConfig.DownloadPath)
	if err != nil {
		return fmt.Errorf("failed to download ollama: %v", err)
	}

	slog.Info("[Install Engine] start install...")
	cmd := exec.Command(file)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install ollama: %v", err)
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
	return nil
}

func (o *OllamaProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	c := o.GetDefaultClient()

	var resp types.ProgressResponse
	if err := c.Do(ctx, http.MethodPost, "/api/pull", req, &resp); err != nil {
		slog.Error("Pull model failed : " + err.Error())
		return &resp, err
	}
	return &resp, nil
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
