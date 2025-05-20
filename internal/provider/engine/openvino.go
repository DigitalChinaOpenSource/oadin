package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"intel.com/aog/internal/client"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

type OpenvinoProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func NewOpenvinoProvider(config *types.EngineRecommendConfig) *OpenvinoProvider {
	if config != nil {
		return &OpenvinoProvider{
			EngineConfig: config,
		}
	}

	AOGDir, err := utils.GetAOGDataDir()
	if err != nil {
		logger.EngineLogger.Error("[OpenVINO] Get AOG data dir failed: " + err.Error())
		return nil
	}

	openvinoPath := fmt.Sprintf("%s/%s/%s", AOGDir, "engine", "openvino")
	if _, err := os.Stat(openvinoPath); os.IsNotExist(err) {
		err := os.MkdirAll(openvinoPath, 0o750)
		if err != nil {
			logger.EngineLogger.Error("[OpenVINO] Create openvino path failed: " + err.Error())
			return nil
		}
	}

	openvinoProvider := new(OpenvinoProvider)
	openvinoProvider.EngineConfig = openvinoProvider.GetConfig()

	return openvinoProvider
}

func (o *OpenvinoProvider) GetDefaultClient() *client.GRPCClient {
	grpcClient, err := client.NewGRPCClient("localhost:9000")
	if err != nil {
		logger.EngineLogger.Error("[OpenVINO] Failed to create gRPC client: " + err.Error())
		return nil
	}

	return grpcClient
}

func (o *OpenvinoProvider) StartEngine(mode string) error {
	logger.EngineLogger.Info("[OpenVINO] Start engine mode: " + mode)
	// 当前暂仅支持windows
	if runtime.GOOS != "windows" {
		logger.EngineLogger.Error("[OpenVINO] Unsupported OS: " + runtime.GOOS)
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	rootPath, err := utils.GetAOGDataDir()
	if err != nil {
		logger.EngineLogger.Error("[OpenVINO] Get AOG data dir failed: " + err.Error())
		return fmt.Errorf("failed get aog dir: %v", err)
	}

	modelDir := fmt.Sprintf("%s/models", o.EngineConfig.EnginePath)
	pidFile := fmt.Sprintf("%s/ovms.pid", rootPath)

	batchContent := fmt.Sprintf(`
	@echo on
	call "%s\\setupvars.bat"
	set PATH=%s\\python\\Scripts;%%PATH%%
	set HF_HOME=%s\\.cache
	set HF_ENDPOINT=https://hf-mirror.com
	%s --port 9000 --rest_port 16666 --config_path %s\\config.json
	`,
		o.EngineConfig.ExecPath,
		o.EngineConfig.ExecPath,
		o.EngineConfig.EnginePath,
		o.EngineConfig.ExecFile,
		modelDir,
	)

	logger.EngineLogger.Debug("[OpenVINO] Batch content: " + batchContent)
	BatchFile := filepath.Join(o.EngineConfig.ExecPath, "start_ovms.bat")
	if _, err = os.Stat(BatchFile); err != nil {
		if err = os.WriteFile(BatchFile, []byte(batchContent), 0o644); err != nil {
			logger.EngineLogger.Error("[OpenVINO] Failed to create batch file: " + err.Error())
			return fmt.Errorf("failed to create temp batch file: %v", err)
		}
	}

	cmd := exec.Command("cmd", "/C", BatchFile)
	if mode == types.EngineStartModeStandard {
		cmd = exec.Command("cmd", "/C", "start", BatchFile)
	}
	cmd.Dir = o.EngineConfig.EnginePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		logger.EngineLogger.Error("[OpenVINO] Failed to start OpenVINO Model Server: " + err.Error())
		return err
	}
	time.Sleep(500 * time.Microsecond)

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		logger.EngineLogger.Error("[OpenVINO] Failed to write PID to file: " + err.Error())
		if killErr := cmd.Process.Kill(); killErr != nil {
			logger.EngineLogger.Error("[OpenVINO] Failed to kill process after PID write error: " + killErr.Error())
		}
		return err
	}

	go func() {
		cmd.Wait()
	}()

	logger.EngineLogger.Info("[OpenVINO] OpenVINO Model Server started successfully")
	return nil
}

