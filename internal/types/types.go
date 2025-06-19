package types

import (
	"container/list"
	"fmt"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	GPUTypeNvidia    = "Nvidia"
	GPUTypeAmd       = "AMD"
	GPUTypeIntelArc  = "Intel Arc"
	GPUTypeIntelCore = "Intel Core"
	GPUTypeNone      = "None"
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
	Model    string `json:"model"`
	Insecure bool   `json:"insecure,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
	Stream   *bool  `json:"stream,omitempty"`

	// Deprecated: set the model name with Model instead
	Name string `json:"name"`
}

type CopyModelRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
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
	EnginePath     string `json:"engine_path"`
	DownloadUrl    string `json:"download_url"`
	DownloadPath   string `json:"download_path"`
	ExecPath       string `json:"exec_path"`
	ExecFile       string `json:"exec_file"`
}

type MCPServerConfig struct {
	Id      string            `json:"id"`
	Name    string            `json:"name"`
	Logo    string            `json:"logo"` // MCP图标
	Args    []string          `json:"args"`
	Command string            `json:"command"`
	Env     map[string]string `json:"env"` // 使用map存储动态环境变量
	Tools   []mcp.Tool        `json:"tools"`
}

type ClientMcpStartRequest struct {
	Ids []string `json:"ids"`
}

type ClientMcpStartResponse struct {
	Id string `json:"id"`
}

type ClientMcpStopRequest struct {
	Ids []string `json:"ids"`
}

type ClientRunToolRequest struct {
	McpId    string         `json:"mcpId"` // 工具调用的MCP ID
	ToolName string         `json:"toolName"`
	ToolArgs map[string]any `json:"toolArgs"`

	MessageId string `json:"messageId"` // 关联的消息ID
}

type ClientRunToolResponset struct {
	*mcp.CallToolResult
	Logo     string `json:"logo"`     // 工具调用的MCP图标
	ToolDesc string `json:"toolDesc"` // 工具描述
}

type ClientGetToolsRequest struct {
	Ids []string `json:"ids"`
}

type ClientGetToolsResponse struct {
	Tools []McpTool `json:"mcpTools"`
}

type McpTool struct {
	McpId string `json:"mcpId"`
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function TypeFunction `json:"function"`
}
type TypeFunction struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  mcp.ToolInputSchema `json:"parameters"`
}
