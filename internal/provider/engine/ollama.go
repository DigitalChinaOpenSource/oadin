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
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"intel.com/aog/internal/client"
	"intel.com/aog/internal/constants"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

const (
	// Default configuration
	DefaultPort = "16677"
	DefaultHost = constants.DefaultHost + ":" + DefaultPort

	// ipex-llm-ollama related
	IpexLlamaDir    = "ipex-llm-ollama"
	OllamaBatchFile = "ollama-serve.bat"

	// Windows download URLs
	WindowsAllGPUURL   = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-all.zip"
	WindowsNvidiaURL   = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64.zip"
	WindowsAMDURL      = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-rocm.zip"
	WindowsIntelArcURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ipex-llm-ollama.zip"
	WindowsBaseURL     = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ollama-windows-amd64-base.zip"

	// Linux download URLs
	LinuxURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/linux/OllamaSetup.exe"

	// macOS download URLs
	MacOSIntelURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/macos/Ollama-darwin.zip"

	// Archive commands
	TarCommand     = "tar"
	TarExtractFlag = "-xf"
	TarDestFlag    = "-C"
	UnzipCommand   = "unzip"
	UnzipDestFlag  = "-d"
	MoveCommand    = "mv"
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
			cmd := exec.Command(o.EngineConfig.ExecPath + "/" + OllamaBatchFile)
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
	dataDir, err := utils.GetAOGDataDir()
	if err != nil {
		slog.Error("Get Byze data dir failed", "error", err)
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
			downloadUrl = WindowsAllGPUURL
		case types.GPUTypeNvidia:
			downloadUrl = WindowsNvidiaURL
		case types.GPUTypeAmd:
			downloadUrl = WindowsAMDURL
		case types.GPUTypeIntelArc:
			execPath = fmt.Sprintf("%s/%s", userDir, "ipex-llm-ollama")
			downloadUrl = WindowsIntelArcURL
		default:
			downloadUrl = WindowsBaseURL
		}

	case "linux":
		execFile = "ollama"
		execPath = fmt.Sprintf("%s/%s", userDir, "ollama")
		downloadUrl = LinuxURL
	case "darwin":
		execFile = "ollama"
		execPath = fmt.Sprintf("/%s/%s/%s/%s/%s", "Applications", "Ollama.app", "Contents", "Resources")
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

	} else {
		if utils.IpexOllamaSupportGPUStatus() {
			// 解压文件
			userDir, err := os.UserHomeDir()
			if err != nil {
				slog.Error("Get user home dir failed: ", err.Error())
				return err
			}
			ipexPath := fmt.Sprintf("%s/%s", userDir)
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				if runtime.GOOS == "windows" {
					unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
					if err := unzipCmd.Run(); err != nil {
						return fmt.Errorf("failed to unzip file: %v", err)
					}
				}
			}

		} else if runtime.GOOS == "windows" {
			ipexPath := o.EngineConfig.ExecPath
			if _, err = os.Stat(ipexPath); os.IsNotExist(err) {
				os.MkdirAll(ipexPath, 0o755)
				unzipCmd := exec.Command(TarCommand, TarExtractFlag, file, TarDestFlag, ipexPath)
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
