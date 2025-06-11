package dto

import (
	"byze/internal/utils/bcode"
)

// 流式消息响应
type StreamMessageResponse struct {
	Bcode *bcode.Bcode `json:"bcode"`
	Data  MessageChunk `json:"data"`
}

// 消息块
type MessageChunk struct {
	ID         string `json:"id"`
	SessionID  string `json:"session_id"`
	Content    string `json:"content"`
	IsComplete bool   `json:"is_complete"`
	Thoughts   string `json:"thoughts,omitempty"`
	Type       string `json:"type,omitempty"`       // "answer"、"thoughts"等
	ModelID    string `json:"model_id,omitempty"`   // 新增字段
	ModelName  string `json:"model_name,omitempty"` // 新增字段
}

// 发送流式消息请求
type SendStreamMessageRequest struct {
	SessionID string `json:"session_id" binding:"required"`
	Content   string `json:"content" binding:"required"`
}
