package dto

// 注意：此文件应该只包含真正需要在系统间传递的DTO结构
// 内部解析用的结构体应该放在各自的service文件中

type EngineManageRequest struct {
	ProviderName  string `json:"provider_name"`
	ModelName     string `json:"model_name" validate:"required"`
	ServiceName   string `json:"service_name"`
	ServiceSource string `json:"service_source"`
}

type DownloadResponse struct {
	Status string `json:"status"`
	Data   string `json:"data,omitempty"`
}

type EngineDownloadRequest struct {
	EngineName string `json:"engineName" validate:"required"`
	Stream     bool   `json:"stream"`
}

type DownloadCheckDistRequest struct {
	EngineName string `json:"engineName" validate:"required"`
}

type ModelDownloadRequest struct {
	EngineName    string `json:"engineName" validate:"required"`
	ModelName     string `json:"modelName" validate:"required"`
	ModelType     string `json:"modelType" validate:"required"`
	Stream        bool   `json:"stream" validate:"required"`
}
