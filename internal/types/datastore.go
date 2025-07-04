package types

import (
	"time"
)

const (
	ServiceSourceLocal  = "local"
	ServiceSourceRemote = "remote"

	FlavorTencent     = "tencent"
	FlavorDeepSeek    = "deepseek"
	FlavorOpenAI      = "openai"
	FlavorOllama      = "ollama"
	FlavorBaidu       = "baidu"
	FlavorAliYun      = "aliyun"
	FlavorSmartVision = "smartvision"

	AuthTypeNone        = "none"
	AuthTypeApiKey      = "apikey"
	AuthTypeToken       = "token"
	AuthTypeCredentials = "credentials"

	ServiceChat        = "chat"
	ServiceModels      = "models"
	ServiceGenerate    = "generate"
	ServiceEmbed       = "embed"
	ServiceTextToImage = "text_to_image"

	HybridPolicyDefault = "default"
	HybridPolicyLocal   = "always_local"
	HybridPolicyRemote  = "always_remote"

	VersionRecordStatusInstalled = 1
	VersionRecordStatusUpdated   = 2
)

var (
	SupportService      = []string{ServiceEmbed, ServiceModels, ServiceChat, ServiceGenerate, ServiceTextToImage}
	SupportHybridPolicy = []string{HybridPolicyDefault, HybridPolicyLocal, HybridPolicyRemote}
	SupportAuthType     = []string{AuthTypeNone, AuthTypeApiKey, AuthTypeToken, AuthTypeCredentials}
	SupportFlavor       = []string{FlavorDeepSeek, FlavorOpenAI, FlavorTencent, FlavorOllama, FlavorBaidu, FlavorAliYun, FlavorSmartVision}
)

// Service  table structure
type Service struct {
	Name           string    `gorm:"primaryKey;column:name" json:"name"`
	HybridPolicy   string    `gorm:"column:hybrid_policy;not null;default:default" json:"hybrid_policy"`
	RemoteProvider string    `gorm:"column:remote_provider;not null;default:''" json:"remote_provider"`
	LocalProvider  string    `gorm:"column:local_provider;not null;default:''" json:"local_provider"`
	Status         int       `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (t *Service) SetCreateTime(time time.Time) {
	t.CreatedAt = time
}

func (t *Service) SetUpdateTime(time time.Time) {
	t.UpdatedAt = time
}

func (t *Service) PrimaryKey() string {
	return "name"
}

func (t *Service) TableName() string {
	return "oadin_service"
}

func (t *Service) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if t.Name != "" {
		index["name"] = t.Name
	}

	return index
}

// ServiceProvider Service provider table structure
type ServiceProvider struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ProviderName  string    `gorm:"column:provider_name" json:"provider_name"`
	ServiceName   string    `gorm:"column:service_name" json:"service_name"`
	ServiceSource string    `gorm:"column:service_source;default:local" json:"service_source"`
	Desc          string    `gorm:"column:desc" json:"desc"`
	Method        string    `gorm:"column:method" json:"method"`
	URL           string    `gorm:"column:url" json:"url"`
	AuthType      string    `gorm:"column:auth_type" json:"auth_type"`
	AuthKey       string    `gorm:"column:auth_key" json:"auth_key"`
	Flavor        string    `gorm:"column:flavor" json:"flavor"`
	ExtraHeaders  string    `gorm:"column:extra_headers;default:'{}'" json:"extra_headers"`
	ExtraJSONBody string    `gorm:"column:extra_json_body;default:'{}'" json:"extra_json_body"`
	Properties    string    `gorm:"column:properties;default:'{}'" json:"properties"`
	Status        int       `gorm:"column:status;not null;default:0" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (t *ServiceProvider) SetCreateTime(time time.Time) {
	t.CreatedAt = time
}

func (t *ServiceProvider) SetUpdateTime(time time.Time) {
	t.UpdatedAt = time
}

func (t *ServiceProvider) PrimaryKey() string {
	return "id"
}

func (t *ServiceProvider) TableName() string {
	return "oadin_service_provider"
}

func (t *ServiceProvider) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if t.ProviderName != "" {
		index["provider_name"] = t.ProviderName
	}

	if t.ServiceSource != "" {
		index["service_source"] = t.ServiceSource
	}

	if t.ServiceName != "" {
		index["service_name"] = t.ServiceName
	}

	if t.Flavor != "" {
		index["flavor"] = t.Flavor
	}
	return index
}

// Model  table structure
type Model struct {
	ID              int       `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	ModelName       string    `gorm:"column:model_name;not null" json:"model_name"`
	ProviderName    string    `gorm:"column:provider_name" json:"provider_name"`
	Status          string    `gorm:"column:status;not null" json:"status"`
	CreatedAt       time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	ThinkingEnabled bool      `gorm:"column:thinking_enabled;default:false" json:"thinkingEnabled"`
	OllamaRegistry  string    `gorm:"column:ollama_registry" json:"ollamaRegistry"`
}

func (t *Model) SetCreateTime(time time.Time) {
	t.CreatedAt = time
}

func (t *Model) SetUpdateTime(time time.Time) {
	t.UpdatedAt = time
}

func (t *Model) PrimaryKey() string {
	return "id"
}

func (t *Model) TableName() string {
	return "oadin_model"
}

func (t *Model) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if t.ModelName != "" {
		index["model_name"] = t.ModelName
	}

	if t.ProviderName != "" {
		index["provider_name"] = t.ProviderName
	}

	return index
}

// VersionUpdateRecord  table structure
type VersionUpdateRecord struct {
	ID           int       `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	Version      string    `gorm:"column:version;not null" json:"version"`
	ReleaseNotes string    `gorm:"column:release_notes;not null" json:"release_notes"`
	Status       int       `gorm:"column:status;not null" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (t *VersionUpdateRecord) SetCreateTime(time time.Time) {
	t.CreatedAt = time
}

func (t *VersionUpdateRecord) SetUpdateTime(time time.Time) {
	t.UpdatedAt = time
}

func (t *VersionUpdateRecord) PrimaryKey() string {
	return "id"
}

func (t *VersionUpdateRecord) TableName() string {
	return "oadin_version_update_record"
}

func (t *VersionUpdateRecord) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if t.Version != "" {
		index["version"] = t.Version
	}

	return index
}

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
