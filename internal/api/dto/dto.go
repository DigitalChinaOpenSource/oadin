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
	ServiceName string `json:"service_name"`
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

type DeleteModelRequest struct {
	ProviderName  string `json:"provider_name"`
	ModelName     string `json:"model_name" validate:"required"`
	ServiceName   string `json:"service_name" validate:"required"`
	ServiceSource string `json:"service_source" validate:"required"`
}

type GetModelsRequest struct {
	ProviderName string `json:"provider_name"`
	ModelName    string `json:"model_name"`
	ServiceName  string `json:"service_name"`
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

type Model struct {
	ModelName    string    `json:"model_name"`
	ProviderName string    `json:"provider_name"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
	ProviderName  string `json:"provider_name" validate:"required"`
	ServiceName   string `json:"service_name"`
	ServiceSource string `json:"service_source"`
	ApiFlavor     string `json:"api_flavor"`
	Desc          string `json:"desc"`
	Method        string `json:"method"`
	Url           string `json:"url"`
	AuthType      string `json:"auth_type"`
	AuthKey       string `json:"auth_key"`
	ExtraHeaders  string `json:"extra_headers"`
	ExtraJsonBody string `json:"extra_json_body"`
	Properties    string `json:"properties"`
}

type DeleteServiceProviderRequest struct {
	ProviderName string `json:"provider_name" validate:"required"`
}

type GetServiceProviderRequest struct{}

type GetServiceProvidersRequest struct {
	ServiceName   string `json:"service_name"`
	ServiceSource string `json:"service_source"`
	ProviderName  string `json:"provider_name"`
	ApiFlavor     string `json:"api_flavor"`
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
