package vega

import "byze/internal/utils/bcode"

type QueryCloudModelJsonRequest struct {
	HybridPolicy string
}

type Chat struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Avatar      string  `json:"avatar"`
	Description string  `json:"description"`
	Class       string  `json:"class"`
	Provider    string  `json:"provider"`
	Size        string  `json:"size"`
	ParamsSize  float64 `json:"params_size"`
}

type Embedding struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Class       string `json:"class"`
	Provider    string `json:"provider"`
	Size        string `json:"size"`
}

type TextToImage struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	Class       string `json:"class"`
	Provider    string `json:"provider"`
	Size        string `json:"size"`
}

type QueryCloudModelJsonData struct {
	Chat        []Chat      `json:"chat"`
	Embedding   []Embedding `json:"embedding"`
	TextToImage TextToImage `json:"text_to_image"`
}

type QueryCloudModelJsonRespond struct {
	bcode.Bcode
	Data *QueryCloudModelJsonData `json:"data"`
}
