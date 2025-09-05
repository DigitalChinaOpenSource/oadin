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

package version

import "log/slog"

const OADINVersion = "2.0.0"

const OADINSpecVersion = "v0.4"

const OADINIcon = ""

const OADINName = "OADIN"

const OADINDescription = "OADIN (AIPC Open Gateway) aims to decouple AI applications on AI PCs from the AI services they rely on. It is designed to provide developers with an extremely simple and easy-to-use infrastructure to install local AI services in their development environments and publish their AI applications without packaging their own AI stacks and models."

var OadinSubVersion = "2.0.5"

// GetOADINVersion slog打印奥丁版本信息
func GetOADINVersion() {
	slog.Info("OADIN Version: " + OADINVersion)
	slog.Info("OADIN Spec Version: " + OADINSpecVersion)
	slog.Info("OADIN Sub Version: " + OadinSubVersion)
	slog.Info("OADIN Name: " + OADINName)
}
