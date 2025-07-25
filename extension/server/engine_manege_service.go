package server

import (
	"context"
	"golang.org/x/sync/singleflight"
	"oadin/extension/api/dto"
	interalDTO "oadin/internal/api/dto"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/provider"
	"oadin/internal/server"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"
	"os"
	"os/exec"
	"sync"
)

var group singleflight.Group
var downloaded sync.Map // 用于永久标记已下载
var downloading sync.Map

type EngineManageService interface {
	Exist(ctx context.Context, req dto.EngineManageRequest) int
	Install(ctx context.Context, req dto.EngineManageRequest) error
}

type EngineManageServiceImpl struct {
	AIGCService server.AIGCService
}

// NewEngineManageService creates a new instance of EngineManageServiceImpl
func NewEngineManageService() EngineManageService {
	return &EngineManageServiceImpl{}
}

func (s *EngineManageServiceImpl) Exist(ctx context.Context, req dto.EngineManageRequest) int {

	isDownloading := IsDownloading(req.ServiceName)

	if isDownloading {
		return 2
	}
	// Implement the logic to check if the engine exists
	// For now, we return true as a placeholder
	recommendConfig := getRecommendConfig(req.ServiceName)
	// Check if ollama is installed locally and if it is available.
	// If it is available, proceed to the next step. Otherwise, prompt that ollama is not installed.
	engineProvider := provider.GetModelEngine(recommendConfig.ModelEngine)
	engineConfig := engineProvider.GetConfig()

	cmd := exec.Command(engineConfig.ExecFile, "-h")
	err := cmd.Run()
	if err != nil {
		logger.LogicLogger.Info("Check model engine " + recommendConfig.ModelEngine + "  not exist...")
		reCheckCmd := exec.Command(engineConfig.ExecPath+"/"+engineConfig.ExecFile, "-h")
		err = reCheckCmd.Run()
		_, isExistErr := os.Stat(engineConfig.ExecPath + "/" + engineConfig.ExecFile)
		if err != nil && isExistErr != nil {
			logger.LogicLogger.Info("Model engine " + recommendConfig.ModelEngine + " not exist")
			return 0
		}
	}

	return 1
}

// Install installs the specified engine if it is not already installed.
func (s *EngineManageServiceImpl) Install(ctx context.Context, req dto.EngineManageRequest) error {
	installReq := &interalDTO.CreateAIGCServiceRequest{
		ServiceName:   req.ServiceName,
		SkipModelFlag: true,
	}
	storeKey := req.ServiceName

	// 第一步，判断是否已下载过
	if _, loaded := downloaded.LoadOrStore(storeKey, struct{}{}); loaded {
		// 已下载，无需再下
		return nil
	}
	// 第二步，防止同一时刻多协程重复下载
	group.Do(storeKey, func() (interface{}, error) {
		downloading.Store(storeKey, true)
		defer downloading.Delete(storeKey)
		// 只有第一个进来的协程真正执行下载
		_, err := s.AIGCService.CreateAIGCService(ctx, installReq)
		if err != nil {
			logger.LogicLogger.Error("Failed to install model engine", "error", err)
			return nil, bcode.ErrAIGCServiceInstallEngine
		}
		return nil, nil
	})

	return nil
}

// 查 询是否下载中
func IsDownloading(key string) bool {
	_, ok := downloading.Load(key)
	return ok
}

func getRecommendConfig(service string) types.RecommendConfig {
	recommendModelMap, _ := RecommendModels()
	recommendModelList := recommendModelMap[service]
	switch service {
	case types.ServiceChat:
		modelName := constants.DefaultChatModelName
		if len(recommendModelList) > 0 {
			modelName = recommendModelList[0].Name
		}
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         modelName,
			EngineDownloadUrl: server.OllamaEngineDownloadURL,
		}
	case types.ServiceEmbed:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         constants.DefaultEmbedModelName,
			EngineDownloadUrl: server.OllamaEngineDownloadURL,
		}
	case types.ServiceModels:
		return types.RecommendConfig{}
	case types.ServiceGenerate:
		modelName := constants.DefaultChatModelName
		if len(recommendModelList) > 0 {
			modelName = recommendModelList[0].Name
		}
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         modelName,
			EngineDownloadUrl: server.OllamaEngineDownloadURL,
		}
	case types.ServiceTextToImage:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultTextToImageModel,
			EngineDownloadUrl: server.OpenvinoEngineDownloadURL,
		}
	case types.ServiceSpeechToText:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultSpeechToTextModel,
			EngineDownloadUrl: server.OpenvinoEngineDownloadURL,
		}
	case types.ServiceSpeechToTextWS:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultSpeechToTextModel,
			EngineDownloadUrl: server.OpenvinoEngineDownloadURL,
		}
	default:
		return types.RecommendConfig{}
	}
}
