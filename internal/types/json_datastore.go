//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

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
	MaxInput      int       `json:"max_input"`
	MaxOutput     int       `json:"max_output"`
	ServiceSource string    `json:"service_source"`
	ServiceName   string    `json:"service_name"`
	CreatedAt     time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP" json:"updated_at"`
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

	return index
}
