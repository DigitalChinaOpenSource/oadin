package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"oadin/internal/client"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/types"
	"oadin/internal/utils"

	"gopkg.in/yaml.v3"
)

const (
	// Default configuration for llamacpp
	llamacppDefaultPort  = "16697"
	llamacppDefaultHost  = constants.DefaultHost + ":" + llamacppDefaultPort
	llamacppServerExec   = "llama-swap.exe"
	LlamaVulkanPath      = "llamacpp-windows-vulkan"
	LlamaCppPath         = "llama-b5757-bin-win-vulkan-x64"
	LlamaSwapPath        = "llama-swap_148_windows_amd64"
	LlamaSwapConfigFile  = "config.yaml"
	llamacppDefaultModel = "Qwen3-8B-GGUF"
	LlamaWhisperPath     = "whisper-1.7.6_windows_amd64"

	// Windows download URLs for llamacpp
	llamacppWindowsBaseURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/llamacpp-windows-vulkan.zip"
	// 模型默认都在modelscope的
	// ggml-org组织下载 	https://www.modelscope.cn/organization/ggml-org
	// Embedding-GGUF	   https://www.modelscope.cn/organization/Embedding-GGUF
)

type llamacppProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

type LlamacppSwapServerConfig struct {
	StartPort int              `yaml:"startPort"`
	LogLevel  string           `yaml:"logLevel"`
	Models    map[string]Model `yaml:"models"`
	Groups    map[string]Group `yaml:"groups,omitempty"`
}

type Model struct {
	Cmd           string `yaml:"cmd"`
	CheckEndpoint string `yaml:"checkEndpoint,omitempty"`
	TTL           int    `yaml:"ttl"`
	UseModelName  string `yaml:"useModelName,omitempty"`
}

type Group struct {
	Persistent bool     `yaml:"persistent,omitempty"`
	Swap       bool     `yaml:"swap,omitempty"`
	Exclusive  bool     `yaml:"exclusive,omitempty"`
	Members    []string `yaml:"members,omitempty"`
}

func pullModelGetOrig(Model, ModelType string) string {
	// 如果ModelType为空或者chat或者generate，则默认为chat,则组织为ggml-org
	// 如果是embed，则组织为Embedding-GGUF
	if ModelType == "" || ModelType == types.ServiceChat || ModelType == types.ServiceGenerate {
		return fmt.Sprintf("ggml-org/%s", Model)
	} else if ModelType == types.ServiceEmbed {
		return fmt.Sprintf("Embedding-GGUF/%s", Model)
	}
	return fmt.Sprintf("ggml-org/%s", Model)
}

func getModelFiles(ctx context.Context, reqData *ModelScopeFileReqData) ([]ModelScopeFile, error) {
	c := GetHttpClient()
	filesReqPath := fmt.Sprintf(ModelScopeGetModelFilesReqPath, reqData.ModelName, reqData.Revision, reqData.Recursive)
	var resp *ModelScopeFileRespData
	err := c.Do(ctx, "GET", filesReqPath, nil, &resp)
	if err != nil {
		return []ModelScopeFile{}, err
	}
	var newResp []ModelScopeFile
	for _, file := range resp.Data.Files {
		// 打印下文件的名称
		logger.EngineLogger.Debug("[llamacpp] getModelFiles file: " + file.Name)
		if file.Name == ".gitignore" || file.Name == ".gitmodules" || file.Type == "tree" {
			continue
		}
		// 如果文件包含mmproj，还包含Q8_0，则加入
		if strings.Contains(file.Name, "mmproj") && strings.Contains(file.Name, "Q8_0") {
			newResp = append(newResp, file)
		}
		// 如果文件包含不包含mmproj，包含Q4_K_M，则加入
		if !strings.Contains(file.Name, "mmproj") && strings.Contains(file.Name, "Q4_K_M") {
			newResp = append(newResp, file)
		}
		// 如果文件包含不包含mmproj，包含q4_k_m，则加入
		if !strings.Contains(file.Name, "mmproj") && strings.Contains(file.Name, "q4_k_m") {
			newResp = append(newResp, file)
		}
	}
	return newResp, err
}

