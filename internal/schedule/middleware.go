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

package schedule

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"intel.com/aog/internal/client"
	"intel.com/aog/internal/logger"
	"intel.com/aog/internal/types"
	"intel.com/aog/internal/utils"
)

type BaseMiddleware struct {
	ServiceName string
}

// GetMiddleware Returns the corresponding middleware chain based on the service type
func GetMiddleware(serviceName string) []TaskMiddleware {
	switch serviceName {
	case types.ServiceTextToImage:
		return []TaskMiddleware{NewTextToImageMiddleware()}
	case types.ServiceSpeechToText:
		return []TaskMiddleware{NewSpeechToTextMiddleware()}
	case types.ServiceSpeechToTextWS:
		return []TaskMiddleware{NewSpeechToTextWSMiddleware()}
	case types.ServiceChat:
		return []TaskMiddleware{NewChatMiddleware()}
	default:
		return []TaskMiddleware{}
	}
}

// TextToImageMiddleware Handles image-related requests
type TextToImageMiddleware struct {
	BaseMiddleware
}

func (m *TextToImageMiddleware) Handle(st *ServiceTask) error {
	if st.Request.Service != m.ServiceName {
		return nil
	}

	body, err := utils.ParseRequestBody(st.Request.HTTP.Body)
	if err != nil {
		return fmt.Errorf("parse request body: %w", err)
	}

	// Skip if no image-related fields
	if !m.hasImageFields(body) {
		return nil
	}

	if err := m.validateImageParams(body); err != nil {
		return err
	}

	if err := m.processImage(body, st.Target.Location); err != nil {
		return err
	}

	newReqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal new request body: %w", err)
	}
	st.Request.HTTP.Body = newReqBody
	return nil
}

func (m *TextToImageMiddleware) hasImageFields(body map[string]interface{}) bool {
	_, typeExists := body["image_type"]
	_, imageExists := body["image"]
	return typeExists || imageExists
}

func (m *TextToImageMiddleware) validateImageParams(body map[string]interface{}) error {
	imageType, typeOk := body["image_type"].(string)
	_, imageOk := body["image"].(string)

	switch {
	case !typeOk && imageOk:
		return errors.New("image_type request param lost")
	case typeOk && !imageOk:
		return errors.New("image request param lost")
	case !utils.Contains(types.SupportImageType, imageType):
		return errors.New("unsupported image type")
	}
	return nil
}

func (m *TextToImageMiddleware) processImage(body map[string]interface{}, location string) error {
	imageType := body["image_type"].(string)
	image := body["image"].(string)

	switch {
	case imageType == types.ImageTypePath && location == types.ServiceSourceRemote:
		return m.handleLocalToRemote(body, image)
	case imageType == types.ImageTypeUrl && location == types.ServiceSourceLocal:
		return m.handleRemoteToLocal(body, image)
	}
	return nil
}

func (m *TextToImageMiddleware) handleLocalToRemote(body map[string]interface{}, imagePath string) error {
	imgData, err := os.ReadFile(imagePath)
	if err != nil {
		logger.LogicLogger.Error("[Middleware] Failed to read image file", "error", err)
		return fmt.Errorf("read image file: %w", err)
	}
	body["image"] = base64.StdEncoding.EncodeToString(imgData)
	return nil
}

func (m *TextToImageMiddleware) handleRemoteToLocal(body map[string]interface{}, imageUrl string) error {
	downLoadPath, err := utils.GetDownloadDir()
	if err != nil {
		return fmt.Errorf("get download directory: %w", err)
	}
	savePath, err := utils.DownloadFile(imageUrl, downLoadPath)
	if err != nil {
		return fmt.Errorf("download file: %w", err)
	}
	body["image"] = savePath
	return nil
}

func NewTextToImageMiddleware() *TextToImageMiddleware {
	return &TextToImageMiddleware{
		BaseMiddleware: BaseMiddleware{
			ServiceName: types.ServiceTextToImage,
		},
	}
}

// SpeechToTextMiddleware Handles SpeechToText-related requests
type SpeechToTextMiddleware struct {
	BaseMiddleware
}

