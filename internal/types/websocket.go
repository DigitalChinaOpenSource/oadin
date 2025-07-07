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

package types

import (
	"encoding/json"
)

// WebSocket消息常量
const (
	// Client Action types
	WSActionRunTask    = "run-task"
	WSActionFinishTask = "finish-task"

	WSSTTTaskTypeUnknown    = "unknown"
	WSSTTTaskTypeRunTask    = WSActionRunTask    // 启动识别任务
	WSSTTTaskTypeAudio      = "audio"            // 音频数据
	WSSTTTaskTypeFinishTask = WSActionFinishTask // 结束识别任务

	// 服务端Event类型
	WSEventTaskStarted     = "task-started"
	WSEventTaskFinished    = "task-finished"
	WSEventResultGenerated = "result-generated"
	WSEventTaskFailed      = "task-failed"

	// 错误码
	WSErrorCodeClientError = "CLIENT_ERROR"
	WSErrorCodeServerError = "SERVER_ERROR"
	WSErrorCodeModelError  = "MODEL_ERROR"
)

// WebSocketParameters 通用参数结构
type WebSocketParameters struct {
	Service      string `json:"service,omitempty"`       // 服务名称，如: "speech-to-text" "speech-to-text-ws"
	Format       string `json:"format,omitempty"`        // 音频格式: pcm/wav/mp3
	SampleRate   int    `json:"sample_rate,omitempty"`   // 采样率，通常为16000
	Language     string `json:"language,omitempty"`      // 语言，如"zh"、"en"
	UseVAD       bool   `json:"use_vad,omitempty"`       // 是否使用VAD
	ReturnFormat string `json:"return_format,omitempty"` // 返回格式，如"text"、"json"、"srt"
}

// SpeechToTextParams 语音识别详细参数结构
type SpeechToTextParams struct {
	// 基本参数
	TaskID      string `json:"task_id,omitempty"` // 任务ID
	Service     string `json:"service,omitempty"`
	Model       string `json:"model,omitempty"`        // 模型名称
	Language    string `json:"language,omitempty"`     // 语言，如"zh"、"en"
	AudioFormat string `json:"audio_format,omitempty"` // 音频格式
	SampleRate  int    `json:"sample_rate,omitempty"`  // 采样率

	// 高级参数
	UseVAD       bool   `json:"use_vad,omitempty"`       // 是否使用VAD
	ReturnFormat string `json:"return_format,omitempty"` // 返回格式
	MaxDuration  int    `json:"max_duration,omitempty"`  // 最大处理时长(秒)

	// 状态参数
	TaskStarted     bool  `json:"task_started,omitempty"`      // 任务是否已启动
	StartTime       int64 `json:"start_time,omitempty"`        // 任务开始时间
	EndTime         int64 `json:"end_time,omitempty"`          // 任务结束时间
	TotalAudioBytes int   `json:"total_audio_bytes,omitempty"` // 总音频字节数
	LastAudioTime   int64 `json:"last_audio_time,omitempty"`   // 最后一次接收音频的时间
}

// NewSpeechToTextParams 创建语音识别参数对象，设置默认值
func NewSpeechToTextParams() *SpeechToTextParams {
	return &SpeechToTextParams{
		Language:     "zh",
		AudioFormat:  "wav",
		SampleRate:   16000,
		UseVAD:       true,
		ReturnFormat: "text",
	}
}

// ToJSON 将参数转换为JSON字符串
func (p *SpeechToTextParams) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// FromJSON 从JSON字符串解析参数
func (p *SpeechToTextParams) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// ============================
// 客户端 => 服务端 Action消息
// ============================

// WebSocketActionMessage 客户端发送到服务端的基础消息结构
type WebSocketActionMessage struct {
	Task       string               `json:"task"`                 // 任务类型，如"speech-to-text-ws"
	Action     string               `json:"action"`               // 动作类型，如"run-task"或"finish-task"
	TaskID     string               `json:"task_id,omitempty"`    // 任务ID，由服务端生成并返回，用于后续消息关联
	Model      string               `json:"model,omitempty"`      // 模型名称
	Parameters *WebSocketParameters `json:"parameters,omitempty"` // 可选参数
}

// WebSocketRunTaskAction 开始任务的消息
type WebSocketRunTaskAction struct {
	WebSocketActionMessage
	// 可以在这里添加RunTask特有的字段
}

