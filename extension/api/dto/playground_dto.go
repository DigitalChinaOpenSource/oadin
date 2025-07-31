package dto

import (
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServerConfig struct {
	Id             string            `json:"id"`
	Name           string            `json:"name"`
	Logo           string            `json:"logo"` // MCP图标
	Args           []string          `json:"args"`
	Command        string            `json:"command"`
	Env            map[string]string `json:"env"` // 使用map存储动态环境变量
	Tools          []mcp.Tool        `json:"tools"`
	StartCacheTime time.Time         // 缓存时间
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

type ClientRunToolResponse struct {
	Content  any    `json:"content"` // Can be TextContent, ImageContent, AudioContent, or EmbeddedResource
	IsError  bool   `json:"isError"`
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
