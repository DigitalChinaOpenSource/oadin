package vega

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
type QueryCloudModelJsonRespond struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		List  []Model `json:"list"`
		Total int     `json:"total"`
	} `json:"data"`
}
type Service struct {
	Id             string `json:"id"`
	Name           string `json:"name"`
	HybridPolicy   string `json:"hybridPolicy"`
	RemoteProvider string `json:"remoteProvider"`
	LocalProvider  string `json:"localProvider"`
	Status         int    `json:"status"`
	CreateBy       string `json:"createBy"`
	UpdateBy       string `json:"updateBy"`
	CreateAt       int    `json:"createAt"`
	UpdateAt       int    `json:"updateAt"`
}

type ServiceProvider struct {
	Id            string   `json:"id"`
	ProviderName  string   `json:"providerName"`
	ServiceName   string   `json:"serviceName"`
	ServiceSource string   `json:"serviceSource"`
	Desc          string   `json:"desc"`
	ApiFlavor     string   `json:"apiFlavor"`
	Method        string   `json:"method"`
	AuthType      string   `json:"authType"`
	AuthKey       string   `json:"authKey"`
	Models        []string `json:"models"`
	Status        int      `json:"status"`
	CreateBy      string   `json:"createBy"`
	UpdateBy      string   `json:"updateBy"`
	CreateAt      int      `json:"createAt"`
	UpdateAt      int      `json:"updateAt"`
}

type QueryCloudSupplierJsonRespond struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Service         []Service         `json:"service"`
		ServiceProvider []ServiceProvider `json:"serviceProvider"`
		Version         string            `json:"version"`
	} `json:"data"`
}