func asyncDownloadModelFile(ctx context.Context, a AsyncDownloadModelFileData, engine *llamacppProvider) {
	defer close(a.DataCh)
	defer close(a.ErrCh)

	for _, fileData := range a.ModelFiles {
		if downloadSingleFileCheck(ctx, a, fileData) {
			logger.EngineLogger.Debug("[llamacpp] Downloaded skip file: " + fileData.Name)
			continue // 如果文件已存在且完整，跳过下载
		}
		if err := downloadSingleFile(ctx, a, fileData); err != nil {
			a.ErrCh <- err
			logger.EngineLogger.Error("[llamacpp] Failed to download file: " + fileData.Name + " " + err.Error())
			return
		}
		logger.EngineLogger.Debug("[llamacpp] Downloaded file: " + fileData.Name)
	}

	logger.EngineLogger.Debug("[llamacpp] Generating generateCmdTxt for model: ", a.ModelName)
	if err := engine.addModelToConfig(a.ModelName, a.ModelType); err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to add model to config: " + err.Error())
		a.ErrCh <- errors.New("Failed to add model to config: " + err.Error())
		return
	}

	logger.EngineLogger.Info("[llamacpp] Pull model completed: " + a.ModelName)
	resp := types.ProgressResponse{Status: "success"}
	if data, err := json.Marshal(resp); err == nil {
		a.DataCh <- data
	} else {
		a.ErrCh <- err
	}
}

func NewLlamacppProvider(config *types.EngineRecommendConfig) *llamacppProvider {
	if config != nil {
		return &llamacppProvider{
			EngineConfig: config,
		}
	}
	OADINDir, err := utils.GetOADINDataDir()
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Get OADIN data dir failed: " + err.Error())
		return nil
	}

	downloadPath := filepath.Join(OADINDir, "engine", "llamacpp")
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
		execFile = filepath.Join(l.EngineConfig.ExecPath, l.EngineConfig.ExecFile)

	case "darwin":
		logger.EngineLogger.Warn("[llamacpp] Darwin system does not support llamacpp yet")

	case "linux":
		logger.EngineLogger.Warn("[llamacpp] linux system does not support llamacpp yet")

	default:
		logger.EngineLogger.Error("[llamacpp] unsupported operating system: " + runtime.GOOS)
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	if mode == types.EngineStartModeDaemon {
		LlamaSwapConfigFilePath := filepath.Join(l.EngineConfig.ExecPath, LlamaSwapConfigFile)
		logger.EngineLogger.Info("[llamacpp] exec file path: ", execFile, l.EngineConfig.Host, LlamaSwapConfigFilePath)
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
		pidFile := filepath.Join(rootPath, "llamacpp.pid")
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
	pidFile := filepath.Join(rootPath, "llamacpp.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		logger.EngineLogger.Info("[LLAMACPP] Stop openvino Model Server not found pidfile: " + pidFile)
		return nil
	}

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
		logger.EngineLogger.Error("Get OADIN data dir failed: " + err.Error())
		return nil
	}

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := filepath.Join(dataDir, "engine", "llamacpp")
	switch runtime.GOOS {
	case "windows":
		execFile = llamacppServerExec
		execPath = filepath.Join(enginePath, LlamaVulkanPath, LlamaSwapPath)

		switch utils.DetectGpuModel() {
		case types.GPUTypeAmd:
			downloadUrl = llamacppWindowsBaseURL
		default:
			downloadUrl = llamacppWindowsBaseURL
		}

	case "linux":
		logger.EngineLogger.Warn("[llamacpp] linux system does not support llamacpp yet")

	case "darwin":
		logger.EngineLogger.Warn("[llamacpp] Darwin system does not support llamacpp yet")

	default:
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           llamacppDefaultHost,
		Origin:         constants.DefaultHost,
		Scheme:         types.ProtocolHTTP,
		EnginePath:     enginePath,
		RecommendModel: llamacppDefaultModel,
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
	// create models directory if it doesn't exist
	modelDir := filepath.Join(l.EngineConfig.EnginePath, "models")
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		err := os.MkdirAll(modelDir, 0o750)
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] Failed to create models directory: " + err.Error())
			return err
		}
	}

	file, err := utils.DownloadFile(l.EngineConfig.DownloadUrl, l.EngineConfig.DownloadPath)
	if err != nil {
		return fmt.Errorf("failed to download llamacpp-server: %v, url: %v", err, l.EngineConfig.DownloadUrl)
	}

	logger.EngineLogger.Info("[llamacpp] start install llamacpp...")

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
		logger.EngineLogger.Info("[llamacpp] llamacpp extracted to engine directory", "path", engineDir)

	case "darwin":
		logger.EngineLogger.Warn("[llamacpp] darwin installation not implemented yet")

	case "linux":
		logger.EngineLogger.Warn("[llamacpp] Linux installation not implemented yet")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// 新建 config.yaml 空 文件
	LlamaSwapConfigFilePath := filepath.Join(l.EngineConfig.ExecPath, LlamaSwapConfigFile)
	logger.EngineLogger.Info("[llamacpp] Create config.yaml file at: " + LlamaSwapConfigFilePath)
	if _, err := os.Stat(LlamaSwapConfigFilePath); os.IsNotExist(err) {
		file, err := os.Create(LlamaSwapConfigFilePath)
		defer file.Close()
		if err != nil {
			logger.EngineLogger.Error("[llamacpp] Failed to create config.yaml: " + err.Error())
			return err
		}
	}
	// 写入默认配置
	defaultConfig := LlamacppSwapServerConfig{
		StartPort: 10001,
		LogLevel:  "info",
		Models:    map[string]Model{},
		Groups:    map[string]Group{},
	}
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to Marshal " + err.Error())
		return err
	}

	err = os.WriteFile(l.getConfigPath(), data, 0o644)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to WriteFile " + err.Error())
		return err
	}

	logger.EngineLogger.Info("[llamacpp] llamacpp model engine install completed")
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

