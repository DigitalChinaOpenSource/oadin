package dto

import (
	"time"

	"intel.com/aog/internal/utils/bcode"
)

type CreateAIGCServiceRequest struct {
	ServiceName   string `json:"service_name" validate:"required"`
	ServiceSource string `json:"service_source" validate:"required"`
	ApiFlavor     string `json:"api_flavor" validate:"required"`
	ProviderName  string `json:"provider_name" validate:"required"`
	Desc          string `json:"desc"`
	Method        string `json:"method"`
	Url           string `json:"url"`
	AuthType      string `json:"auth_type" validate:"required"`
	AuthKey       string `json:"auth_key"`
	ExtraHeaders  string `json:"extra_headers"`
	ExtraJsonBody string `json:"extra_json_body"`
	Properties    string `json:"properties"`
	SkipModelFlag bool   `json:"skip_model" default:"true"`
	ModelName     string `json:"model_name"`
}

type UpdateAIGCServiceRequest struct {
	ServiceName    string `json:"service_name" validate:"required"`
	HybridPolicy   string `json:"hybrid_policy"`
	RemoteProvider string `json:"remote_provider"`
	LocalProvider  string `json:"local_provider"`
}

type DeleteAIGCServiceRequest struct{}

type GetAIGCServiceRequest struct{}

type ExportServiceRequest struct {
	ServiceName  string `json:"service_name"`
	ProviderName string `json:"provider_name"`
	ModelName    string `json:"model_name"`
}

type ExportServiceResponse struct {
	Version          string                          `json:"version"`
	Services         map[string]ServiceEntry         `json:"services"`
	ServiceProviders map[string]ServiceProviderEntry `json:"service_providers"`
}
type ServiceEntry struct {
	ServiceProviders ServiceProviderInfo `json:"service_providers"`
	HybridPolicy     string              `json:"hybrid_policy"`
}
type ServiceProviderInfo struct {
	Local  string `json:"local"`
	Remote string `json:"remote"`
}
type ServiceProviderEntry struct {
	ServiceName   string   `json:"service_name"`
	ServiceSource string   `json:"service_source"`
	Desc          string   `json:"desc"`
	APIFlavor     string   `json:"api_flavor"`
	Method        string   `json:"method"`
	URL           string   `json:"url"`
	AuthType      string   `json:"auth_type"`
	AuthKey       string   `json:"auth_key"`
	Models        []string `json:"models"`
}

type ImportServiceRequest struct {
	Version          string                          `json:"version"`
	Services         map[string]ServiceEntry         `json:"services"`
	ServiceProviders map[string]ServiceProviderEntry `json:"service_providers"`
}

type ImportServiceResponse struct {
	bcode.Bcode
}

type GetAIGCServicesRequest struct {
	ServiceName string `json:"service_name,omitempty"`
}

type CreateAIGCServiceResponse struct {
	bcode.Bcode
}

type UpdateAIGCServiceResponse struct {
	bcode.Bcode
}

type DeleteAIGCServiceResponse struct{}

type GetAIGCServiceResponse struct{}

type GetAIGCServicesResponse struct {
	bcode.Bcode
	Data []Service `json:"data"`
}

type Service struct {
	ServiceName    string    `json:"service_name"`
	HybridPolicy   string    `json:"hybrid_policy"`
	RemoteProvider string    `json:"remote_provider"`
	LocalProvider  string    `json:"local_provider"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateModelRequest struct {
	ProviderName  string `json:"provider_name"`
	ModelName     string `json:"model_name" validate:"required"`
	ServiceName   string `json:"service_name" validate:"required"`
	ServiceSource string `json:"service_source" validate:"required"`
}

type CreateModelStreamRequest struct {
	ProviderName  string `json:"provider_name"`
	ModelName     string `json:"model_name" validate:"required"`
	ServiceName   string `json:"service_name"`
	ServiceSource string `json:"service_source"`
}

type DeleteModelRequest struct {
	ProviderName  string `json:"provider_name"`
	ModelName     string `json:"model_name" validate:"required"`
	ServiceName   string `json:"service_name" validate:"required"`
	ServiceSource string `json:"service_source" validate:"required"`
}

type GetModelsRequest struct {
	ProviderName string `form:"provider_name,omitempty"`
	ModelName    string `form:"model_name,omitempty"`
	ServiceName  string `form:"service_name,omitempty"`
}

type GetModelListRequest struct {
	ServiceSource string `form:"service_source" validate:"required"`
	Flavor        string `form:"flavor" validate:"required"`
}

type ModelStreamCancelRequest struct {
	ModelName string `json:"model_name" validate:"required"`
}

type CreateModelResponse struct {
	bcode.Bcode
}

type DeleteModelResponse struct {
	bcode.Bcode
}

type GetModelsResponse struct {
	bcode.Bcode
	Data []Model `json:"data"`
}

type RecommendModelResponse struct {
	bcode.Bcode
	Data map[string][]RecommendModelData `json:"data"`
}

type ModelStreamCancelResponse struct {
	bcode.Bcode
}

type Model struct {
	ModelName    string    `json:"model_name"`
	ProviderName string    `json:"provider_name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type LocalSupportModelData struct {
	OllamaId    string  `json:"id"`
	Name        string  `json:"name"`
	Avatar      string  `json:"avatar"`
	Description string  `json:"description"`
	Class       string  `json:"class"`
	Flavor      string  `json:"flavor"`
	Size        string  `json:"size"`
	ParamsSize  float32 `json:"params_size"`
}