// WebSocketFinishTaskAction 结束任务的消息
type WebSocketFinishTaskAction struct {
	WebSocketActionMessage
	// 可以在这里添加FinishTask特有的字段
}

// ============================
// 服务端 => 客户端 Event消息
// ============================

// WebSocketEventHeader 所有服务端事件的通用头部
type WebSocketEventHeader struct {
	TaskID       string `json:"task_id"`                 // 任务ID
	Event        string `json:"event"`                   // 事件类型
	ErrorCode    string `json:"error_code,omitempty"`    // 错误码，仅在task-failed事件中出现
	ErrorMessage string `json:"error_message,omitempty"` // 错误信息，仅在task-failed事件中出现
}

// WebSocketEventMessage 服务端发送到客户端的基础消息结构
type WebSocketEventMessage struct {
	Header  WebSocketEventHeader `json:"header"`            // 事件头部
	Payload interface{}          `json:"payload,omitempty"` // 根据事件类型不同，包含不同的数据
}

// WebSocketSentence 识别结果中的句子结构
type WebSocketSentence struct {
	BeginTime *int   `json:"begin_time"` // 开始时间，毫秒
	EndTime   *int   `json:"end_time"`   // 结束时间，毫秒，可能为null
	Text      string `json:"text"`       // 识别的文本
}

// WebSocketRecognitionOutput 语音识别输出结构
type WebSocketRecognitionOutput struct {
	Sentence WebSocketSentence `json:"sentence"` // 识别的句子
}

// WebSocketResultPayload 识别结果的有效载荷
type WebSocketResultPayload struct {
	Output WebSocketRecognitionOutput `json:"output"` // 识别输出
}

// WebSocketResultEvent 识别结果事件
type WebSocketResultEvent struct {
	Header  WebSocketEventHeader   `json:"header"`
	Payload WebSocketResultPayload `json:"payload"`
}

// ============================
// 辅助函数
// ============================

// NewRunTaskAction 创建开始任务的消息
func NewRunTaskAction(model string, parameters *WebSocketParameters) WebSocketRunTaskAction {
	return WebSocketRunTaskAction{
		WebSocketActionMessage: WebSocketActionMessage{
			Task:       ServiceSpeechToTextWS,
			Action:     WSActionRunTask,
			Model:      model,
			Parameters: parameters,
		},
	}
}

// NewFinishTaskAction 创建结束任务的消息
func NewFinishTaskAction(taskID, model string) WebSocketFinishTaskAction {
	return WebSocketFinishTaskAction{
		WebSocketActionMessage: WebSocketActionMessage{
			Task:   ServiceSpeechToTextWS,
			Action: WSActionFinishTask,
			TaskID: taskID,
			Model:  model,
		},
	}
}

// NewTaskStartedEvent 创建任务开始事件
func NewTaskStartedEvent(taskID string) WebSocketEventMessage {
	return WebSocketEventMessage{
		Header: WebSocketEventHeader{
			TaskID: taskID,
			Event:  WSEventTaskStarted,
		},
		Payload: map[string]interface{}{},
	}
}

// NewTaskFinishedEvent 创建任务完成事件
func NewTaskFinishedEvent(taskID string) WebSocketEventMessage {
	return WebSocketEventMessage{
		Header: WebSocketEventHeader{
			TaskID: taskID,
			Event:  WSEventTaskFinished,
		},
		Payload: map[string]interface{}{},
	}
}

// NewTaskFailedEvent 创建任务失败事件
func NewTaskFailedEvent(taskID, errorCode, errorMessage string) WebSocketEventMessage {
	return WebSocketEventMessage{
		Header: WebSocketEventHeader{
			TaskID:       taskID,
			Event:        WSEventTaskFailed,
			ErrorCode:    errorCode,
			ErrorMessage: errorMessage,
		},
		Payload: map[string]interface{}{},
	}
}

// NewResultGeneratedEvent 创建识别结果事件
func NewResultGeneratedEvent(taskID string, beginTime *int, endTime *int, text string) WebSocketResultEvent {
	return WebSocketResultEvent{
		Header: WebSocketEventHeader{
			TaskID: taskID,
			Event:  WSEventResultGenerated,
		},
		Payload: WebSocketResultPayload{
			Output: WebSocketRecognitionOutput{
				Sentence: WebSocketSentence{
					BeginTime: beginTime,
					EndTime:   endTime,
					Text:      text,
				},
			},
		},
	}
}
