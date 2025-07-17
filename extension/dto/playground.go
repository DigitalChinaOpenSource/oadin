package dto

import (
	"oadin/extension/entity"
	"oadin/internal/utils/bcode"
)

// 创建会话请求
type CreateSessionRequest struct {
	Title        string `json:"title"`
	ModelId      string `json:"modelId"`
	ModelName    string `json:"modelName"`
	EmbedModelId string `json:"embedModelId"`
}

type CreateSessionResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  Session      `json:"data"`
}

type Session struct {
	Id              string `json:"id"`
	Title           string `json:"title"`
	ModelId         string `json:"modelId"`
	ModelName       string `json:"modelName"`
	EmbedModelId    string `json:"embedModelId"`
	ThinkingEnabled bool   `json:"thinkingEnabled"`
	ThinkingActive  bool   `json:"thinkingActive"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type GetSessionsResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Session    `json:"data"`
}

type SendMessageRequest struct {
	SessionId string   `json:"sessionId"`
	Content   string   `json:"content"`
	Tools     []Tool   `json:"tools,omitempty"`
	McpIds    []string `json:"mcpIds,omitempty"`
}

// 发送消息响应
type SendMessageResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Message    `json:"data"`
}

// 消息信息
type Message struct {
	Id        string               `json:"id"`
	SessionId string               `json:"sessionId"`
	Role      string               `json:"role"`
	Content   string               `json:"content"`
	CreatedAt string               `json:"createdAt"`
	Thoughts  string               `json:"thoughts,omitempty"`
	Type      string               `json:"type,omitempty"`
	ModelId   string               `json:"modelId,omitempty"`
	ModelName string               `json:"modelName,omitempty"`
	ToolCalls []entity.ToolMessage `json:"toolCalls,omitempty"`
}

// 获取会话消息请求
type GetMessagesRequest struct {
	SessionId string `json:"sessionId"`
}

// 获取会话消息响应
type GetMessagesResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Message    `json:"data"`
}

// 切换会话模型请求
type ChangeSessionModelRequest struct {
	SessionId    string `json:"sessionId"`
	ModelId      string `json:"modelId"`
	EmbedModelId string `json:"embedModelId,omitempty"`
}

// 删除会话请求
type DeleteSessionRequest struct {
	SessionId string `json:"sessionId"`
}

// 删除会话响应
type DeleteSessionResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
}

type ChangeSessionModelResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  *Session     `json:"data"`
}
