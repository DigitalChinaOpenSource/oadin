package response

type AboutUsResponse struct {
	Name        string `json:"name"`
	EnName      string `json:"enName"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Logo        string `json:"logo"`

	// 官方网站
	OfficialWebsite string `json:"officialWebsite"`
	// 版权声明
	Copyright string `json:"copyright"`
}

type ChangeListResponse struct {
}