func (m *SpeechToTextMiddleware) Handle(st *ServiceTask) error {
	if st.Request.Service != m.ServiceName {
		return nil
	}

	// TODO: Implement audio processing logic
	//
	// 1. Audio format conversion
	// 2. Sample rate adjustment
	// 3. Audio encoding conversion, etc.
	logger.LogicLogger.Debug("[Middleware] Processing audio request")
	return nil
}

func NewSpeechToTextMiddleware() *SpeechToTextMiddleware {
	return &SpeechToTextMiddleware{
		BaseMiddleware: BaseMiddleware{
			ServiceName: types.ServiceSpeechToText,
		},
	}
}

// SpeechToTextWSMiddleware 处理WebSocket的语音识别请求
type SpeechToTextWSMiddleware struct {
	BaseMiddleware
}

func (m *SpeechToTextWSMiddleware) Handle(st *ServiceTask) error {
	// 检查是否是目标服务类型
	if st.Request.Service != m.ServiceName {
		return nil
	}

	// 获取WebSocket连接ID
	connID := st.Request.WebSocketConnID
	if connID == "" {
		return fmt.Errorf("missing WebSocket connection ID")
	}

	logger.LogicLogger.Debug("[SpeechToTextWS] Processing WebSocket request",
		"connID", connID,
		"taskID", st.Schedule.Id)

	// 从WebSocketManager获取连接 - 添加更多调试信息
	wsManager := client.GetGlobalWebSocketManager()
	logger.LogicLogger.Debug("[SpeechToTextWS] Attempting to get WebSocket connection",
		"connID", connID,
		"managerInstanceAddr", fmt.Sprintf("%p", wsManager),
		"connectionsCount", wsManager.GetActiveConnectionCount())

	wsConn, exists := wsManager.GetConnection(connID)
	if !exists {
		logger.LogicLogger.Error("[SpeechToTextWS] WebSocket connection not found",
			"connID", connID,
			"managerInstanceAddr", fmt.Sprintf("%p", wsManager),
			"connectionsCount", wsManager.GetActiveConnectionCount())
		return fmt.Errorf("WebSocket connection not found: %s", connID)
	}

	// 获取并记录STT参数
	sttParams := wsConn.GetSTTParams()
	logger.LogicLogger.Debug("[SpeechToTextWS] Retrieved connection parameters",
		"connID", connID,
		"format", sttParams.AudioFormat,
		"sampleRate", sttParams.SampleRate,
		"language", sttParams.Language,
		"totalBytes", sttParams.TotalAudioBytes,
		"taskStarted", wsConn.IsTaskStarted(st.Schedule.Id))

	// 获取任务类型
	taskType := wsConn.GetTaskType(st.Schedule.Id)
	logger.LogicLogger.Debug("[SpeechToTextWS] Task type",
		"connID", connID,
		"taskType", taskType)

	// 根据任务类型执行不同的参数校验和处理
	switch taskType {
	case types.WSSTTTaskTypeAudio:
		return m.validateAudioData(st, wsConn)
	case types.WSSTTTaskTypeRunTask:
		return m.validateRunTask(st, wsConn)
	case types.WSSTTTaskTypeFinishTask:
		return m.validateFinishTask(st, wsConn)
	default:
		logger.LogicLogger.Warn("[SpeechToTextWS] Unknown task type",
			"taskType", taskType,
			"connID", wsConn.ID)
		return fmt.Errorf("unknown task type: %v", taskType)
	}
}

