package vega

import (
	"time"
)

type QueryCloudModelJsonRequest struct {
	HybridPolicy string
	Page         int
	PageSize     int
}

type Model struct {
	Id             string   `json:"id"`
	Flavor         string   `json:"flavor"`
	Name           string   `json:"name"`
	Avatar         string   `json:"avatar"`
	Description    string   `json:"description"`
	Type           string   `json:"type"`
	Tags           []string `json:"tags"`
	DeployMode     string   `json:"deployMode"`
	OllamaId       string   `json:"ollamaId"`
	ParameterScale string   `json:"parameter_scale"`
	FileSize       string   `json:"fileSize"`
	MaxInput       string   `json:"maxInput"`
	MaxOutput      string   `json:"maxOutput"`
	// Popularity     int       `json:"popularity"`
	// CreateAt       time.Time `json:"create_at"`
	// UpdateAt       time.Time `json:"update_at"`
	// CreateBy       string    `json:"create_by"`
	// UpdateBy       string    `json:"update_by"`
	// 本地
}

type QueryCloudSupplierJsonRequest struct {
	Page     int
	PageSize int
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
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []Model `json:"list"`
	} `json:"data"`
}

type QueryCloudSupplierJsonRespond struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List []Supplier `json:"list"`
	} `json:"data"`
}
