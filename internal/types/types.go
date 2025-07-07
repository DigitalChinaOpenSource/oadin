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
	"container/list"
	"fmt"
	"net/http"
	"time"

	"intel.com/aog/internal/constants"
)

const (
	ServiceSourceLocal  = "local"
	ServiceSourceRemote = "remote"

	FlavorAOG      = "aog"
	FlavorTencent  = "tencent"
	FlavorDeepSeek = "deepseek"
	FlavorOpenAI   = "openai"
	FlavorOllama   = "ollama"
	FlavorBaidu    = "baidu"
	FlavorAliYun   = "aliyun"
	FlavorOpenvino = "openvino"

	AuthTypeNone   = "none"
	AuthTypeApiKey = "apikey"
	AuthTypeToken  = "token"
	AuthTypeSign   = "sign"

	ServiceChat           = "chat"
	ServiceModels         = "models"
	ServiceGenerate       = "generate"
	ServiceEmbed          = "embed"
	ServiceTextToImage    = "text-to-image"
	ServiceTextToSpeech   = "text-to-speech"
	ServiceSpeechToText   = "speech-to-text"
	ServiceTextToVideo    = "text-to-video"
	ServiceSpeechToTextWS = "speech-to-text-ws"

	ServiceChatAvatar         = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/chat.svg"
	ServiceTextToImageAvatar  = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/text-to-image.svg"
	ServiceEmbedAvatar        = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/Embed.svg"
	ServiceGenerateAvatar     = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/Generate.svg"
	ServiceSpeechToTextAvatar = constants.BaseDownloadURL + constants.UrlDirPathWindows + "/Speech-to-text.svg"

	ImageTypeUrl    = "url"
	ImageTypePath   = "path"
	ImageTypeBase64 = "base64"

	HybridPolicyDefault = "default"
	HybridPolicyLocal   = "always_local"
	HybridPolicyRemote  = "always_remote"

	ProtocolHTTP        = "HTTP"
	ProtocolHTTPS       = "HTTPS"
	ProtocolGRPC        = "GRPC"
	ProtocolGRPC_STREAM = "GRPC_STREAM"

	ExposeProtocolHTTP      = "HTTP"
	ExposeProtocolWEBSOCKET = "WEBSOCKET"

	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	EngineStartModeDaemon   = "daemon"
	EngineStartModeStandard = "standard"

	VersionRecordStatusInstalled = 1
	VersionRecordStatusUpdated   = 2

	WSTaskTypeAudio = "audio"

	AudioWav  = "wav"
	AudioMp3  = "mp3"
	AudioM4a  = "m4a"
	AudioOgg  = "ogg"
	AudioFlac = "flac"
	AudioAac  = "aac"
	AudioMp4  = "mp4"

	GPUTypeNvidia    = "Nvidia"
	GPUTypeAmd       = "AMD"
	GPUTypeIntelArc  = "Intel Arc"
	GPUTypeIntelCore = "Intel Core"
	GPUTypeNone      = "None"
)

var (
	SupportService      = []string{ServiceEmbed, ServiceModels, ServiceChat, ServiceGenerate, ServiceTextToImage, ServiceSpeechToText, ServiceSpeechToTextWS}
	SupportHybridPolicy = []string{HybridPolicyDefault, HybridPolicyLocal, HybridPolicyRemote}
	SupportAuthType     = []string{AuthTypeNone, AuthTypeApiKey, AuthTypeSign, AuthTypeToken}
	SupportFlavor       = []string{FlavorDeepSeek, FlavorOpenAI, FlavorTencent, FlavorOllama, FlavorBaidu, FlavorAliYun, FlavorOpenvino}
	SupportModelEngine  = []string{FlavorOpenvino, FlavorOllama}
	SupportImageType    = []string{ImageTypeUrl, ImageTypeBase64, ImageTypePath}
	SupportAudioType    = []string{AudioWav, AudioMp3}
)

type HTTPContent struct {
	Body   []byte
	Header http.Header
}

func (hc HTTPContent) String() string {
	return fmt.Sprintf("HTTPContent{Header: %+v, Body: %s}", hc.Header, string(hc.Body))
}

type HTTPErrorResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (hc *HTTPErrorResponse) Error() string {
	return fmt.Sprintf("HTTPErrorResponse{StatusCode: %d, Header: %+v, Body: %s}", hc.StatusCode, hc.Header, string(hc.Body))
}

// ConversionStepDef NOTE: we use YAML instead of JSON here because it's easier to read and write
// In particular, it supports multiline strings which greatly help write
// jsonata templates
type ConversionStepDef struct {
	Converter string `yaml:"converter"`
	Config    any    `yaml:"config"`
}

type ScheduleDetails struct {
	Id           uint64
	IsRunning    bool
	ListMark     *list.Element
	TimeEnqueue  time.Time
	TimeRun      time.Time
	TimeComplete time.Time
}

type DropAction struct{}

func (d *DropAction) Error() string {
	return "Need to drop this content"
}

type ServiceProviderProperties struct {
	MaxInputTokens        int      `json:"max_input_tokens"`
	SupportedResponseMode []string `json:"supported_response_mode"`
	ModeIsChangeable      bool     `json:"mode_is_changeable"`
	Models                []string `json:"models"`
	XPU                   []string `json:"xpu"`
}

type RecommendConfig struct {
	ModelEngine       string `json:"model_engine"`
	ModelName         string `json:"model_name"`
	EngineDownloadUrl string `json:"engine_download_url"`
}

// ListResponse is the response from [Client.List].
type ListResponse struct {
	Models []ListModelResponse `json:"models"`
}

// ListModelResponse is a single model description in [ListResponse].
type ListModelResponse struct {
	Name       string       `json:"name"`
	Model      string       `json:"model"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details,omitempty"`
}

type EngineVersionResponse struct {
	Version string `json:"version"`
}

// ModelDetails provides details about a model.
type ModelDetails struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// PullModelRequest is the request passed to [Client.Pull].
type PullModelRequest struct {
	Model     string `json:"model"`
	Insecure  bool   `json:"insecure,omitempty"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Stream    *bool  `json:"stream,omitempty"`
	ModelType string `json:"model_type,omitempty"`

	// Deprecated: set the model name with Model instead
	Name string `json:"name"`
}

// DeleteRequest is the request passed to [Client.Delete].
type DeleteRequest struct {
	Model string `json:"model"`
}

// [PullProgressFunc] and [PushProgressFunc].
type ProgressResponse struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type PullProgressFunc func(ProgressResponse) error

type EngineRecommendConfig struct {
	Host           string `json:"host"`
	Origin         string `json:"origin"`
	Scheme         string `json:"scheme"`
	RecommendModel string `json:"recommend_model"`
	DownloadUrl    string `json:"download_url"`
	DownloadPath   string `json:"download_path"`
	EnginePath     string `json:"engine_path"`
	ExecPath       string `json:"exec_path"`
	ExecFile       string `json:"exec_file"`
}

type PathDiskSizeInfo struct {
	FreeSize  int `json:"free_size"`
	TotalSize int `json:"total_size"`
	UsageSize int `json:"usage_size"`
}
