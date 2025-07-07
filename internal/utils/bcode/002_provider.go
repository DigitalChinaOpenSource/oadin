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
	ServiceProviderCode = NewBcode(http.StatusOK, 20000, "service provider interface call success")

	ErrServiceProviderBadRequest = NewBcode(http.StatusBadRequest, 20001, "bad request")

	ErrProviderInvalid = NewBcode(http.StatusBadRequest, 20002, "provider invalid")

	ErrProviderIsUnavailable = NewBcode(http.StatusBadRequest, 20003, "service provider is unavailable")

	ErrProviderModelEmpty = NewBcode(http.StatusBadRequest, 20004, "provider model empty")

	ErrProviderUpdateFailed = NewBcode(http.StatusInternalServerError, 20005, "provider update failed")

	ErrProviderAuthInfoLost = NewBcode(http.StatusBadRequest, 20006, "provider api auth info lost")

	ErrProviderServiceUrlNotFormat = NewBcode(http.StatusBadRequest, 20007, "provider service url is irregular")
)
