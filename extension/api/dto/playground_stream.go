package dto

import (
	"oadin/internal/utils/bcode"
)

// 流式消息响应
type StreamMessageResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  MessageChunk `json:"data"`
}

// 消息块
type MessageChunk struct {
	ID            string     `json:"id"`
	SessionID     string     `json:"session_id"`
	Content       string     `json:"content"`
	IsComplete    bool       `json:"is_complete"`
	Thoughts      string     `json:"thoughts,omitempty"`
	Type          string     `json:"type,omitempty"`           // "answer"、"thoughts"等
	ModelID       string     `json:"model_id,omitempty"`       // 新增字段
	ModelName     string     `json:"model_name,omitempty"`     // 新增字段
	ToolCalls     []ToolCall `json:"tool_calls,omitempty"`     // 新增，支持Ollama工具调用
	TotalDuration int64      `json:"total_duration,omitempty"` // 新增总耗时
	ToolGroupID   string     `json:"tool_group_id,omitempty"`  // 新增工具组ID
}

// 发送流式消息请求
type SendStreamMessageRequest struct {
	SessionID   string   `json:"SessionID" binding:"required"`
	Content     string   `json:"content"`
	ToolGroupID string   `json:"toolGroupID,omitempty"` // 关联的消息ID
	McpIds      []string `json:"mcpIds,omitempty"`
	Tools       []Tool   `json:"tools,omitempty"`
}
