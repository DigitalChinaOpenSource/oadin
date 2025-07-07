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

package constants

// Application information
const (
	AppName           = "aog"
	BaseDownloadURL   = "https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/aog"
	UrlDirPathWindows = "/windows"
)

// model related
const (
	RecommendModel           = "deepseek-r1:7b"
	DefaultChatModelName     = "deepseek-r1:7b"
	DefaultEmbedModelName    = "bge-m3"
	DefaultTextToImageModel  = "OpenVINO/stable-diffusion-v1-5-fp16-ov"
	DefaultSpeechToTextModel = "NamoLi/whisper-large-v3-ov"
)

// Network related
const (
	DefaultHTTPPort   = "16688"
	DefaultHost       = "127.0.0.1"
	DefaultHTTPSPort  = "443"
	DefaultHTTPPort80 = "80"
)

// Data format constants
const (
	// Data size units
	Byte = 1

	// Decimal units
	Thousand = 1000
	Million  = Thousand * 1000
	Billion  = Million * 1000

	// Byte units (decimal)
	KiloByte = Byte * 1000
	MegaByte = KiloByte * 1000
	GigaByte = MegaByte * 1000
	TeraByte = GigaByte * 1000

	// Byte units (binary)
	KibiByte = Byte * 1024
	MebiByte = KibiByte * 1024
	GibiByte = MebiByte * 1024
)
