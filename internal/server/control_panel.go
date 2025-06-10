package server

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"byze/internal/api/dto"
	"byze/internal/provider"
	"byze/internal/types"
	"byze/internal/utils"
	"byze/internal/utils/bcode"
)

func GetModelFilePath(ctx context.Context) (*dto.GetModelFilePathResponse, error) {
	engineName := types.FlavorOllama
	engine := provider.GetModelEngine(engineName)
	engineConfig := engine.GetConfig()
	defaultPath := filepath.Join(engineConfig.EnginePath, "models")
	res := &dto.GetModelFilePathData{}
	value := os.Getenv("BYZE_OLLAMA_MODELS")
	if value != "" {
		res.Path = value
	} else {
		res.Path = defaultPath
	}

	return &dto.GetModelFilePathResponse{
		*bcode.ControlPanelCode,
		res,
	}, nil
}

func GetFilePathSize(ctx context.Context, req *dto.GetPathDiskSizeInfoRequest) (*dto.GetPathDiskSizeInfoResponse, error) {
	err := CheckFilePath(req.Path)
	if err != nil {
		return nil, err
	}
	res := &dto.GetPathDiskSizeInfoResponse{}
	sizeInfo, err := utils.SystemDiskSize(req.Path)
	if err != nil {
		return nil, err
	}
	res.Data = sizeInfo
	res.Bcode = *bcode.ControlPanelCode
	return res, nil
}

func ModifyModelFilePath(ctx context.Context, req *dto.ModifyModelFilePathRequest) (*dto.ModifyModelFilePathResponse, error) {
	if req.TargetPath == req.SourcePath {
		return &dto.ModifyModelFilePathResponse{}, bcode.ControlPanelPathStatusError
	}
	err := CheckFilePath(req.TargetPath)
	if err != nil {
		return nil, err
	}
	isTargetDirEmpty := utils.IsDirEmpty(req.TargetPath)
	if !isTargetDirEmpty {
		return &dto.ModifyModelFilePathResponse{}, bcode.ControlPanelPathStatusError
	}

	// Stop the engine before migration to avoid errors caused by processes still using the files.
	engine := provider.GetModelEngine("ollama")
	_ = engine.StopEngine()

	isSourceDirEmpty := utils.IsDirEmpty(req.SourcePath)
	if !isSourceDirEmpty {
		sourcePathSize, err := utils.GetFilePathTotalSize(req.SourcePath)
		if err != nil {
			return nil, bcode.ControlPanelPathCheckError
		}
		targetDiskSizeInfo, err := utils.SystemDiskSize(req.TargetPath)
		if err != nil {
			return nil, bcode.ControlPanelPathCheckError
		}
		if int(sourcePathSize) > targetDiskSizeInfo.FreeSize {
			return nil, bcode.ControlPanelPathSizeError
		}
		err = utils.CopyDir(req.SourcePath, req.TargetPath)
		if err != nil {
			return nil, bcode.ControlPanelCopyDirError
		}
	}

	envInfo := &utils.EnvVariables{
		Name:  "BYZE_OLLAMA_MODELS",
		Value: req.TargetPath,
	}
	err = utils.ModifySystemUserVariables(envInfo)
	if err != nil {
		err = utils.ClearDir(req.TargetPath)
		if err != nil {
			return nil, bcode.ControlPanelCopyDirError
		}
		return nil, bcode.ControlPanelCopyDirError
	}
	os.Setenv("BYZE_OLLAMA_MODELS", req.TargetPath)

	time.Sleep(1 * time.Second)

	err = utils.ClearDir(req.SourcePath)
	if err != nil {
		return nil, err
	}

	err = engine.InitEnv()
	if err != nil {
		return nil, err
	}

	err = engine.StartEngine()
	if err != nil {
		return nil, err
	}

	res := &dto.ModifyModelFilePathResponse{}
	res.Bcode = *bcode.ControlPanelCode
	res.Data = struct{}{}
	return res, nil
}

func CheckFilePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return bcode.ControlPanelPathCheckError
	}
	if !info.IsDir() {
		return bcode.ControlPanelPathCheckError
	}
	permissionStatus := utils.CheckPathPermission(path)
	if !permissionStatus {
		return bcode.ControlPanelPathPermissionError
	}
	return nil
}
