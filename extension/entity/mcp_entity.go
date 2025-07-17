package entity

import "time"

// McpKit 主实体结构体
type McpUserConfig struct {
	ID        int       `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	MCPID     string    `json:"mcp_id" gorm:"column:mcp_id;not null;"`                         // 关联的MCP主键
	Kits      string    `json:"kits" gorm:"column:kits;"`                                      // 禁用的工具集合（JSON数组存储）
	UserID    string    `json:"user_id" gorm:"column:user_id;"`                                // 关联用户
	Status    int       `json:"status" gorm:"column:status;default:0"`                         // 启用状态：0-禁用，1-启用
	Auth      string    `json:"auth" gorm:"column:auth;"`                                      // 认证授权码
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;default:CURRENT_TIMESTAMP"` // 创建时间
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`            // 修改时间
}

func (t *McpUserConfig) SetCreateTime(time time.Time) {
	t.CreatedAt = time
}

func (t *McpUserConfig) SetUpdateTime(time time.Time) {
	t.UpdatedAt = time
}

func (t *McpUserConfig) PrimaryKey() string {
	return "id"
}
func (t *McpUserConfig) TableName() string {
	return "mcp_user_config"
}

func (t *McpUserConfig) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if t.UserID != "" {
		index["user_id"] = t.UserID
	}

	if t.MCPID != "" {
		index["mcp_id"] = t.MCPID
	}

	return index

}
