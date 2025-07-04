package server

import (
	"context"
	"log/slog"
	"oadin/internal/api/dto"
	"oadin/internal/cache"
	"oadin/internal/provider"
	"oadin/internal/utils/bcode"
)

type System interface {
	ModifyRegistry(ctx context.Context, url string) error
	SetProxy(ctx context.Context, req dto.ProxyRequest) error
	SwitchProxy(ctx context.Context, enabled bool) error
	GetSystemSettings(ctx context.Context) (*cache.SystemSettings, error)
	GetOllamaRegistry() (string, error)
	RestartOllama(ctx context.Context) error
}

type SystemImpl struct {
}

// NewSystemImpl 创建一个新的 SystemImpl 实例
func NewSystemImpl() System {
	return &SystemImpl{}
}

// ModifyRegistry 修改仓库地址
func (s *SystemImpl) ModifyRegistry(ctx context.Context, url string) error {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("设置仓库地址失败", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "设置仓库地址失败")
	}
	settings.OllamaRegistry = url

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("设置仓库地址失败", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "设置仓库地址失败")
	}
	return nil
}

// SetProxy 设置代理地址
func (s *SystemImpl) SetProxy(ctx context.Context, req dto.ProxyRequest) error {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("设置代理地址失败", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "设置代理地址失败")
	}
	// 使用临时变量存储代理设置
	tempProxy := settings.SystemProxy

	// 更新代理地址
	settings.SystemProxy.Username = req.Username
	settings.SystemProxy.Password = req.Password
	settings.SystemProxy.Endpoint = req.Endpoint
	// 设置代理状态
	settings.SystemProxy.Enabled = true

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("设置代理地址失败", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "设置代理地址失败")
	}

	// 重启Ollama服务以应用新的代理设置
	err = s.RestartOllama(ctx)
	if err != nil {
		slog.Error("重启Ollama失败", "error", err)
		// 回滚代理设置
		settings.SystemProxy = tempProxy // 清空代理设置
		cache.WriteSystemSettings(settings)
		return err
	}

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("设置代理地址失败", "error", err)
		// 回滚代理设置
		settings.SystemProxy = tempProxy // 清空代理设置
		return bcode.HttpError(bcode.ControlPanelSystemError, "设置代理地址失败")
	}
	return nil

}

// SwitchProxy 切换代理启用状态
func (s *SystemImpl) SwitchProxy(ctx context.Context, enabled bool) error {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	temp := settings.SystemProxy.Enabled
	// 更新代理启用状态
	settings.SystemProxy.Enabled = enabled
	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)

	if settings.SystemProxy.Endpoint == "" && enabled {
		slog.Error("代理地址不能为空，请先设置代理地址")
		settings.SystemProxy.Enabled = temp // 回滚代理启用状态
		// 将修改后的设置写回用户配置文件
		err = cache.WriteSystemSettings(settings)
		return bcode.HttpError(bcode.ControlPanelSystemError, "代理地址不能为空，请先设置代理地址")
	}

	err = s.RestartOllama(ctx)
	if err != nil {
		slog.Error("切换代理启用状态失败", "error", err)

		settings.SystemProxy.Enabled = temp // 回滚代理启用状态
		// 将修改后的设置写回用户配置文件
		err = cache.WriteSystemSettings(settings)
		return bcode.HttpError(bcode.ControlPanelSystemError, "切换代理启用状态失败")
	}
	return nil
}

func (s *SystemImpl) RestartOllama(ctx context.Context) error {
	// 探查ollama服务是否在运行模型
	engine := provider.GetModelEngine("ollama")
	engineConfig := engine.GetConfig()
	if engineConfig.StartStatus == 0 {
		return bcode.HttpError(bcode.ErrModelEngineIsBeingOperatedOn, "无法切换代理启用状态，当前有模型正在运行，请先停止模型")
	}
	engineConfig.StartStatus = 0
	defer func() {
		engineConfig.StartStatus = 1
	}()
	err := engine.HealthCheck()
	if err != nil {
		slog.Error("无法切换代理启用状态，ollama服务已关闭", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，ollama服务已关闭")
	}

	runModels, err := engine.GetRunModels(ctx)
	if err != nil {
		slog.Error("获取正在运行的模型失败", "error", err)
		return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，有正在运行的模型")
	}
	if len(runModels.Models) != 0 {
		slog.Error("无法切换代理启用状态，当前有模型正在运行，请先停止模型")
		return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，当前有模型正在运行，请先停止模型")
	} else {
		err = engine.StopEngine()
		if err != nil {
			slog.Error("停止引擎失败", "error", err)
			return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，ollama服务无法重启")
		}
		err = engine.InitEnv()
		if err != nil {
			return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，ollama服务无法重启")
		}
		err = engine.StartEngine()
		if err != nil {
			slog.Error("启动引擎失败", "error", err)
			return bcode.HttpError(bcode.ControlPanelSystemError, "无法切换代理启用状态，ollama服务无法重启")
		}

	}
	return nil

}

// GetSystemSettings 获取系统设置
func (s *SystemImpl) GetSystemSettings(ctx context.Context) (*cache.SystemSettings, error) {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("获取系统设置失败", "error", err)
		return nil, err
	}
	return &settings, nil
}

// GetOllamaRegistry 获取Ollama仓库地址
func (s *SystemImpl) GetOllamaRegistry() (string, error) {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("获取Ollama仓库地址失败", "error", err)
		return "", err
	}
	return settings.OllamaRegistry, nil
}
