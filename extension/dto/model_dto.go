package dto

type SmartVisionSupportModelRequest struct {
	EnvType string `form:"env_type" validate:"required"`
}
