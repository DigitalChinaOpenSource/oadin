package entity

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
	ThinkingEnabled bool      `json:"thinking_enabled"` // 模型是否支持深度思考
	ThinkingActive  bool      `json:"thinking_active"`  // 当前会话是否启用深度思考
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Entity接口
func (s *ChatSession) SetCreateTime(t time.Time) { s.CreatedAt = t }
func (s *ChatSession) SetUpdateTime(t time.Time) { s.UpdatedAt = t }
func (s *ChatSession) PrimaryKey() string        { return "id" }
func (s *ChatSession) TableName() string         { return "chat_sessions" }
func (s *ChatSession) Index() map[string]interface{} {
	// 只有ID有值时才添加到查询条件中，否则返回空索引以查询所有记录
	if s.ID != "" {
		return map[string]interface{}{
			"id": s.ID,
		}
	}
	return map[string]interface{}{}
}

// 对话消息
type ChatMessage struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	Role          string    `json:"role"`
	Content       string    `json:"content"`
	Order         int       `json:"order" gorm:"column:msg_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ModelID       string    `json:"model_id,omitempty"`         // 新增字段，模型ID
	ModelName     string    `json:"model_name,omitempty"`       // 新增字段，模型名称
	TotalDuration int64     `json:"total_duration,omitempty"`   // 总耗时，单位秒
	IsToolGroupID bool      `json:"is_tool_group_id,omitempty"` // 是否是工具组ID
}

func (m *ChatMessage) SetCreateTime(t time.Time) { m.CreatedAt = t }
func (m *ChatMessage) SetUpdateTime(t time.Time) { m.UpdatedAt = t }
func (m *ChatMessage) PrimaryKey() string        { return "id" }
func (m *ChatMessage) TableName() string         { return "chat_messages" }
func (m *ChatMessage) Index() map[string]interface{} {
	result := map[string]interface{}{}

	// 只有当ID有值时，才添加ID条件
	if m.ID != "" {
		result["id"] = m.ID
	}

	// 只有当SessionID有值时，才添加session_id条件
	if m.SessionID != "" {
		result["session_id"] = m.SessionID
	}

	return result
}

// 工具调用
type ToolMessage struct {
	ID             string    `json:"id"`
	SessionID      string    `json:"session_id"`
	MessageId      string    `json:"message_id"`       // 关联的用户消息ID
	AssistantMsgID string    `json:"assistant_msg_id"` // 关联的大模型回复消息ID
	McpId          string    `json:"mcp_id"`           // 工具调用的MCP ID
	Logo           string    `json:"mcp_image"`        // 工具调用的MCP图标
	Name           string    `json:"name"`             // 工具名称
	Desc           string    `json:"desc"`             // 工具描述
	InputParams    string    `json:"input_params"`     // 输入参数
	OutputParams   string    `json:"output_params"`    // 输出参数
	Status         bool      `json:"status"`           // 工具调用状态
	ExecutionTime  int64     `json:"execution_time"`   // 大模型选择工具的思考时间，单位为秒
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (m *ToolMessage) SetCreateTime(t time.Time) { m.CreatedAt = t }
func (m *ToolMessage) SetUpdateTime(t time.Time) { m.UpdatedAt = t }

func (m *ToolMessage) PrimaryKey() string {
	return "id"
}

func (m *ToolMessage) TableName() string {
	return "tool_messages"
}

func (m *ToolMessage) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if m.ID != "" {
		index["id"] = m.ID
	}

	return index
}
