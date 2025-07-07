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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"time"

	"github.com/google/uuid"

	"intel.com/aog/internal/client"
	"intel.com/aog/internal/constants"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

const (
	ModelScopeSCHEME               = "https"
	ModelScopeEndpointCN           = "www.modelscope.cn"
	ModelScopeEndpointAI           = "www.modelscope.ai"
	ModelScopeGetModelFilesReqPath = "/api/v1/models/%s/repo/files?Revision=%s&Recursive=%s"
	ModelScopeModelDownloadReqPath = "/api/v1/models/%s/repo?Revision=%s&FilePath=%s"
	ModelScopeRevision             = "master"
	BufferSize                     = 64 * 1024

	// OpenVINO Server配置
	OpenvinoGRPCPort     = "9000"
	OpenvinoGRPCHost     = constants.DefaultHost + ":" + OpenvinoGRPCPort
	OpenvinoHTTPPort     = "16666"
	OpenvinoHTTPHost     = constants.DefaultHost + ":" + OpenvinoHTTPPort
	OpenvinoVersion      = "2025.0.0"
	OpenvinoDefaultModel = "stable-diffusion-v-1-5-ov-fp16"

	// 下载URL
	OVMSWindowsDownloadURL = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/ovms_windows.zip"
	ScriptsDownloadURL     = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/scripts.zip"
)

type OpenvinoProvider struct {
	EngineConfig *types.EngineRecommendConfig
}

func QuotePlus(s string) string {
	// 先进行标准URL编码
	encoded := url.QueryEscape(s)
	// 将空格替换为+号（Python quote_plus的行为）
	encoded = strings.ReplaceAll(encoded, "%20", "+")
	// 特殊处理加号本身
	encoded = strings.ReplaceAll(encoded, "+", "%2B")
	return encoded
}

type ModelScopeFileReqData struct {
	ModelName string `json:"model_name"`
	Revision  string `json:"revision"`
	Recursive string `json:"recursive"`
	FilePath  string `json:"file_path"`
	Stream    bool   `json:"stream"`
}

type ModelScopeFileRespData struct {
	Code int                `json:"Code"`
	Data ModelScopeFileData `json:"Data"`
}

type ModelScopeFileData struct {
	Files []ModelScopeFile `json:"Files"`
}

type ModelScopeFile struct {
	Name     string `json:"Name"`
	Path     string `json:"Path"`
	Digest   string `json:"Sha256"`
	Size     int64  `json:"Size"`
	IsLFS    bool   `json:"IsLFS"`
	Revision string `json:"Revision"`
	Type     string `json:"Type"`
}

type AsyncDownloadModelFileData struct {
	ModelName      string
	ModelType      string
	DataCh         chan []byte
	ErrCh          chan error
	ModelFiles     []ModelScopeFile
	LocalModelPath string
}

func CheckFileDigest(ExceptDigest string, FilePath string) bool {
	// open file
	file, err := os.Open(FilePath)
	if err != nil {
		os.RemoveAll(FilePath)
		return false
	}
	defer file.Close()

	// create SHA-256
	hash := sha256.New()

	buf := make([]byte, BufferSize)

	// Read the file in chunks and update the hash
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return false
		}
		if n == 0 { // read finish
			break
		}

		hash.Write(buf[:n]) // update hash
	}
	hexDigest := hex.EncodeToString(hash.Sum(nil))
	if ExceptDigest != hexDigest {
		os.RemoveAll(FilePath)
		return false
	}
	return true
}

func GetHttpClient() *client.Client {
	d := GetModelScopeDomain(true)
	return client.NewClient(&url.URL{
		Scheme: ModelScopeSCHEME,
		Host:   d,
	}, http.DefaultClient)
}

func GetModelScopeDomain(cnSite bool) string {
	if cnSite {
		return ModelScopeEndpointCN
	} else {
		return ModelScopeEndpointAI
	}
}

