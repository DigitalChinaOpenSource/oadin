package server

import (
	"byze/internal/api/dto"
	"byze/internal/cache"
	"context"
	"log/slog"
)

type System interface {
	ModifyRegistry(ctx context.Context, url string) error
	SetProxy(ctx context.Context, req dto.ProxyRequest) error
	SwitchProxy(ctx context.Context, enabled bool) error
	GetSystemSettings(ctx context.Context) (*cache.SystemSettings, error)
	GetOllamaRegistry() (string, error)
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
		return err
	}
	settings.OllamaRegistry = url

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("设置仓库地址失败", "error", err)
		return err
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
		return err
	}

	// 更新代理地址
	settings.SystemProxy.Username = req.Username
	settings.SystemProxy.Password = req.Password
	settings.SystemProxy.Endpoint = req.Endpoint

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("设置代理地址失败", "error", err)
		return err
	}
	return nil

}

// SwitchProxy 切换代理启用状态
func (s *SystemImpl) SwitchProxy(ctx context.Context, enabled bool) error {
	// 从用户配置文件中读取系统设置
	var settings cache.SystemSettings
	err := cache.ReadSystemSettings(&settings)
	if err != nil {
		slog.Error("切换代理启用状态失败", "error", err)
		return err
	}

	// 更新代理启用状态
	settings.SystemProxy.Enabled = enabled

	// 将修改后的设置写回用户配置文件
	err = cache.WriteSystemSettings(settings)
	if err != nil {
		slog.Error("切换代理启用状态失败", "error", err)
		return err
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
