package types

import "time"

// SupportModel  table structure
type SupportModel struct {
	Id            string    `json:"id"`
	OllamaId      string    `json:"Ollama_id"`
	Name          string    `json:"name"`
	Avatar        string    `json:"avatar"`
	Description   string    `json:"description"`
	Class         []string  `json:"class"`
	Flavor        string    `json:"flavor"`
	ApiFlavor     string    `json:"api_flavor"`
	Size          string    `json:"size"`
	ParamSize     float32   `json:"params_size"`
	InputLength   int       `json:"input_length"`
	OutputLength  int       `json:"output_length"`
	ServiceSource string    `json:"service_source"`
	ServiceName   string    `json:"service_name"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
	Think         bool      `json:"think"`
	ThinkSwitch   bool      `json:"think_switch"`
	Tools         bool      `json:"tools"` // 是否支持工具调用
}

func (s *SupportModel) TableName() string {
	return "support_model"
}

func (s *SupportModel) SetCreateTime(time time.Time) {
	s.CreatedAt = time
}

func (s *SupportModel) SetUpdateTime(time time.Time) {
	s.UpdatedAt = time
}

func (s *SupportModel) PrimaryKey() string {
	return "name"
}

func (s *SupportModel) Index() map[string]interface{} {
	index := make(map[string]interface{})
	if s.Name != "" {
		index["name"] = s.Name
	}
	if s.Id != "" {
		index["id"] = s.Id
	}

	return index
}
