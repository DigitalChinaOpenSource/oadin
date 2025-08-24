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

package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"oadin/internal/utils/directory"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"oadin/internal/client"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/types"
	"oadin/internal/utils"
)

const (
	// Default configuration
	DefaultPort = "16677"
	DefaultHost = constants.DefaultHost + ":" + DefaultPort

	// ipex-llm-ollama related
	IpexLlamaDir    = "ipex-llm-ollama"
	OllamaBatchFile = "ollama-serve.bat"

	// Windows download URLs
	WindowsAllGPUURL        = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-all.zip"
	WindowsNvidiaURL        = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64.zip"
	WindowsAMDURL           = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-rocm.zip"
	WindowsIntelArcURL      = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ipex-llm-ollama.zip"
	WindowsBaseURL          = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-base.zip"
	WindowsDDLDependsX64URL = "https://aka.ms/vs/17/release/vc_redist.x64.exe"
	WindowsDDLDependsX86URL = "https://aka.ms/vs/17/release/vc_redist.x86.exe"

	// Linux download URLs
	LinuxURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/OllamaSetup.exe"

	// macOS download URLs
	MacOSIntelURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/Ollama-darwin.zip"

	// Archive commands
	TarCommand     = "tar"
	TarExtractFlag = "-xf"
	TarDestFlag    = "-C"
	UnzipCommand   = "unzip"
	UnzipDestFlag  = "-d"
	MoveCommand    = "mv"
)

var OllamaDDLDependsList = []string{
	"vcruntime140.dll",
	"vcruntime140_1.dll",
	"msvcp140.dll",
	"msvcp140_1.dll",
	"msvcp140_2.dll",
	"concrt140.dll",
}

type OllamaProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func NewOllamaProvider(config *types.EngineRecommendConfig) *OllamaProvider {
	if config != nil {
		return &OllamaProvider{
			EngineConfig: config,
		}
	}

	OADINDir, err := utils.GetOADINDataDir()
	if err != nil {
		slog.Error("Get OADIN data dir failed: ", err.Error())
		logger.EngineLogger.Error("[Ollama] Get OADIN data dir failed: " + err.Error())
		return nil
	}

	downloadPath := fmt.Sprintf("%s/%s/%s", OADINDir, "engine", "ollama")
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
	host := DefaultHost
	if o.EngineConfig.Host != "" {
		host = o.EngineConfig.Host
	}

	// default scheme
	scheme := types.ProtocolHTTP
	if o.EngineConfig.Scheme == types.ProtocolHTTPS {
		scheme = types.ProtocolHTTPS
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
		logger.EngineLogger.Info("[Ollama] start ipex-llm-ollama...")
		execFile = o.EngineConfig.ExecPath + "/" + o.EngineConfig.ExecFile
		logger.EngineLogger.Info("[Ollama] exec file path: " + execFile)
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
		if runtime.GOOS == "windows" {
			utils.SetCmdSysProcAttr(cmd)
		}
		err := cmd.Start()
		if err != nil {
			logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
			return fmt.Errorf("failed to start ollama: %v", err)
		}

		rootPath, err := utils.GetOADINDataDir()
		if err != nil {
			logger.EngineLogger.Error("[Ollama] failed get oadin dir: " + err.Error())
			return fmt.Errorf("failed get oadin dir: %v", err)
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
			cmd := exec.Command(o.EngineConfig.ExecPath + "/" + OllamaBatchFile)
			if runtime.GOOS == "windows" {
				utils.SetCmdSysProcAttr(cmd)
			}
			err := cmd.Start()
			if err != nil {
				logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
				return fmt.Errorf("failed to start ollama: %v", err)
			}
			logger.EngineLogger.Info("ipex运行环境变量: %s", cmd.Env)
		} else {
			cmd := exec.Command(execFile, "serve")
			if runtime.GOOS == "windows" {
				utils.SetCmdSysProcAttr(cmd)
			}
			err := cmd.Start()
			if err != nil {
				logger.EngineLogger.Error("[Ollama] failed to start ollama: " + err.Error())
				return fmt.Errorf("failed to start ollama: %v", err)
			}
		}
	}

	return nil
}