func GetModelFiles(ctx context.Context, reqData *ModelScopeFileReqData) ([]ModelScopeFile, error) {
	c := GetHttpClient()
	filesReqPath := fmt.Sprintf(ModelScopeGetModelFilesReqPath, reqData.ModelName, reqData.Revision, reqData.Recursive)
	var resp *ModelScopeFileRespData
	err := c.Do(ctx, "GET", filesReqPath, nil, &resp)
	if err != nil {
		return []ModelScopeFile{}, err
	}
	var newResp []ModelScopeFile
	for _, file := range resp.Data.Files {
		if file.Name == ".gitignore" || file.Name == ".gitmodules" || file.Type == "tree" {
			continue
		}
		newResp = append(newResp, file)
	}
	return newResp, err
}

func DownloadModelFileRequest(ctx context.Context, reqData *ModelScopeFileReqData, reqHeader map[string]string) (chan []byte, chan error) {
	c := GetHttpClient()
	modelReqPath := fmt.Sprintf(ModelScopeModelDownloadReqPath, reqData.ModelName, reqData.Revision, reqData.FilePath)
	dataCh, errCh := c.StreamResponse(ctx, "GET", modelReqPath, nil, reqHeader)
	return dataCh, errCh
}

func AsyncDownloadModelFile(ctx context.Context, a AsyncDownloadModelFileData, engine *OpenvinoProvider) {
	defer close(a.DataCh)
	defer close(a.ErrCh)

	for _, fileData := range a.ModelFiles {
		if err := downloadSingleFile(ctx, a, fileData); err != nil {
			a.ErrCh <- err
			logger.EngineLogger.Error("[OpenVINO] Failed to download file: " + fileData.Name + " " + err.Error())
			return
		}
		logger.EngineLogger.Debug("[OpenVINO] Downloaded file: " + fileData.Name)
	}

	// 所有文件下载成功
	// 下载完成后再执行后续逻辑
	logger.EngineLogger.Debug("[OpenVINO] Generating graph.pbtxt for model: " + a.ModelName)
	if err := engine.generateGraphPBTxt(a.ModelName, a.ModelType); err != nil {
		slog.Error("Failed to generate graph.pbtxt", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to generate graph.pbtxt: " + err.Error())
		a.ErrCh <- errors.New("Failed to generate graph.pbtxt: " + err.Error())
		return
	}

	logger.EngineLogger.Debug("[OpenVINO] Adding model to config: " + a.ModelName)
	if err := engine.addModelToConfig(a.ModelName, a.ModelType); err != nil {
		slog.Error("Failed to add model to config", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to add model to config: " + err.Error())
		a.ErrCh <- errors.New("Failed to add model to config: " + err.Error())
		return
	}

	logger.EngineLogger.Info("[OpenVINO] Pull model completed: " + a.ModelName)
	resp := types.ProgressResponse{Status: "success"}
	if data, err := json.Marshal(resp); err == nil {
		a.DataCh <- data
	} else {
		a.ErrCh <- err
	}
}

func downloadSingleFile(ctx context.Context, a AsyncDownloadModelFileData, fileData ModelScopeFile) error {
	filePath := filepath.Join(a.LocalModelPath, fileData.Path)

	// 创建目录（如果需要）
	if strings.Contains(fileData.Path, "/") {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
			return err
		}
	}

	// 打开文件（追加模式）
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 获取当前文件长度（用于断点续传）
	partSize, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// 如果文件已存在并且大小匹配，进行 hash 校验
	if partSize >= fileData.Size {
		if CheckFileDigest(fileData.Digest, filePath) {
			return nil // 跳过下载
		}
		// 删除损坏文件，重新下载
		_ = os.Remove(filePath)
		return downloadSingleFile(ctx, a, fileData)
	}

	// 构造请求
	headers := map[string]string{
		"Range":               fmt.Sprintf("bytes=%d-%d", partSize, fileData.Size-1),
		"snapshot-identifier": uuid.New().String(),
		"X-Request-ID":        strings.ReplaceAll(uuid.New().String(), "-", ""),
	}
	fp := QuotePlus(fileData.Path)
	reqData := &ModelScopeFileReqData{
		ModelName: a.ModelName,
		Revision:  ModelScopeRevision,
		FilePath:  fp,
		Stream:    true,
	}

	reqDataCh, reqErrCh := DownloadModelFileRequest(ctx, reqData, headers)

	// 下载内容
	digest := sha256.New()
	for {
		select {
		case data, ok := <-reqDataCh:
			if !ok {
				// 检查文件完整性
				if partSize != fileData.Size {
					return fmt.Errorf("file %s incomplete: got %d bytes, expected %d", fileData.Name, partSize, fileData.Size)
				}

				// 先检查下载过程中计算的 digest
				downloadHash := hex.EncodeToString(digest.Sum(nil))
				if downloadHash != fileData.Digest {
					logger.EngineLogger.Warn("[OpenVINO] Download digest mismatch for file %s, recalculating from file: expected %s, got %s",
						fileData.Name, fileData.Digest, downloadHash)

					// 重新读取文件计算 digest
					if CheckFileDigest(fileData.Digest, filePath) {
						logger.EngineLogger.Info("[OpenVINO] File digest verification passed after recalculation for file: %s", fileData.Name)
						return nil
					} else {
						// 删除损坏文件，重新下载
						logger.EngineLogger.Error("[OpenVINO] File digest verification failed after recalculation for file: %s, will retry download", fileData.Name)
						_ = os.Remove(filePath)
						return downloadSingleFile(ctx, a, fileData)
					}
				}

				logger.EngineLogger.Debug("[OpenVINO] File download completed successfully: %s", fileData.Name)
				return nil // 完成
			}
			if len(data) == 0 {
				continue
			}
			n, err := f.Write(data)
			if err != nil {
				return err
			}
			digest.Write(data)
			partSize += int64(n)

			// 写入进度
			progress := types.ProgressResponse{
				Status:    fmt.Sprintf("pulling %s", fileData.Name),
				Digest:    fileData.Digest,
				Total:     fileData.Size,
				Completed: partSize,
			}
			if dataBytes, err := json.Marshal(progress); err == nil {
				a.DataCh <- dataBytes
			}
		case err := <-reqErrCh:
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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
	grpcClient, err := client.NewGRPCClient(OpenvinoGRPCHost)
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
	%s --port 9000 --grpc_bind_address 127.0.0.1 --config_path %s\\config.json
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
		downloadUrl = OVMSWindowsDownloadURL
		enginePath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino")
	case "linux":
		// todo 这里需要区分 centos 和 ubuntu(22/24) 的版本 后续实现
		execFile = "ovms"
		execPath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino/ovms")
		downloadUrl = ""
		enginePath = fmt.Sprintf("%s/%s", AOGDir, "engine/openvino")
	case "darwin":
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
		Host:           OpenvinoHTTPHost,
		Origin:         "127.0.0.1",
		Scheme:         types.ProtocolHTTP,
		RecommendModel: OpenvinoDefaultModel,
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
		Version: OpenvinoVersion,
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

	// 新建 config.json 空 文件
	configFile := fmt.Sprintf("%s/config.json", modelDir)
	_, err := os.Create(configFile)
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
	scriptZipUrl := ScriptsDownloadURL
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
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray
	dataCh := make(chan []byte)
	errCh := make(chan error)
	defer close(dataCh)
	defer close(errCh)
	localModelPath := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, req.Model)
	if _, err := os.Stat(localModelPath); err != nil {
		_ = os.MkdirAll(localModelPath, 0o755)
	}
	fileReq := &ModelScopeFileReqData{
		ModelName: req.Model,
		Revision:  ModelScopeRevision,
		Recursive: "True",
	}
	modelFiles, err := GetModelFiles(ctx, fileReq)
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
	go AsyncDownloadModelFile(ctx, *AsyncDownloadFuncParams, o)

	return newDataCh, newErrorCh
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

			// To ensure the successful unloading of the model from memory， wait 5 seconds.
			time.Sleep(5 * time.Second)

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

func (o *OpenvinoProvider) addModelToConfig(modelName, modelType string) error {
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
		// BasePath:  o.EngineConfig.EnginePath + "/models",
		GraphPath: "graph.pbtxt",
	}
	config.MediapipeConfigList = append(config.MediapipeConfigList, newModel)

	return o.saveConfig(config)
}

func (o *OpenvinoProvider) generateGraphPBTxt(modelName, modelType string) error {
	modelDir := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, modelName)
	if err := os.MkdirAll(modelDir, 0o750); err != nil {
		slog.Error("Failed to create model directory", "error", err)
		logger.EngineLogger.Error("[OpenVINO] Failed to create model directory: " + err.Error())
		return err
	}

	enginePath := strings.Replace(o.EngineConfig.EnginePath, "\\", "/", -1)

	var template string
	switch modelType {
	case types.ServiceTextToImage:
		template = fmt.Sprintf(GraphPBTxtTextToImage, modelName, enginePath)
	case types.ServiceSpeechToText:
		template = fmt.Sprintf(GraphPBTxtSpeechToText, modelName, enginePath)
	case types.ServiceSpeechToTextWS:
		template = fmt.Sprintf(GraphPBTxtSpeechToText, modelName, enginePath)
	default:
		slog.Error("Unsupported model type: " + modelType)
		logger.EngineLogger.Error("[OpenVINO] Unsupported model type: " + modelType)
		return fmt.Errorf("unsupported model type: %s", modelType)
	}

	graphPath := fmt.Sprintf("%s/graph.pbtxt", modelDir)
	return os.WriteFile(graphPath, []byte(template), 0o644)
}

func (o *OpenvinoProvider) PullModel(ctx context.Context, req *types.PullModelRequest, fn types.PullProgressFunc) (*types.ProgressResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	modelArray := append(client.ModelClientMap[req.Model], cancel)
	client.ModelClientMap[req.Model] = modelArray
	localModelPath := fmt.Sprintf("%s/models/%s", o.EngineConfig.EnginePath, req.Model)
	if _, err := os.Stat(localModelPath); err != nil {
		_ = os.MkdirAll(localModelPath, 0o755)
	}

	fileReq := &ModelScopeFileReqData{
		ModelName: req.Model,
		Revision:  ModelScopeRevision,
		Recursive: "True",
	}
	modelFiles, err := GetModelFiles(ctx, fileReq)
	if err != nil {
		return nil, err
	}

	logger.EngineLogger.Debug("[OpenVINO] modelFiles: " + fmt.Sprintf("%+v", modelFiles))

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
	go AsyncDownloadModelFile(ctx, *AsyncDownloadFuncParams, o)

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

const (
	GraphPBTxtSpeechToText = `input_stream: "OVMS_PY_TENSOR:audio"
input_stream: "OVMS_PY_TENSOR_PARAMS:params"
output_stream: "OVMS_PY_TENSOR:result"

node {
  name: "%s"
  calculator: "PythonExecutorCalculator"
  input_side_packet: "PYTHON_NODE_RESOURCES:py"

  input_stream: "INPUT:audio"
  input_stream: "PARAMS:params"
  output_stream: "OUTPUT:result"
  node_options: {
    [type.googleapis.com/mediapipe.PythonExecutorCalculatorOptions]: {
      handler_path: "%s/scripts/speech-to-text/whisper.py"
    }
  }
}`

	GraphPBTxtTextToImage = `input_stream: "OVMS_PY_TENSOR:prompt"
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
      handler_path: "%s/scripts/text-to-image/stable_diffusion.py"
    }
  }
}`
)
