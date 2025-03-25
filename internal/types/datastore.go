package types

import (
	"time"
)

const (
	ServiceSourceLocal  = "local"
	ServiceSourceRemote = "remote"

	FlavorTencent  = "tencent"
	FlavorDeepSeek = "deepseek"
	FlavorOpenAI   = "openai"
	FlavorOllama   = "ollama"
	FlavorBaidu    = "baidu"
	FlavorAliYun   = "aliyun"

	AuthTypeNone   = "none"
	AuthTypeApiKey = "apikey"
	AuthTypeToken  = "token"

	ServiceChat        = "chat"
	ServiceModels      = "models"
	ServiceGenerate    = "generate"
	ServiceEmbed       = "embed"
	ServiceTextToImage = "text_to_image"

	HybridPolicyDefault = "default"
	HybridPolicyLocal   = "always_local"
	HybridPolicyRemote  = "always_remote"
)

var (
	SupportService      = []string{ServiceEmbed, ServiceModels, ServiceChat, ServiceGenerate}
	SupportHybridPolicy = []string{HybridPolicyDefault, HybridPolicyLocal, HybridPolicyRemote}
	SupportAuthType     = []string{AuthTypeNone, AuthTypeApiKey, AuthTypeToken}
	SupportFlavor       = []string{FlavorDeepSeek, FlavorOpenAI, FlavorTencent, FlavorOllama}
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
	return "aog_service"
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
	Status        int       `gorm:"column:status;not null;default:1" json:"status"`
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
	return "aog_service_provider"
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
	ID           int       `gorm:"primaryKey;column:id;autoIncrement" json:"id"`
	ModelName    string    `gorm:"column:model_name;not null" json:"model_name"`
	ProviderName string    `gorm:"column:provider_name" json:"provider_name"`
	Status       string    `gorm:"column:status;not null" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
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
	return "aog_model"
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