func (o *OllamaProvider) StopEngine(ctx context.Context) error {
	rootPath, err := utils.GetOADINDataDir()
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed get oadin dir: " + err.Error())
		return fmt.Errorf("failed get oadin dir: %v", err)
	}
	pidFile := fmt.Sprintf("%s/ollama.pid", rootPath)
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		logger.EngineLogger.Info("[Ollama] Stop openvino Model Server not found pidfile: " + pidFile)
		return nil
	}
	// unload model
	runningModels, err := o.GetRunningModels(ctx)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed get running models: " + err.Error())
		return fmt.Errorf("failed get running models:: %v", err)
	}
	runningModelList := []string{}
	for _, model := range runningModels.Models {
		runningModelList = append(runningModelList, model.Name)
	}
	err = o.UnloadModel(ctx, &types.UnloadModelRequest{Models: runningModelList})
	if err != nil {
		logger.EngineLogger.Error("[Ollama] failed unload model: " + err.Error())
		return fmt.Errorf("failed unload model: %v", err)
	}

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

	// 此处改造成使用program files目录
	executableDir, err := directory.GetWindowsPaths()

	if err != nil {
		logger.EngineLogger.Error("[Ollama] Get user home dir failed: ", err.Error())
		return nil
	}
	homeDir, err := os.UserHomeDir()

	downloadPath, _ := utils.GetDownloadDir()
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		err := os.MkdirAll(downloadPath, 0o755)
		if err != nil {
			logger.EngineLogger.Error("[Ollama] Create download dir failed: ", err.Error())
			return nil
		}
	}
	// 模型文件的路径
	dataDir := executableDir.ProgramData + "/Oadin"

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := fmt.Sprintf("%s/%s", dataDir, "engine/ollama")
	switch runtime.GOOS {
	case "windows":
		execFile = "ollama.exe"
		execPath = fmt.Sprintf("%s/%s", executableDir.ProgramFiles, "/Oadin/ollama")

		switch utils.DetectGpuModel() {
		case types.GPUTypeNvidia + "," + types.GPUTypeAmd:
			downloadUrl = WindowsAllGPUURL
		case types.GPUTypeNvidia:
			downloadUrl = WindowsNvidiaURL
		case types.GPUTypeAmd:
			downloadUrl = WindowsAMDURL
		case types.GPUTypeIntelArc:
			execPath = fmt.Sprintf("%s/%s", executableDir.ProgramFiles, "/Oadin/ipex-llm-ollama")
			downloadUrl = WindowsIntelArcURL
		default:
			downloadUrl = WindowsBaseURL
		}

	case "linux":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", homeDir, "ollama")
		downloadUrl = LinuxURL
	case "darwin":
		execFile = "ollama"
		execPath = fmt.Sprintf("/%s/%s/%s/%s", "Applications", "Ollama.app", "Contents", "Resources")
		downloadUrl = MacOSIntelURL
	default:
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           DefaultHost,
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

	logger.EngineLogger.Info("[Install Engine] start install...")
	if runtime.GOOS == "darwin" {
		files, err := os.ReadDir(o.EngineConfig.DownloadPath)
		if err != nil {
			logger.EngineLogger.Error("[Install Engine] read dir failed: ", o.EngineConfig.DownloadPath)
		}
		for _, f := range files {
			if f.IsDir() && f.Name() == "__MACOSX" {
				fPath := filepath.Join(o.EngineConfig.DownloadPath, f.Name())
				os.RemoveAll(fPath)
			}
		}
		appPath := filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		if _, err = os.Stat(appPath); os.IsNotExist(err) {
			unzipCmd := exec.Command(UnzipCommand, file, UnzipDestFlag, o.EngineConfig.DownloadPath)
			if err := unzipCmd.Run(); err != nil {
				return fmt.Errorf("failed to unzip file: %v", err)
			}
			appPath = filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		}

		// move it to Applications
		applicationPath := filepath.Join("/Applications/", "Ollama.app")
		if _, err = os.Stat(applicationPath); os.IsNotExist(err) {
			mvCmd := exec.Command(MoveCommand, appPath, "/Applications/")
			if err := mvCmd.Run(); err != nil {
				return fmt.Errorf("failed to move ollama to Applications: %v", err)
			}
		}

	} else if runtime.GOOS == "windows" {
		if utils.IpexOllamaSupportGPUStatus() {
			// 解压文件
			if err != nil {
				logger.EngineLogger.Error("Get user home dir failed: ", err.Error())
				return err
			}
			ipexPath := filepath.Join(o.GetConfig().ExecPath, "ipex-llm-ollama")
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				if runtime.GOOS == "windows" {
					unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
					if err := unzipCmd.Run(); err != nil {
						logger.EngineLogger.Error("[Install Engine] model engine install completed err : " + err.Error())
						return fmt.Errorf("failed to unzip file: %v", err)
					}
				}
			}

		} else {
			filePath := o.EngineConfig.ExecPath
			if _, err = os.Stat(filePath); os.IsNotExist(err) {
				os.MkdirAll(filePath, 0o755)
				unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, filePath)
				if err := unzipCmd.Run(); err != nil {
					logger.EngineLogger.Info("[Install Engine] model engine install completed err : " + err.Error())
					return fmt.Errorf("failed to unzip file: %v", err)
				}
			}
		}
		err = o.InstallEngineExtraDepends(context.Background())
		if err != nil {
			return fmt.Errorf("[Install Engine DDL Depends] completed")
		}
	} else {
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	logger.LogicLogger.Info("[Install Engine] model engine install completed")
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
	currModelPath := os.Getenv("OADIN_OLLAMA_MODELS")
	if currModelPath != "" {
		modelPath = currModelPath
	}
	slog.Info("[Init Env] start model..." + modelPath)
	err = os.Setenv("OLLAMA_MODELS", modelPath)
	if err != nil {
		return fmt.Errorf("failed to set OLLAMA_MODELS: %w", err)
	}
	ollamaModels := os.Getenv("OLLAMA_MODELS")
	slog.Info("[Init Env] end model...ollama" + ollamaModels)

	if utils.IpexOllamaSupportGPUStatus() {
		err = os.Setenv("OLLAMA_NUM_GPU", "999")
		if err != nil {
			return fmt.Errorf("failed to set OLLAMA_NUM_GPU: %w", err)
		}
		err = os.Setenv("ZES_ENABLE_SYSMAN", "1")
		if err != nil {
			return fmt.Errorf("failed to set ZES_ENABLE_SYSMAN: %w", err)
		}
		err = os.Setenv("SYCL_CACHE_PERSISTENT", "1")
		if err != nil {
			return fmt.Errorf("failed to set SYCL_CACHE_PERSISTENT: %w", err)
		}

		// 加长ipex的 长度设置
		err = os.Setenv("OLLAMA_NUM_CTX", "8192")
		if err != nil {
			return fmt.Errorf("failed to set OLLAMA_NUM_CTX: %w", err)
		}
		err = os.Setenv("OLLAMA_NUM_PARALLEL", "1")
		if err != nil {
			return fmt.Errorf("failed to set OLLAMA_NUM_PARALLEL: %w", err)
		}
	}

	return nil
}

