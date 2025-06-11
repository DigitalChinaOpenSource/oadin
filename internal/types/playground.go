package types

import (
	"time"
)

// 对话会话
type ChatSession struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	ModelID         string    `json:"model_id"`         // 使用的chat模型
	ModelName       string    `json:"model_name"`       // 新增字段，模型名称
	EmbedModelID    string    `json:"embed_model_id"`   // 使用的embedding模型
	ThinkingEnabled bool      `json:"thinking_enabled"` // 是否启用思考模式
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Entity接口
func (s *ChatSession) SetCreateTime(t time.Time) { s.CreatedAt = t }
func (s *ChatSession) SetUpdateTime(t time.Time) { s.UpdatedAt = t }
func (s *ChatSession) PrimaryKey() string        { return s.ID }
func (s *ChatSession) TableName() string         { return "chat_sessions" }
func (s *ChatSession) Index() map[string]interface{} {
	return map[string]interface{}{
		"id": s.ID,
	}
}

// 对话消息
type ChatMessage struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Order     int       `json:"order" gorm:"column:msg_order"`
	CreatedAt time.Time `json:"created_at"`
	ModelID   string    `json:"model_id,omitempty"`   // 新增字段，模型ID
	ModelName string    `json:"model_name,omitempty"` // 新增字段，模型名称
}

func (m *ChatMessage) SetCreateTime(t time.Time) { m.CreatedAt = t }
func (m *ChatMessage) SetUpdateTime(t time.Time) {}
func (m *ChatMessage) PrimaryKey() string        { return m.ID }
func (m *ChatMessage) TableName() string         { return "chat_messages" }
func (m *ChatMessage) Index() map[string]interface{} {
	return map[string]interface{}{
		"id":         m.ID,
		"session_id": m.SessionID,
	}
}

// 聊天请求模型
type ChatRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	Temperature float32             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
	Stream      bool                `json:"stream,omitempty"`
	Options     map[string]any      `json:"options,omitempty"`
	Tools       []map[string]any    `json:"tools,omitempty"` // 新增，支持Ollama工具调用
}

// 聊天响应模型
type ChatResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Model      string `json:"model"`
	ModelName  string `json:"model_name,omitempty"` // 新增字段
	Content    string `json:"content"`
	IsComplete bool   `json:"is_complete"`        // 流式输出时，是否是最后一个块
	Thoughts   string `json:"thoughts,omitempty"` // 深度思考的结果
	Type       string `json:"type,omitempty"`     // "answer"、"thoughts"等
}
