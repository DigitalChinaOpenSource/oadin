package vega

import (
	"byze/internal/utils/bcode"
	"time"
)

type QueryCloudModelJsonRequest struct {
	HybridPolicy string
}

type Model struct {
	Id             string    `json:"id"`
	SupplierId     string    `json:"supplier_id"`
	SupplierName   string    `json:"supplier_name"`
	Name           string    `json:"name"`
	Avatar         string    `json:"avatar"`
	Description    string    `json:"description"`
	Type           string    `json:"type"`
	Tags           []string  `json:"tags"`
	DeployMode     int       `json:"deploy_mode"`
	ParameterScale float32   `json:"parameter_scale"`
	FileSize       int       `json:"file_size"`
	MaxInput       int       `json:"max_input"`
	MaxOutput      int       `json:"max_output"`
	Popularity     int       `json:"popularity"`
	CreateAt       time.Time `json:"create_at"`
	UpdateAt       time.Time `json:"update_at"`
	CreateBy       string    `json:"create_by"`
	UpdateBy       string    `json:"update_by"`
}

type Supplier struct {
	Id           string    `json:"id"`
	Name         string    `json:"name"`
	Origin       string    `json:"origin"`
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	ServerName   string    `json:"server_name"`
	ContextScale int       `json:"context_scale"`
	CreateAt     time.Time `json:"create_at"`
	UpdateAt     time.Time `json:"update_at"`
	CreateBy     string    `json:"create_by"`
	UpdateBy     string    `json:"update_by"`
}

type SupplierAuth struct {
	Id            string    `json:"id"`
	UserId        string    `json:"user_id"`
	SupplierId    string    `json:"supplier_id"`
	RequestMethod string    `json:"request_method"`
	RequestUrl    string    `json:"request_url"`
	AuthType      string    `json:"auth_type"`
	AauthSecret   string    `json:"auth_secret"`
	Status        string    `json:"status"`
	CreateAt      time.Time `json:"create_at"`
	UpdateAt      time.Time `json:"update_at"`
}

type QueryCloudModelJsonRespond struct {
	bcode.Bcode
	Data []Model `json:"data"`
}
