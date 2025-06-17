package dto

import (
	"byze/internal/utils/bcode"
)

// 切换会话深度思考模式请求
type ToggleThinkingRequest struct {
	SessionId string `json:"sessionId"`
	Enabled   *bool  `json:"enabled"`
}

// 切换会话深度思考模式响应
type ToggleThinkingResponse struct {
	Bcode          *bcode.Bcode `json:"bcode"`
	ThinkingActive bool         `json:"thinking_active"`
}