func (l *llamacppProvider) getConfigPath() string {
	LlamaSwapConfigFilePath := filepath.Join(l.EngineConfig.ExecPath, LlamaSwapConfigFile)
	logger.EngineLogger.Info("[llamacpp] getConfigPath: ", LlamaSwapConfigFilePath)
	return LlamaSwapConfigFilePath
}

func (l *llamacppProvider) loadConfig() (*LlamacppSwapServerConfig, error) {
	configPath := l.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to read config file: " + err.Error())
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config LlamacppSwapServerConfig
	// 使用 yaml 解析配置文件
	if err := yaml.Unmarshal(data, &config); err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to unmarshal yaml config: " + err.Error())
		return nil, fmt.Errorf("failed to unmarshal yaml config: %v", err)
	}

	return &config, nil
}

func (l *llamacppProvider) saveConfig(config *LlamacppSwapServerConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to marshal yaml config: " + err.Error())
		return fmt.Errorf("failed to marshal yaml config: %v", err)
	}

	err = os.WriteFile(l.getConfigPath(), data, 0o644)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] saveConfig WriteFile: ", err.Error())
		return fmt.Errorf("failed to saveConfig WriteFile: %v", err)
	}

	// 重启后读取最新的配置文件
	l.StopEngine(context.Background())
	l.StartEngine("daemon")
	return nil
}

func (l *llamacppProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	config, err := l.loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		logger.EngineLogger.Error("[llamacpp] Failed to load config: " + err.Error())
		return nil, err
	}

	modelList := make([]types.ListModelResponse, 0)
	for key := range config.Models {
		modelList = append(modelList, types.ListModelResponse{
			Name: key,
		})
	}

	return &types.ListResponse{
		Models: modelList,
	}, nil
}

