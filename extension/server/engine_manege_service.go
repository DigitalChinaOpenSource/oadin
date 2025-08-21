package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"oadin/extension/api/dto"
	interalDTO "oadin/internal/api/dto"
	"oadin/internal/constants"
	"oadin/internal/logger"
	"oadin/internal/provider"
	"oadin/internal/server"
	"oadin/internal/types"
	"oadin/internal/datastore"
)

// Engine status constants
const (
	EngineNotExist    = 0
	EngineExist       = 1
	EngineDownloading = 2
)

var (
	// Single flight group to prevent duplicate downloads
	downloadGroup singleflight.Group
	// Map to track permanently downloaded engines
	downloadedEngines sync.Map
	// Map to track currently downloading engines
	downloadingEngines sync.Map
)

// EngineManageService defines the interface for engine management operations
type EngineManageService interface {
	// Exist checks if an engine exists and returns its status
	// Returns: 0 = not exist, 1 = exist, 2 = downloading
	Exist(ctx context.Context, req dto.EngineManageRequest) int
	// Install installs the specified engine if it is not already installed
	Install(ctx context.Context, req dto.EngineManageRequest) error

	CreateAIGCServiceSync(ctx context.Context, req *interalDTO.CreateAIGCServiceRequest) error
	CheckLocalModelExist(ctx context.Context, request dto.ModelDownloadRequest) error
}

// EngineManageServiceImpl implements the EngineManageService interface
type EngineManageServiceImpl struct {
	AIGCService server.AIGCService
	Ds datastore.Datastore
}

// NewEngineManageService creates a new instance of EngineManageServiceImpl
func NewEngineManageService() EngineManageService {
	return &EngineManageServiceImpl{
		AIGCService: server.NewAIGCService(),
		Ds: datastore.GetDefaultDatastore(),
	}
}

// Exist checks if the specified engine exists and returns its status
func (s *EngineManageServiceImpl) Exist(ctx context.Context, req dto.EngineManageRequest) int {
	// Check if engine is currently downloading
	if IsDownloading(req.ServiceName) {
		return EngineDownloading
	}

	// Get recommended configuration for the service
	recommendConfig := getRecommendConfig(req.ServiceName)
	if recommendConfig.ModelEngine == "" {
		logger.LogicLogger.Warn("No recommended engine found for service", "service", req.ServiceName)
		return EngineNotExist
	}

	// Get engine provider and configuration
	engineProvider := provider.GetModelEngine(recommendConfig.ModelEngine)
	if engineProvider == nil {
		logger.LogicLogger.Error("Failed to get engine provider", "engine", recommendConfig.ModelEngine)
		return EngineNotExist
	}

	engineConfig := engineProvider.GetConfig()

	// Try to execute the engine binary to check if it exists
	if exists := checkEngineExists(engineConfig); exists {
		return EngineExist
	}

	logger.LogicLogger.Info("Model engine not found", "engine", recommendConfig.ModelEngine)
	return EngineNotExist
}

// checkEngineExists checks if the engine binary exists and is executable
func checkEngineExists(engineConfig *types.EngineRecommendConfig) bool {
	// Try to execute the engine binary with help flag
	cmd := exec.Command(engineConfig.ExecFile, "-h")
	if err := cmd.Run(); err != nil {
		logger.LogicLogger.Debug("Primary engine check failed", "execFile", engineConfig.ExecFile, "error", err)

		// Try with full path
		fullPath := engineConfig.ExecPath + "/" + engineConfig.ExecFile
		reCheckCmd := exec.Command(fullPath, "-h")
		if err := reCheckCmd.Run(); err != nil {
			logger.LogicLogger.Debug("Secondary engine check failed", "fullPath", fullPath, "error", err)

			// Check if file exists at the expected location
			if _, statErr := os.Stat(fullPath); statErr != nil {
				logger.LogicLogger.Debug("Engine file not found", "path", fullPath, "error", statErr)
				return false
			}
		}
	}

	logger.LogicLogger.Debug("Engine found and executable", "engine", engineConfig.ExecFile)
	return true
}

// Install installs the specified engine if it is not already installed
func (s *EngineManageServiceImpl) Install(ctx context.Context, req dto.EngineManageRequest) error {
	storeKey := req.ServiceName

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use singleflight to prevent concurrent downloads and handle completion state
	result, err, shared := downloadGroup.Do(storeKey, func() (interface{}, error) {
		// Double-check if already downloaded after acquiring lock
		if _, loaded := downloadedEngines.Load(storeKey); loaded {
			logger.LogicLogger.Info("Engine already downloaded (detected in critical section)", "service", req.ServiceName)
			return true, nil // true indicates already downloaded
		}

		// Perform the actual installation with context
		success, installErr := s.performInstallWithContext(ctx, req)
		if installErr != nil {
			return false, installErr
		}

		// Mark as permanently downloaded only on success
		downloadedEngines.Store(storeKey, struct{}{})
		logger.LogicLogger.Info("Engine marked as permanently downloaded", "service", req.ServiceName)
		return success, nil
	})

	if err != nil {
		logger.LogicLogger.Error("Engine installation failed", "service", req.ServiceName, "error", err, "shared", shared)
		return err
	}

	if shared {
		logger.LogicLogger.Info("Engine installation completed by another goroutine", "service", req.ServiceName)
	}

	// result should be true if installation was successful or already completed
	if success, ok := result.(bool); ok && success {
		return nil
	}

	return fmt.Errorf("unexpected installation result for service %s", req.ServiceName)
}

