package dto

import (
	"oadin/extension/utils/bcode"
	"oadin/internal/types"
)

type ModifyModelFilePathRequest struct {
	SourcePath string `json:"source_path" validate:"required"`
	TargetPath string `json:"target_path" validate:"required"`
}

type ModifyModelFilePathResponse struct {
	bcode.Bcode
	Data struct{}
}

type GetPathDiskSizeInfoRequest struct {
	Path string `form:"path" validate:"required"`
}

type GetPathDiskSizeInfoResponse struct {
	bcode.Bcode
	Data *types.PathDiskSizeInfo `json:"data"`
}

type GetModelFilePathResponse struct {
	bcode.Bcode
	Data *GetModelFilePathData `json:"data"`
}

type GetModelFilePathData struct {
	Path string `json:"path"`
}