func (l *llamacppProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	req.Model = pullModelGetOrig(req.Model, req.ModelType)
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray
	dataCh := make(chan []byte)
	errCh := make(chan error)
	defer close(dataCh)
	defer close(errCh)
	localModelPath := filepath.Join(l.EngineConfig.EnginePath, "models", req.Model)
	if _, err := os.Stat(localModelPath); err != nil {
		_ = os.MkdirAll(localModelPath, 0o755)
	}
	fileReq := &ModelScopeFileReqData{
		ModelName: req.Model,
		Revision:  ModelScopeRevision,
		Recursive: "True",
	}
	modelFiles, err := getModelFiles(ctx, fileReq)
	if err != nil {
		errCh <- err
		return dataCh, errCh
	}
	if len(modelFiles) == 0 {
		errCh <- errors.New("no model files found")
		return dataCh, errCh
	}
	sort.Slice(modelFiles, func(i, j int) bool {
		return modelFiles[i].Size > modelFiles[j].Size
	})

	newDataCh := make(chan []byte)
	newErrorCh := make(chan error, 1)

	AsyncDownloadFuncParams := &AsyncDownloadModelFileData{
		ModelFiles:     modelFiles,
		ModelName:      req.Model,
		DataCh:         newDataCh,
		ErrCh:          newErrorCh,
		LocalModelPath: localModelPath,
		ModelType:      req.ModelType,
	}
	go asyncDownloadModelFile(ctx, *AsyncDownloadFuncParams, l)

	return newDataCh, newErrorCh
}

func (l *llamacppProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	err := l.UnloadModel(ctx, &types.UnloadModelRequest{Models: []string{req.Model}})
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to unload model: " + err.Error())
		return err
	}

	// 更新配置文件
	err = l.deleteModelToConfig(req.Model)
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to deleteModelToConfig: " + err.Error())
		return err
	}

	req.Model = pullModelGetOrig(req.Model, "")
	modelDir := filepath.Join(l.EngineConfig.EnginePath, "models", req.Model)
	if err := os.RemoveAll(modelDir); err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to remove model directory: " + err.Error())
		return err
	}

	return nil
}

func (l *llamacppProvider) deleteModelToConfig(modelName string) error {
	config, err := l.loadConfig()
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to load config: ", err.Error())
		return err
	}

	delete(config.Models, modelName)

	return l.saveConfig(config)
}

func (l *llamacppProvider) addModelToConfig(modelName, modelType string) error {
	config, err := l.loadConfig()
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to load config: ", err.Error())
		return err
	}

	for key := range config.Models {
		if key == modelName {
			return nil
		}
	}

	if config.Models == nil {
		config.Models = map[string]Model{}
	}

	cmd := l.generateCmdTxt(modelName, modelType)
	logger.EngineLogger.Info("[llamacpp] addModelToConfig cmd: ", cmd)
	newModel := Model{
		Cmd:           cmd,
		CheckEndpoint: "/health",
		TTL:           300,
	}
	// modelName为ggml-org/Qwen2.5-VL-7B-Instruct-GGUF，这里处理以下，只获取最后面的模型名字
	modelName = filepath.Base(modelName)
	config.Models[modelName] = newModel

	return l.saveConfig(config)
}

const (
	ServiceChatVlCmd = `%s\llama-server.exe -m "%s\models\%s\%s-Q4_K_M.gguf" --mmproj "%s\models\%s\mmproj-%s-Q8_0.gguf" -ngl 99 --port ${PORT}`
	ServiceChatCmd   = `%s\llama-server.exe -m "%s\models\%s\%s-Q4_K_M.gguf" -ngl 99 --port ${PORT}`
	ServiceEmbedCmd  = `%s\llama-server.exe -m "%s\models\%s\%s-q4_k_m.gguf" --embedding --pooling cls -ub 8192 -ngl 99 --port ${PORT}`

	// ServiceSpeechToTextCmd = `%s\whisper-server.exe -m "%s\models\%s\%s-Q4_K_M.gguf" --ngl 99 --port ${PORT}`
)