// performInstallWithContext performs the actual installation process with context awareness
func (s *EngineManageServiceImpl) performInstallWithContext(ctx context.Context, req dto.EngineManageRequest) (bool, error) {
	storeKey := req.ServiceName

	// Mark as downloading
	downloadingEngines.Store(storeKey, true)
	defer func() {
		downloadingEngines.Delete(storeKey)
		logger.LogicLogger.Debug("Removed downloading marker", "service", req.ServiceName)
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		logger.LogicLogger.Warn("Installation cancelled due to context", "service", req.ServiceName, "error", ctx.Err())
		return false, ctx.Err()
	default:
	}

	logger.LogicLogger.Info("Starting engine installation", "service", req.ServiceName)

	installReq := &interalDTO.CreateAIGCServiceRequest{
		ServiceName:   req.ServiceName,
		SkipModelFlag: true,
	}

	// Perform installation with context
	_, err := s.AIGCService.CreateAIGCService(ctx, installReq)
	if err != nil {
		logger.LogicLogger.Error("Failed to install model engine", "service", req.ServiceName, "error", err)
		return false, fmt.Errorf("failed to install engine for service %s: %w", req.ServiceName, err)
	}

	logger.LogicLogger.Info("Engine installation completed successfully", "service", req.ServiceName)
	return true, nil
}


func (s *EngineManageServiceImpl) CreateAIGCServiceSync(ctx context.Context, req *interalDTO.CreateAIGCServiceRequest) error {
	// Synchronously create the AIGC service
	_, err := s.AIGCService.CreateAIGCService(ctx, req)
	if err == nil {
		time.Sleep(1 * time.Second)
	}
	return err
}

func (s *EngineManageServiceImpl) CheckLocalModelExist(ctx context.Context, request dto.ModelDownloadRequest) error {
	m := &types.Model{}
	m.ModelName = request.ModelName
	m.ProviderName = fmt.Sprintf("local_%s_%s", request.EngineName, request.ModelType)
	m.Status = "downloaded"
	err := s.Ds.Get(ctx, m)
	fmt.Println("CheckLocalModelExist:", err)
	return err
}

// IsDownloading checks if an engine is currently being downloaded
func IsDownloading(serviceName string) bool {
	_, ok := downloadingEngines.Load(serviceName)
	return ok
}

// getRecommendConfig returns the recommended configuration for a given service
func getRecommendConfig(service string) types.RecommendConfig {
	recommendModelMap, err := RecommendModels()
	if err != nil {
		logger.LogicLogger.Error("Failed to get recommend models", "error", err)
		return types.RecommendConfig{}
	}

	recommendModelList := recommendModelMap[service]

	switch service {
	case types.ServiceChat, types.ServiceGenerate:
		return createOllamaConfig(recommendModelList, constants.DefaultChatModelName)

	case types.ServiceEmbed:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOllama,
			ModelName:         constants.DefaultEmbedModelName,
			EngineDownloadUrl: server.OllamaEngineDownloadURL,
		}

	case types.ServiceTextToImage:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultTextToImageModel,
			EngineDownloadUrl: server.OpenvinoEngineDownloadURL,
		}

	case types.ServiceSpeechToText, types.ServiceSpeechToTextWS:
		return types.RecommendConfig{
			ModelEngine:       types.FlavorOpenvino,
			ModelName:         constants.DefaultSpeechToTextModel,
			EngineDownloadUrl: server.OpenvinoEngineDownloadURL,
		}

	case types.ServiceModels:
		return types.RecommendConfig{}

	default:
		logger.LogicLogger.Warn("Unknown service type", "service", service)
		return types.RecommendConfig{}
	}
}

// createOllamaConfig creates a configuration for Ollama-based services
func createOllamaConfig(recommendModelList []dto.RecommendModelData, defaultModelName string) types.RecommendConfig {
	modelName := defaultModelName
	if len(recommendModelList) > 0 {
		modelName = recommendModelList[0].Name
	}

	return types.RecommendConfig{
		ModelEngine:       types.FlavorOllama,
		ModelName:         modelName,
		EngineDownloadUrl: server.OllamaEngineDownloadURL,
	}
}