type RecommendModelData struct {
	Service         string   `json:"service_name"`
	Flavor          string   `json:"api_flavor"`
	Method          string   `json:"method" default:"POST"`
	Desc            string   `json:"desc"`
	Url             string   `json:"url"`
	AuthType        string   `json:"auth_type"`
	AuthApplyUrl    string   `json:"auth_apply_url"`
	AuthFields      []string `json:"auth_fields"`
	Name            string   `json:"name"`
	ServiceProvider string   `json:"service_provider_name"`
	Size            string   `json:"size"`
	IsRecommended   bool     `json:"is_recommended" default:"false"`
	Status          string   `json:"status"`
	Avatar          string   `json:"avatar"`
	CanSelect       bool     `json:"can_select" default:"false"`
	Class           string   `json:"class"`
	OllamaId        string   `json:"ollama_id"`
	ParamsSize      float32  `json:"params_size"`
}

type CreateServiceProviderRequest struct {
	ServiceName   string   `json:"service_name" validate:"required"`
	ServiceSource string   `json:"service_source" validate:"required"`
	ApiFlavor     string   `json:"api_flavor" validate:"required"`
	ProviderName  string   `json:"provider_name" validate:"required"`
	Desc          string   `json:"desc"`
	Method        string   `json:"method"`
	Url           string   `json:"url"`
	AuthType      string   `json:"auth_type"`
	AuthKey       string   `json:"auth_key"`
	Models        []string `json:"models"`
	ExtraHeaders  string   `json:"extra_headers"`
	ExtraJsonBody string   `json:"extra_json_body"`
	Properties    string   `json:"properties"`
}

type UpdateServiceProviderRequest struct {
	ProviderName  string   `json:"provider_name" validate:"required"`
	ServiceName   string   `json:"service_name"`
	ServiceSource string   `json:"service_source"`
	ApiFlavor     string   `json:"api_flavor"`
	Desc          string   `json:"desc"`
	Method        string   `json:"method"`
	Url           string   `json:"url"`
	AuthType      string   `json:"auth_type"`
	AuthKey       string   `json:"auth_key"`
	Models        []string `json:"models"`
	ExtraHeaders  string   `json:"extra_headers"`
	ExtraJsonBody string   `json:"extra_json_body"`
	Properties    string   `json:"properties"`
}

type DeleteServiceProviderRequest struct {
	ProviderName string `json:"provider_name" validate:"required"`
}

type GetServiceProviderRequest struct{}

type GetServiceProvidersRequest struct {
	ServiceName   string `json:"service_name,omitempty"`
	ServiceSource string `json:"service_source,omitempty"`
	ProviderName  string `json:"provider_name,omitempty"`
	ApiFlavor     string `json:"api_flavor,omitempty"`
}

type CreateServiceProviderResponse struct {
	bcode.Bcode
}

type UpdateServiceProviderResponse struct {
	bcode.Bcode
}

type DeleteServiceProviderResponse struct {
	bcode.Bcode
}

type GetServiceProviderResponse struct{}

type GetServiceProvidersResponse struct {
	bcode.Bcode
	Data []ServiceProvider `json:"data"`
}

type ServiceProvider struct {
	ProviderName  string    `json:"provider_name"`
	ServiceName   string    `json:"service_name"`
	ServiceSource string    `json:"service_source"`
	Desc          string    `json:"desc"`
	AuthType      string    `json:"auth_type"`
	AuthKey       string    `json:"auth_key"`
	Flavor        string    `json:"flavor"`
	Properties    string    `json:"properties"`
	Models        []string  `json:"models"`
	Status        int       `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
