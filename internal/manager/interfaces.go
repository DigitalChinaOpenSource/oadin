package manager

import "oadin/internal/provider"

// ModelStateManager 模型状态管理接口
type ModelStateManager interface {
	GetCurrentModel() string
	MarkModelInUse(modelName string) error
	MarkModelIdle(modelName string) error
	SetCurrentModel(modelName string)
}

// ModelLoader 模型加载接口
type ModelLoader interface {
	SwitchModel(newModel string, provider provider.ModelServiceProvider) error
	LoadModel(modelName string, provider provider.ModelServiceProvider) error
	UnloadModel(modelName string, provider provider.ModelServiceProvider) error
}

// QueueStatusChecker 队列状态检查接口
type QueueStatusChecker interface {
	HasPendingRequests() bool
	HasActiveRequests() bool
	GetCurrentRequest() *QueuedRequest
}