// 验证和处理启动任务命令
func (m *SpeechToTextWSMiddleware) validateRunTask(st *ServiceTask, session *client.WebSocketConnection) error {
	sttParams := session.GetSTTParams()

	logger.LogicLogger.Debug("[SpeechToTextWS] Validating run-task command",
		"connID", session.ID,
		"language", sttParams.Language,
		"sampleRate", sttParams.SampleRate)

	// 验证语言参数
	if sttParams.Language == "" {
		sttParams.Language = "<|zh|>" // 默认使用中文
		logger.LogicLogger.Info("[SpeechToTextWS] Using default language",
			"language", sttParams.Language,
			"connID", session.ID)
	}

	// 验证采样率参数
	validSampleRates := []int{8000, 16000, 22050, 44100, 48000}
	validRate := false
	for _, rate := range validSampleRates {
		if sttParams.SampleRate == rate {
			validRate = true
			break
		}
	}

	if !validRate {
		logger.LogicLogger.Warn("[SpeechToTextWS] Invalid sample rate, using default",
			"provided", sttParams.SampleRate,
			"default", 16000,
			"connID", session.ID)
		sttParams.SampleRate = 16000
	}

	// 验证返回格式
	validFormats := []string{"text", "json", "srt", "vtt"}
	if !utils.Contains(validFormats, sttParams.ReturnFormat) {
		logger.LogicLogger.Warn("[SpeechToTextWS] Invalid return format, using default",
			"provided", sttParams.ReturnFormat,
			"default", "text",
			"connID", session.ID)
		sttParams.ReturnFormat = "text"
	}

	logger.LogicLogger.Info("[SpeechToTextWS] Task validated",
		"connID", session.ID)

	return nil
}

// 验证和处理音频数据
func (m *SpeechToTextWSMiddleware) validateAudioData(st *ServiceTask, session *client.WebSocketConnection) error {
	sttParams := session.GetSTTParams()

	// 检查连接ID是否有效
	if session.ID == "" {
		return fmt.Errorf("invalid WebSocket connection ID")
	}

	// 检查音频数据大小
	dataSize := len(st.Request.HTTP.Body)
	if dataSize == 0 {
		return fmt.Errorf("empty audio data received")
	}

	if dataSize > 10*1024*1024 { // 10MB限制
		return fmt.Errorf("audio data too large: %d bytes (max 10MB)", dataSize)
	}

	// 更新会话中的音频处理状态
	sttParams.TotalAudioBytes += dataSize
	sttParams.LastAudioTime = time.Now().Unix()

	logger.LogicLogger.Debug("[SpeechToTextWS] Processed audio data chunk",
		"connID", session.ID,
		"size", dataSize,
		"totalBytes", sttParams.TotalAudioBytes)

	return nil
}

// 验证和处理结束任务命令
func (m *SpeechToTextWSMiddleware) validateFinishTask(st *ServiceTask, session *client.WebSocketConnection) error {
	// 检查连接ID是否存在
	if session.ID == "" {
		return fmt.Errorf("invalid WebSocket connection ID")
	}

	// 任务完成状态日志记录
	startTime, endTime := session.GetTaskTimes(st.Schedule.Id)
	duration := endTime - startTime
	sttParams := session.GetSTTParams()

	logger.LogicLogger.Info("[SpeechToTextWS] Task validated for finish",
		"connID", session.ID,
		"duration", duration,
		"totalBytes", sttParams.TotalAudioBytes)

	return nil
}

func NewSpeechToTextWSMiddleware() *SpeechToTextWSMiddleware {
	return &SpeechToTextWSMiddleware{
		BaseMiddleware: BaseMiddleware{
			ServiceName: types.ServiceSpeechToTextWS,
		},
	}
}

// ChatMiddleware Handles chat-related requests
type ChatMiddleware struct {
	BaseMiddleware
}

func (m *ChatMiddleware) Handle(st *ServiceTask) error {
	if st.Request.Service != m.ServiceName {
		return nil
	}

	// TODO: Implement chat processing logic
	// 1. Message format validation
	// 2. Sensitive word filtering
	// 3. Message length limitation, etc.
	logger.LogicLogger.Debug("[Middleware] Processing chat request")
	return nil
}

func NewChatMiddleware() *ChatMiddleware {
	return &ChatMiddleware{
		BaseMiddleware: BaseMiddleware{
			ServiceName: types.ServiceChat,
		},
	}
}

// ExecuteMiddleware Executes the middleware chain
func ExecuteMiddleware(st *ServiceTask) error {
	middlewares := GetMiddleware(st.Request.Service)
	for _, m := range middlewares {
		if err := m.Handle(st); err != nil {
			return fmt.Errorf("middleware execution failed: %w", err)
		}
	}
	return nil
}