func (l *llamacppProvider) generateCmdTxt(modelName, modelType string) string {
	execPath := filepath.Join(l.EngineConfig.EnginePath, LlamaVulkanPath, LlamaCppPath)
	enginePath := l.EngineConfig.EnginePath
	modelNameSample := strings.Replace(filepath.Base(modelName), "-GGUF", "", 1)
	modelName = strings.Replace(modelName, "/", "\\", 1)
	logger.EngineLogger.Debug("[llamacpp] generateCmdTxt execPath: ", execPath)
	logger.EngineLogger.Debug("[llamacpp] generateCmdTxt enginePath: ", enginePath)
	logger.EngineLogger.Debug("[llamacpp] generateCmdTxt modelNameSample: ", modelNameSample)

	var template string
	switch modelType {
	case types.ServiceChat:
		// 如果 modelName 包含 "vl"（不区分大小写），使用 ServiceChatVlCmd，否则用 ServiceChatCmd
		if strings.Contains(strings.ToLower(modelName), "vl") {
			template = fmt.Sprintf(ServiceChatVlCmd, execPath, enginePath, modelName, modelNameSample, enginePath, modelName, modelNameSample)
		} else {
			template = fmt.Sprintf(ServiceChatCmd, execPath, enginePath, modelName, modelNameSample)
		}
	case types.ServiceGenerate:
		template = fmt.Sprintf(ServiceChatCmd, execPath, enginePath, modelName, modelNameSample)
	case types.ServiceEmbed:
		template = fmt.Sprintf(ServiceEmbedCmd, execPath, enginePath, modelName, modelNameSample)
	default:
		logger.EngineLogger.Error("[llamacpp] Unsupported model type: " + modelType)
		return ""
	}

	return template
}

func (l *llamacppProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	req.Model = pullModelGetOrig(req.Model, req.ModelType)
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray
	localModelPath := filepath.Join(l.EngineConfig.EnginePath, "models", req.Model)
	logger.EngineLogger.Info("[llamacpp] localModelPath: ", localModelPath)

	if _, err := os.Stat(localModelPath); err != nil {
		_ = os.MkdirAll(localModelPath, 0o755)
	}

	fileReq := &ModelScopeFileReqData{
		ModelName: req.Model,
		Revision:  ModelScopeRevision,
		Recursive: "True",
	}
	modelFiles, err := getModelFiles(ctx, fileReq)
	if err != nil {
		return nil, err
	}

	logger.EngineLogger.Info("[llamacpp] modelFiles: " + fmt.Sprintf("%+v", modelFiles))

	if len(modelFiles) == 0 {
		return nil, errors.New("no model files found")
	}
	sort.Slice(modelFiles, func(i, j int) bool {
		return modelFiles[i].Size > modelFiles[j].Size
	})

	newDataCh := make(chan []byte)
	newErrorCh := make(chan error, 1)

	AsyncDownloadFuncParams := &AsyncDownloadModelFileData{
		ModelFiles:     modelFiles,
		ModelType:      req.ModelType,
		ModelName:      req.Model,
		DataCh:         newDataCh,
		ErrCh:          newErrorCh,
		LocalModelPath: localModelPath,
	}
	go asyncDownloadModelFile(ctx, *AsyncDownloadFuncParams, l)

	// 用于标记是否成功完成下载
	downloadDone := false

	for {
		select {
		case data, ok := <-newDataCh:
			if !ok {
				// dataCh 关闭 -> 下载完成
				if data == nil {
					downloadDone = true
				}
			}
			// data 可用于进度通知
			if fn != nil && data != nil {
				// fn(data) // 进度回调
				fmt.Printf("进度回调")
			}
		case err, ok := <-newErrorCh:
			if ok && err != nil {
				return nil, err
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		// 下载完成且错误通道关闭了
		if downloadDone && len(newErrorCh) == 0 {
			break
		}
	}
	return &types.ProgressResponse{}, nil
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
	config, err := l.loadConfig()
	if err != nil {
		logger.EngineLogger.Error("[llamacpp] Failed to load config: " + err.Error())
		return nil, err
	}

	modelList := make([]types.ListModelResponse, 0)
	for key := range config.Models {
		modelList = append(modelList, types.ListModelResponse{
			Name: key,
		})
	}

	return &types.ListResponse{
		Models: modelList,
	}, nil
}

func (l *llamacppProvider) GetOperateStatus() int {
	return 0
}

func (l *llamacppProvider) SetOperateStatus(status int) {}
