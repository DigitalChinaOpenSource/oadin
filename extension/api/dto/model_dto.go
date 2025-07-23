package dto

import (
	"oadin/internal/utils/bcode"
	"time"
)

type SmartVisionSupportModelRequest struct {
	EnvType string `form:"env_type" validate:"required"`
}

type SmartVisionSupportModelResponse struct {
	bcode.Bcode
	Data SmartVisionSupportModelRes `json:"data"`
}
type SmartVisionSupportModelRes struct {
	Code int                    `json:"code"`
	Data []SmartVisionModelData `json:"data"`
}

type SmartVisionModelData struct {
	ID                 int                          `json:"id"`
	Name               string                       `json:"name"`
	Avatar             string                       `json:"avatar"`
	Type               int                          `json:"type"`
	Provider           string                       `json:"provider"`
	ModelKey           string                       `json:"modelKey"`
	CredentialParamsID string                       `json:"credentialParamsId"`
	Introduce          string                       `json:"introduce"`
	Tags               []string                     `json:"tags"`
	CredentialParams   []SmartVisionCredentialParam `json:"credentialParams"`
	CanSelect          bool                         `json:"can_select"`
	CreatedAt          time.Time                    `json:"created_at"`
}

type SmartVisionCredentialParam struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        string      `json:"type"`
	Placeholder string      `json:"placeholder"`
	Required    int         `json:"required"`
	Value       interface{} `json:"value"`
	Sort        int         `json:"sort"`
	CreateTime  int64       `json:"createTime"`
	UpdateTime  int64       `json:"updateTime"`
}

type GetSupportModelRequest struct {
	Flavor        string `form:"flavor"`
	EnvType       string `form:"env_type"`
	ServiceSource string `form:"service_source" validate:"required"`
	Mine          bool   `form:"mine" default:"false"`
	PageSize      int    `form:"page_size"`
	Page          int    `form:"page"`
}

type GetSupportModelResponseData struct {
	Data      []RecommendModelData `json:"data"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
	Total     int                  `json:"total"`
	TotalPage int                  `json:"total_page"`
}

type GetSupportModelResponse struct {
	bcode.Bcode
	Data GetSupportModelResponseData `json:"data"`
}

type RecommendModelData struct {
	Id                  string    `json:"id"`
	Service             string    `json:"service_name"`
	ApiFlavor           string    `json:"api_flavor"`
	Flavor              string    `json:"flavor"`
	Method              string    `json:"method" default:"POST"`
	Desc                string    `json:"desc"`
	Url                 string    `json:"url"`
	AuthType            string    `json:"auth_type"`
	AuthApplyUrl        string    `json:"auth_apply_url"`
	AuthFields          []string  `json:"auth_fields"`
	Name                string    `json:"name"`
	ServiceProvider     string    `json:"service_provider_name"`
	Size                string    `json:"size"`
	IsRecommended       bool      `json:"is_recommended" default:"false"`
	Status              string    `json:"status"`
	Avatar              string    `json:"avatar"`
	CanSelect           bool      `json:"can_select" default:"false"`
	Class               []string  `json:"class"`
	OllamaId            string    `json:"ollama_id"`
	ParamsSize          float32   `json:"params_size"`
	InputLength         int       `json:"input_length"`
	OutputLength        int       `json:"output_length"`
	Source              string    `json:"source"`
	SmartVisionProvider string    `json:"smartvision_provider"`
	SmartVisionModelKey string    `json:"smartvision_model_key"`
	IsDownloaded        bool      `json:"is_downloaded" default:"false"`
	Think               bool      `json:"think" default:"false"`
	ThinkSwitch         bool      `json:"think_switch" default:"false"`
	Tools               bool      `json:"tools" default:"false"`
	Context             float32   `json:"context" default:"0"`
	CreatedAt           time.Time `json:"created_at"`
}
