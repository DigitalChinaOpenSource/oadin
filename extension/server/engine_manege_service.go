package server

import (
	"context"
	"oadin/extension/api/dto"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/provider"
	"oadin/internal/server"
	"oadin/internal/types"
	"os"
	"os/exec"
)

type EngineManageService interface {
	Exist(ctx context.Context, req dto.EngineManageRequest) bool
}

type EngineManageServiceImpl struct {
}

// NewEngineManageService creates a new instance of EngineManageServiceImpl
func NewEngineManageService() EngineManageService {
	return &EngineManageServiceImpl{}
}

func (s *EngineManageServiceImpl) Exist(ctx context.Context, req dto.EngineManageRequest) bool {
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
			return false
		}
	}

	return true
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
