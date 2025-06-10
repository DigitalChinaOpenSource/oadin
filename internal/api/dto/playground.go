package dto

import (
	"byze/internal/utils/bcode"
)

// 创建会话请求
type CreateSessionRequest struct {
	Title           string `json:"title"`
	ModelId         string `json:"modelId"`
	EmbedModelId    string `json:"embedModelId"`
	ThinkingEnabled bool   `json:"thinkingEnabled"`
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
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

type GetSessionsResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Session    `json:"data"`
}

type SendMessageRequest struct {
	SessionId string `json:"sessionId"`
	Content   string `json:"content"`
}

// 发送消息响应
type SendMessageResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  []Message    `json:"data"`
}

// 消息信息
type Message struct {
	Id        string `json:"id"`
	SessionId string `json:"sessionId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	Thoughts  string `json:"thoughts,omitempty"`
	Type      string `json:"type,omitempty"`
	ModelId   string `json:"modelId,omitempty"`
	ModelName string `json:"modelName,omitempty"`
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
// 用于切换指定会话的chat模型
// sessionId: 会话ID, modelId: 新模型ID
// embedModelId/thinkingEnabled 可选
type ChangeSessionModelRequest struct {
	SessionId       string `json:"sessionId"`
	ModelId         string `json:"modelId"`
	EmbedModelId    string `json:"embedModelId,omitempty"`
	ThinkingEnabled *bool  `json:"thinkingEnabled,omitempty"`
}

type ChangeSessionModelResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  *Session     `json:"data"`
}
