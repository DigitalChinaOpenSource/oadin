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
	TaskCode = NewBcode(http.StatusOK, 40000, "Service interface call success")

	// Client Error Codes (4xx) - These are returned to users
	ErrNoTargetProvider            = NewBcode(http.StatusNotFound, 40001, "There is no available target provider for the request")
	ErrReadRequestBody             = NewBcode(http.StatusBadRequest, 40002, "Failed to read request body")
	ErrUnmarshalRequestBody        = NewBcode(http.StatusBadRequest, 40003, "Failed to unmarshal request body")
	ErrUnSupportContentType        = NewBcode(http.StatusUnsupportedMediaType, 40004, "Unsupported content type")
	ErrUnSupportRequestMethod      = NewBcode(http.StatusMethodNotAllowed, 40005, "Unsupported request method")
	ErrUnsupportedCloseNotifier    = NewBcode(http.StatusNotImplemented, 40006, "Unsupported CloseNotifier")
	ErrUnsupportedFlusher          = NewBcode(http.StatusNotImplemented, 40007, "Unsupported Flusher")
	ErrNotExistDefaultProvider     = NewBcode(http.StatusNotFound, 40008, "The default provider does not exist")
	ErrModelUnDownloaded           = NewBcode(http.StatusNotFound, 40009, "The model has not been downloaded yet")
	ErrProviderNotExist            = NewBcode(http.StatusNotFound, 40010, "The provider does not exist")
	ErrUnmarshalProviderProperties = NewBcode(http.StatusInternalServerError, 40011, "Failed to unmarshal provider properties")
	ErrMiddlewareHandle            = NewBcode(http.StatusInternalServerError, 40012, "Middleware handle error")
	ErrFlavorConvertRequest        = NewBcode(http.StatusUnprocessableEntity, 40013, "Flavor convert request error")
	ErrFlavorConvertResponse       = NewBcode(http.StatusInternalServerError, 40014, "Flavor convert response error")
	ErrReadResponseBody            = NewBcode(http.StatusInternalServerError, 40015, "Failed to read response body")
	ErrReadResponseChunk           = NewBcode(http.StatusInternalServerError, 40016, "Failed to read response chunk")
	ErrInvokeServiceProvider       = NewBcode(http.StatusBadGateway, 40017, "Failed to invoke service provider")

	// New error codes
	// Service-related errors
	ErrUnsupportedServiceType = NewBcode(http.StatusNotImplemented, 40030, "Unsupported service type")
	ErrPrepareRequest         = NewBcode(http.StatusBadRequest, 40031, "Failed to prepare request")
	ErrSendRequest            = NewBcode(http.StatusBadGateway, 40032, "Failed to send request to service")
	ErrReceiveResponse        = NewBcode(http.StatusBadGateway, 40033, "Failed to receive response from service")

	// WebSocket-related errors
	ErrWebSocketUpgradeFailed   = NewBcode(http.StatusBadRequest, 40040, "Failed to upgrade connection to WebSocket")
	ErrMissingWebSocketConnID   = NewBcode(http.StatusBadRequest, 40041, "Missing WebSocket connection ID")
	ErrWebSocketMessageFormat   = NewBcode(http.StatusBadRequest, 40042, "Unrecognized WebSocket message format")
	ErrWebSocketSendMessage     = NewBcode(http.StatusInternalServerError, 40043, "Failed to send message to WebSocket client")
	ErrWebSocketSessionCreation = NewBcode(http.StatusInternalServerError, 40044, "Failed to create WebSocket session")

	// Authentication-related errors
	ErrAuthInfoParsing      = NewBcode(http.StatusBadRequest, 40050, "Failed to parse authentication information")
	ErrAuthenticationFailed = NewBcode(http.StatusUnauthorized, 40051, "Authentication failed")

	// Data processing-related errors
	ErrJSONParsing         = NewBcode(http.StatusBadRequest, 40060, "Failed to parse JSON data")
	ErrParameterValidation = NewBcode(http.StatusBadRequest, 40061, "Parameter validation failed")

	// Task processing-related errors
	ErrUnknownTaskType = NewBcode(http.StatusBadRequest, 40070, "Unknown task type")
	ErrTaskProcessing  = NewBcode(http.StatusInternalServerError, 40071, "Task processing failed")

	// GRPC-related errors
	ErrGRPCStreamSend    = NewBcode(http.StatusBadGateway, 40080, "Failed to send data to GRPC stream")
	ErrGRPCStreamReceive = NewBcode(http.StatusBadGateway, 40081, "Failed to receive data from GRPC stream")
	ErrGRPCConnection    = NewBcode(http.StatusBadGateway, 40082, "Failed to establish GRPC connection")
)