func (o *OllamaProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	logger.EngineLogger.Info("[Ollama] Pull model: " + req.Name)

	o.ListModels(ctx)

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
	reqHeader := make(map[string]string)
	reqHeader["Content-Type"] = "application/json"
	reqHeader["Accept"] = "application/json"
	dataCh, errCh := c.StreamResponse(ctx, http.MethodPost, "/api/pull", req, reqHeader)

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

// CopyModel 复制模型
func (o *OllamaProvider) CopyModel(ctx context.Context, req *types.CopyModelRequest) error {
	go func() {
		fmt.Println("Will copy model after 3 seconds: " + req.Source + " to " + req.Destination)
		time.Sleep(1 * time.Second) // 延迟3秒执行
		ctx := context.Background()
		c := o.GetDefaultClient() // 假设你已声明o
		if err := c.Do(ctx, http.MethodPost, "/api/copy", req, nil); err != nil {
			fmt.Println("[Service] copy model failed : " + err.Error())
			slog.Error("copy model failed : " + err.Error())
			// todo: 后续处理
			return
		}
		fmt.Println("Model copy complete: " + req.Destination)
	}()
	return nil
}

func (o *OllamaProvider) GetRunningModels(ctx context.Context) (*types.ListResponse, error) {
	c := o.GetDefaultClient()
	var lr types.ListResponse
	if err := c.Do(ctx, http.MethodGet, "/api/ps", nil, &lr); err != nil {
		slog.Error("[Service] Get run model list failed :" + err.Error())
		return nil, err
	}
	return &lr, nil
}

func (o *OllamaProvider) UnloadModel(ctx context.Context, req *types.UnloadModelRequest) error {
	c := o.GetDefaultClient()
	// chat model
	for _, model := range req.Models {
		reqBody := &types.OllamaUnloadModelRequest{
			Model:     model,
			KeepAlive: 0, // 默认为0，表示不保留模型
		}
		if err := c.Do(ctx, http.MethodPost, "/api/generate", reqBody, nil); err != nil {
			slog.Error("Unload model failed : " + err.Error())
			return err
		}
		slog.Info("Ollama: Model %s unloaded successfully\n", model)
	}
	time.Sleep(2 * time.Second)
	reCheckRunModels, err := o.GetRunningModels(ctx)
	if err != nil {
		logger.EngineLogger.Error("[Ollama] Get run models failed : " + err.Error())
		return err
	}
	// embed model
	for _, model := range reCheckRunModels.Models {
		reqBody := &types.OllamaUnloadModelRequest{
			Model:     model.Name,
			KeepAlive: 0,
		}
		if err := c.Do(ctx, http.MethodPost, "/api/embed", reqBody, nil); err != nil {
			slog.Error("Unload model failed : " + err.Error())
			return err
		}
		slog.Info("Ollama: Model %s unloaded successfully\n", model)
	}
	return nil
}

func (o *OllamaProvider) LoadModel(ctx context.Context, req *types.LoadRequest) error {
	// c := o.GetDefaultClient()
	// lr := &types.OllamaLoadModelRequest{
	// 	Model: req.Model,
	// }
	// if err := c.Do(ctx, http.MethodPost, "/api/generate", lr, nil); err != nil {
	// 	logger.EngineLogger.Error("[Ollama] Load model failed: " + req.Model + " , error: " + err.Error())
	// 	return err
	// }
	return nil
}

var OllamaOperateStatus = 1

func (o *OllamaProvider) GetOperateStatus() int {
	return OllamaOperateStatus
}

func (o *OllamaProvider) SetOperateStatus(status int) {
	OllamaOperateStatus = status
	slog.Info("Ollama operate status set to", "status", OllamaOperateStatus)
}

func (o *OllamaProvider) InstallEngineStream(ctx context.Context, newDataChan chan []byte, newErrChan chan error) {
	defer close(newDataChan)
	defer close(newErrChan)

	// 下载
	onProgress := func(downloaded, total int64) {
		if total > 0 {
			progress := types.ProgressResponse{
				Status:    "downloading",
				Total:     total,
				Completed: downloaded,
			}
			if dataBytes, err := json.Marshal(progress); err == nil {
				newDataChan <- dataBytes
			}
		}
	}
	file, err := utils.DownloadFileWithProgress(o.EngineConfig.DownloadUrl, o.EngineConfig.DownloadPath, onProgress)
	if err != nil {
		newErrChan <- err
		return
	}
	logger.EngineLogger.Info("[Install Engine] start install...")
	if runtime.GOOS == "darwin" {
		files, err := os.ReadDir(o.EngineConfig.DownloadPath)
		if err != nil {
			logger.EngineLogger.Error("[Install Engine] read dir failed: ", o.EngineConfig.DownloadPath)
			newErrChan <- err
			return
		}
		for _, f := range files {
			if f.IsDir() && f.Name() == "__MACOSX" {
				fPath := filepath.Join(o.EngineConfig.DownloadPath, f.Name())
				os.RemoveAll(fPath)
			}
		}
		appPath := filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		if _, err = os.Stat(appPath); os.IsNotExist(err) {
			unzipCmd := exec.Command(UnzipCommand, file, UnzipDestFlag, o.EngineConfig.DownloadPath)
			if err := unzipCmd.Run(); err != nil {
				newErrChan <- err
				return
			}
			appPath = filepath.Join(o.EngineConfig.DownloadPath, "Ollama.app")
		}

		// move it to Applications
		applicationPath := filepath.Join("/Applications/", "Ollama.app")
		if _, err = os.Stat(applicationPath); os.IsNotExist(err) {
			mvCmd := exec.Command(MoveCommand, appPath, "/Applications/")
			if err := mvCmd.Run(); err != nil {
				newErrChan <- err
				return
			}
		}

	} else {
		if utils.IpexOllamaSupportGPUStatus() {
			// 解压文件
			//userDir, err := os.UserHomeDir()
			// 目录改造
			if err != nil {
				newErrChan <- err
				return
			}
			ipexPath := o.GetConfig().ExecPath
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				if runtime.GOOS == "windows" {
					unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
					if err := unzipCmd.Run(); err != nil {
						logger.EngineLogger.Error("[Install Engine] model engine install completed err : ", err.Error())
						newErrChan <- err
						return
					}
				}
			} else {
				execPath := filepath.Join(o.GetConfig().ExecPath, o.GetConfig().ExecFile)
				if _, err = os.Stat(execPath); os.IsNotExist(err) {
					unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
					if err := unzipCmd.Run(); err != nil {
						logger.EngineLogger.Error("[Install Engine] model engine install completed err 2: ", err.Error())
						newErrChan <- err
						return
					}
				}
			}

		} else if runtime.GOOS == "windows" {
			ipexPath := o.EngineConfig.ExecPath
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
				if err := unzipCmd.Run(); err != nil {
					logger.LogicLogger.Info("[Install Engine] model engine install completed err : ", err.Error())
					newErrChan <- err
					return
				}
			}
		} else {
			err := fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
			newErrChan <- err
			return
		}
	}

	logger.EngineLogger.Info("[ollama] model engine install completed")
}

func (o *OllamaProvider) InstallEngineExtraDepends(ctx context.Context) error {
	// check ddl file exists
	missingList := []string{}
	for _, f := range OllamaDDLDependsList {
		checkStatus := utils.CheckDllExists(f)
		if !checkStatus {
			missingList = append(missingList, f)
		}
	}
	if len(missingList) > 0 {
		// download ddl depends
		var downloadUrl string
		if runtime.GOARCH == "amd64" {
			downloadUrl = WindowsDDLDependsX64URL
		} else {
			downloadUrl = WindowsDDLDependsX86URL
		}
		file, err := utils.DownloadFile(downloadUrl, o.EngineConfig.ExecPath)
		if err != nil {
			logger.LogicLogger.Error("[Install Engine DDL Depends] download url failed: ", downloadUrl)
			return err
		}
		logger.EngineLogger.Info("[Install Engine DDL Depends] install Successfully")
		cmd := exec.Command(file, "/quiet", "/norestart")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return nil
}