func (o *OpenvinoProvider) StopEngine() error {
	pidFile := "ovms.pid"
	data, err := os.ReadFile(pidFile)
	if err != nil {
		slog.Error("Failed to read PID file", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to read PID file: " + err.Error())
		return err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		slog.Error("Failed to parse PID", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Invalid PID format: " + err.Error())
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		slog.Error("Failed to find process", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to find process: " + err.Error())
		return err
	}
	err = process.Kill()
	if err != nil {
		slog.Error("Failed to kill process", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to kill process: " + err.Error())
		return err
	}

	err = os.Remove(pidFile)
	if err != nil {
		slog.Error("Failed to remove PID file", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to remove PID file: " + err.Error())
		return err
	}

	return nil
}

func (o *OpenvinoProvider) GetConfig() *types.EngineRecommendConfig {
	downloadPath, err := utils.GetDownloadDir()
	if _, err = os.Stat(downloadPath); os.IsNotExist(err) {
		err = os.MkdirAll(downloadPath, 0o755)
		if err != nil {
			slog.Error("Create download path failed: " + err.Error())
			logger.EngineLogger.Error("[OpenVINO] Create download path failed: " + err.Error())
			return nil
		}
	}

	AOGDir, err := utils.GetAOGDataDir()
	if err != nil {
		slog.Error("Get AOG data dir failed: " + err.Error())
		logger.EngineLogger.Error("[OpenVINO] Get AOG data dir failed: " + err.Error())
		return nil
	}

	execFile := ""
	execPath := ""
	downloadUrl := ""
	enginePath := ""
	switch runtime.GOOS {
	case "windows":
		execPath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino/ovms")
		execFile = "ovms.exe"
		downloadUrl = "http://120.232.136.73:31619/aogdev/ovms_windows.zip"
		enginePath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino")
	case "linux":
		// todo 这里需要区分 centos 和 ubuntu(22/24) 的版本 后续实现
		execFile = "ovms"
		execPath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino/ovms")
		downloadUrl = ""
		enginePath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino")
	default:
		slog.Error("Unsupported OS: " + runtime.GOOS)
		logger.EngineLogger.Error("[OpenVINO] Unsupported OS: " + runtime.GOOS)
		return nil
	}

	return &types.EngineRecommendConfig{
		Host:           "127.0.0.1:16666",
		Origin:         "127.0.0.1",
		Scheme:         "http",
		RecommendModel: "stable-diffusion-v-1-5-ov-fp16",
		DownloadUrl:    downloadUrl,
		DownloadPath:   downloadPath,
		EnginePath:     enginePath,
		ExecPath:       execPath,
		ExecFile:       execFile,
	}
}

func (o *OpenvinoProvider) HealthCheck() error {
	c := o.GetDefaultClient()
	health, err := c.ServerLive()
	if err != nil || !health.GetLive() {
		logger.EngineLogger.Debug("[OpenVINO] OpenVINO Model Server is not healthy: " + err.Error())
		return err
	}

	logger.EngineLogger.Info("[OpenVINO] OpenVINO Model Server is healthy")
	return nil
}

func (o *OpenvinoProvider) GetVersion(ctx context.Context, resp *types.EngineVersionResponse) (*types.EngineVersionResponse, error) {
	return &types.EngineVersionResponse{
		Version: "2025.0.0",
	}, nil
}

func (o *OpenvinoProvider) InstallEngine() error {
	modelDir := fmt.Sprintf("%s/models", o.EngineConfig.EnginePath)
	if _, err := os.Stat(modelDir); os.IsNotExist(err) {
		err := os.MkdirAll(modelDir, 0o750)
		if err != nil {
			slog.Error("Failed to create models directory", "error", err)
			logger.EngineLogger.Error("[OpenVINO] Failed to create models directory: " + err.Error())
			return err
		}
	}

	file, err := utils.DownloadFile(o.EngineConfig.DownloadUrl, o.EngineConfig.DownloadPath)
	if err != nil {
		slog.Error("Failed to download OpenVINO Model Server", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to download OpenVINO Model Server: " + err.Error())
		return fmt.Errorf("failed to download ovms: %v", err)
	}

	// 解压ovms文件
	err = utils.UnzipFile(file, o.EngineConfig.EnginePath)
	if err != nil {
		slog.Error("Failed to unzip OpenVINO Model Server", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to unzip OpenVINO Model Server: " + err.Error())
		return fmt.Errorf("failed to unzip ovms: %v", err)
	}

	// 下载py 脚本文件压缩包
	// scriptZipUrl := "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/byze/windows/scripts.zip"
	scriptZipUrl := "http://120.232.136.73:31619/aogdev/scripts.zip"
	scriptZipFile, err := utils.DownloadFile(scriptZipUrl, o.EngineConfig.EnginePath)
	if err != nil {
		slog.Error("Failed to download scripts.zip", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to download scripts.zip: " + err.Error())
		return fmt.Errorf("failed to download scripts.zip: %v", err)
	}

	// 解压py 脚本文件
	err = utils.UnzipFile(scriptZipFile, o.EngineConfig.EnginePath)
	if err != nil {
		slog.Error("Failed to unzip scripts.zip", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to unzip scripts.zip: " + err.Error())
		return fmt.Errorf("failed to unzip scripts.zip: %v", err)
	}

	// 新建 config.json 空 文件
	configFile := fmt.Sprintf("%s/config.json", modelDir)
	_, err = os.Create(configFile)
	if err != nil {
		slog.Error("Failed to create config.json", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to create config.json: " + err.Error())
		return fmt.Errorf("failed to create config.json: %v", err)
	}
	// 写入默认config配置
	defaultConfig := OpenvinoModelServerConfig{
		MediapipeConfigList: []ModelConfig{},
		ModelConfigList:     []interface{}{},
	}
	err = o.saveConfig(&defaultConfig)
	if err != nil {
		slog.Error("Failed to save config.json", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to save config.json: " + err.Error())
		return fmt.Errorf("failed to save config.json: %v", err)
	}
	execPath := strings.Replace(o.EngineConfig.ExecPath, "/", "\\", -1)
	enginePath := strings.Replace(o.EngineConfig.EnginePath, "/", "\\", -1)

	// 1. 构造批处理命令（确保所有命令在同一个会话中执行）
	batchContent := fmt.Sprintf(`
	@echo on
	call "%s\\setupvars.bat"
	set PATH=%s\\python\\Scripts;%%PATH%%
	python -m pip install -r "%s\\scripts\\requirements.txt" -i https://mirrors.aliyun.com/pypi/simple/
	`, execPath, execPath, enginePath)

	logger.EngineLogger.Debug("[OpenVINO] Batch content: " + batchContent)

	// 2. 创建临时批处理文件
	tmpBatchFile := filepath.Join(os.TempDir(), "run_install.bat")
	if err := os.WriteFile(tmpBatchFile, []byte(batchContent), 0o644); err != nil {
		slog.Error("Failed to create temp batch file", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to create temp batch file: " + err.Error())
		return fmt.Errorf("failed to create temp batch file: %v", err)
	}
	defer os.Remove(tmpBatchFile) // 执行后删除临时文件

	// 3. 执行批处理文件
	cmd := exec.Command("cmd", "/C", tmpBatchFile)
	cmd.Dir = enginePath

	var stdout, stderr bytes.Buffer

	// 实时输出 stdout 和 stderr
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout) // 同时输出到控制台和缓冲区
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr) // 同时输出到控制台和缓冲区

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to run batch script", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to run batch script: " + err.Error())
		return fmt.Errorf("failed to run batch script: %v\nStdout: %s\nStderr: %s",
			err, stdout.String(), stderr.String())
	}

	slog.Info("[Install Engine] openvino model engine install completed")
	logger.EngineLogger.Info("[OpenVINO] OpenVINO Model Server install completed")

	return nil
}

func (o *OpenvinoProvider) InitEnv() error {
	// todo  set env
	return nil
}

type ModelConfig struct {
	Name      string `json:"name"`
	BasePath  string `json:"base_path"`
	GraphPath string `json:"graph_path"`
}

type OpenvinoModelServerConfig struct {
	MediapipeConfigList []ModelConfig `json:"mediapipe_config_list"`
	ModelConfigList     []interface{} `json:"model_config_list"`
}

func (o *OpenvinoProvider) getConfigPath() string {
	return fmt.Sprintf("%s/models/config.json", o.EngineConfig.EnginePath)
}

func (o *OpenvinoProvider) loadConfig() (*OpenvinoModelServerConfig, error) {
	configPath := o.getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Error("Failed to read config file", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to read config file: " + err.Error())
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config OpenvinoModelServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		slog.Error("Failed to unmarshal config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to unmarshal config: " + err.Error())
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}

func (o *OpenvinoProvider) saveConfig(config *OpenvinoModelServerConfig) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		slog.Error("Failed to marshal config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to marshal config: " + err.Error())
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	return os.WriteFile(o.getConfigPath(), data, 0o644)
}

func (o *OpenvinoProvider) ListModels(ctx context.Context) (*types.ListResponse, error) {
	config, err := o.loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to load config: " + err.Error())
		return nil, err
	}

	modelList := make([]types.ListModelResponse, 0)
	for _, model := range config.MediapipeConfigList {
		modelList = append(modelList, types.ListModelResponse{
			Name: model.Name,
		})
	}

	return &types.ListResponse{
		Models: modelList,
	}, nil
}

func (o *OpenvinoProvider) PullModelStream(ctx context.Context, req *types.PullModelRequest) (chan []byte, chan error) {
	dataCh := make(chan []byte)
	errCh := make(chan error)
	return dataCh, errCh
}

func (o *OpenvinoProvider) DeleteModel(ctx context.Context, req *types.DeleteRequest) error {
	config, err := o.loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to load config: " + err.Error())
		return err
	}

	for i, model := range config.MediapipeConfigList {
		if model.Name == req.Model {
			config.MediapipeConfigList = append(config.MediapipeConfigList[:i], config.MediapipeConfigList[i+1:]...)
			err = o.saveConfig(config)
			if err != nil {
				slog.Error("Failed to save config after deleting model", "error", err)
				logger.EngineLogger.Error("[OpenVINO] Failed to save config after deleting model: " + err.Error())
				return err
			}

			modelDir := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, req.Model)
			if err := os.RemoveAll(modelDir); err != nil {
				slog.Error("Failed to remove model directory", "error", err)
				logger.EngineLogger.Error("[OpenVINO] Failed to remove model directory: " + err.Error())
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("model %s not found", req.Model)
}

func (o *OpenvinoProvider) addModelToConfig(modelName string) error {
	config, err := o.loadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to load config: " + err.Error())
		return err
	}

	for _, model := range config.MediapipeConfigList {
		if model.Name == modelName {
			return nil
		}
	}

	newModel := ModelConfig{
		Name: modelName,
		//BasePath:  o.EngineConfig.EnginePath + "/models",
		GraphPath: "graph.pbtxt",
	}
	config.MediapipeConfigList = append(config.MediapipeConfigList, newModel)

	return o.saveConfig(config)
}

func (o *OpenvinoProvider) generateGraphPbtxt(modelName, modelType string) error {
	modelDir := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, modelName)
	if err := os.MkdirAll(modelDir, 0o750); err != nil {
		slog.Error("Failed to create model directory", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to create model directory: " + err.Error())
		return err
	}

	enginePath := strings.Replace(o.EngineConfig.EnginePath, "\\", "/", -1)

	var template string
	switch modelType {
	case "text-to-image":
		template = fmt.Sprintf(`input_stream: "OVMS_PY_TENSOR:prompt"
input_stream: "OVMS_PY_TENSOR_BATCH:batch"
input_stream: "OVMS_PY_TENSOR_HEIGHT:height"
input_stream: "OVMS_PY_TENSOR_WIDTH:width"
output_stream: "OVMS_PY_TENSOR:image"

node {
  name: "%s"
  calculator: "PythonExecutorCalculator"
  input_side_packet: "PYTHON_NODE_RESOURCES:py"

  input_stream: "INPUT:prompt"
  input_stream: "BATCH:batch"
  input_stream: "HEIGHT:height"
  input_stream: "WIDTH:width"
  output_stream: "OUTPUT:image"
  node_options: {
    [type.googleapis.com/mediapipe.PythonExecutorCalculatorOptions]: {
      handler_path: "%s/scripts/text-to-image/model.py"
    }
  }
}`, modelName, enginePath)
	default:
		slog.Error("Unsupported model type: " + modelType)
		logger.EngineLogger.Error("[OpenVINO] Unsupported model type: " + modelType)
		return fmt.Errorf("unsupported model type: %s", modelType)
	}

	graphPath := fmt.Sprintf("%s/graph.pbtxt", modelDir)
	return os.WriteFile(graphPath, []byte(template), 0o644)
}

func (o *OpenvinoProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	// 当前暂时使用 modelscope 拉取模型
	// 后续使用 python 脚本拉取（区分 huggingface 和 modelscope）
	localModelPath := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, req.Model)
	scriptPath := fmt.Sprintf("%s/scripts/model.py", o.EngineConfig.EnginePath)

	logger.EngineLogger.Info("[OpenVINO] Pulling model: " + req.Model)

	if _, err := os.Stat(localModelPath); os.IsNotExist(err) {
		err := os.MkdirAll(localModelPath, 0o750)
		if err != nil {
			slog.Error("Failed to create model directory", "error", err)
			logger.EngineLogger.Error("[OpenVINO] Failed to create model directory: " + err.Error())
			return nil, err
		}
	}

	batchContent := fmt.Sprintf(`
		@echo on
		call "%s\\setupvars.bat"
		set PATH=%s\\python\\Scripts;%%PATH%%
		set HF_HOME=%s\\.cache
		set HF_ENDPOINT=https://hf-mirror.com
		python  %s --model_name %s --local_dir %s
		`,
		o.EngineConfig.ExecPath, o.EngineConfig.ExecPath, o.EngineConfig.EnginePath, scriptPath, req.Model, localModelPath)

	logger.EngineLogger.Debug("[OpenVINO] Batch content: " + batchContent)

	tmpBatchFile := filepath.Join(os.TempDir(), "pull_model.bat")
	if err := os.WriteFile(tmpBatchFile, []byte(batchContent), 0o644); err != nil {
		slog.Error("Failed to create temp batch file", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to create temp batch file: " + err.Error())
		return nil, err
	}
	defer os.Remove(tmpBatchFile)

	// 执行批处理文件并实时输出日志
	cmd := exec.Command("cmd", "/C", tmpBatchFile)
	cmd.Dir = o.EngineConfig.ExecPath

	// 实时输出 stdout 和 stderr
	cmd.Stdout = io.MultiWriter(os.Stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr)

	err := cmd.Run()
	if err != nil {
		slog.Error("Failed to pull model", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to pull model: " + err.Error())
		return nil, err
	}

	// 生成对应的graph.pbtxt文件
	logger.EngineLogger.Debug("[OpenVINO] Generating graph.pbtxt for model: " + req.Model)
	if err := o.generateGraphPbtxt(req.Model, req.ModelType); err != nil {
		slog.Error("Failed to generate graph.pbtxt", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to generate graph.pbtxt: " + err.Error())
		return nil, err
	}

	// 添加到配置
	logger.EngineLogger.Debug("[OpenVINO] Adding model to config: " + req.Model)
	if err := o.addModelToConfig(req.Model); err != nil {
		slog.Error("Failed to add model to config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to add model to config: " + err.Error())
		return nil, err
	}

	logger.EngineLogger.Info("[OpenVINO] Pull model completed: " + req.Model)

	return nil, nil
}
