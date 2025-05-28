package vega

import (
	"time"
)

type QueryCloudModelJsonRequest struct {
	DeployMode string `json:"deployMode"`
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
	Popularity     int      `json:"popularity"`
	OllamaId       string   `json:"ollamaId"`
	ParameterScale float32  `json:"parameterScale"`
	FileSize       string   `json:"fileSize"`
	MaxInput       string   `json:"maxInput"`
	MaxOutput      string   `json:"maxOutput"`
	CreateBy       string   `json:"createBy"`
	UpdateBy       string   `json:"updateBy"`
	CreateAt       int      `json:"createAt"`
	UpdateAt       int      `json:"updateAt"`
}

type QueryCloudSupplierJsonRequest struct {
	Page     int
	PageSize int
}

type Supplier struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	Flavor        string    `json:"flavor"`
	ServiceSource string    `json:"serviceSource"`
	ServerName    string    `json:"serverName"`
	Status        int       `json:"status"`
	ContextScale  string    `json:"contextScale"`
	Description   string    `json:"description"`
	RequestMethod string    `json:"requestMethod"`
	RequestUrl    string    `json:"requestUrl"`
	AuthType      string    `json:"authType"`
	AuthKey       string    `json:"authKey"`
	CreateBy      string    `json:"createBy"`
	UpdateBy      string    `json:"updateBy"`
	CreateAt      time.Time `json:"createAt"`
	UpdateAt      time.Time `json:"updateAt"`
}

type QueryCloudModelJsonRespond struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List  []Model `json:"list"`
		Total int     `json:"total"`
	} `json:"data"`
}

type QueryCloudSupplierJsonRespond struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List  []Supplier `json:"list"`
		Total int        `json:"total"`
	} `json:"data"`
}
