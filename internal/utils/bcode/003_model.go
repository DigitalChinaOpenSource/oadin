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

package bcode

import "net/http"

var (
	ModelCode = NewBcode(http.StatusOK, 30000, "service interface call success")

	ErrModelBadRequest = NewBcode(http.StatusBadRequest, 30001, " bad request")

	ErrModelIsExist = NewBcode(http.StatusConflict, 30002, "provider model already exist")

	ErrModelRecordNotFound = NewBcode(http.StatusNotFound, 30003, "model not exist")

	ErrAddModel = NewBcode(http.StatusInternalServerError, 30004, "model insert db failed")

	ErrDeleteModel = NewBcode(http.StatusInternalServerError, 30005, "model delete db failed")

	ErrEngineDeleteModel = NewBcode(http.StatusInternalServerError, 30006, "engine delete model failed")

	ErrNoRecommendModel = NewBcode(http.StatusNotFound, 30007, "No Recommend Model")
)
