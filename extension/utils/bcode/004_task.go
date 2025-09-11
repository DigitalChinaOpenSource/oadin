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
	TaskCode = NewBcode(http.StatusOK, 40100, "Service interface call success")

	// Client Error Codes (4xx) - These are returned to users
	ErrNoTargetProvider            = NewBcode(http.StatusNotFound, 40101, "There is no available target provider for the request")
	ErrReadRequestBody             = NewBcode(http.StatusBadRequest, 40102, "Failed to read request body")
	ErrUnmarshalRequestBody        = NewBcode(http.StatusBadRequest, 40103, "Failed to unmarshal request body")
	ErrUnSupportContentType        = NewBcode(http.StatusUnsupportedMediaType, 40104, "Unsupported content type")
	ErrUnSupportRequestMethod      = NewBcode(http.StatusMethodNotAllowed, 40105, "Unsupported request method")
	ErrUnsupportedCloseNotifier    = NewBcode(http.StatusNotImplemented, 40106, "Unsupported CloseNotifier")
	ErrUnsupportedFlusher          = NewBcode(http.StatusNotImplemented, 40107, "Unsupported Flusher")
	ErrNotExistDefaultProvider     = NewBcode(http.StatusNotFound, 40108, "The default provider does not exist")
	ErrModelUnDownloaded           = NewBcode(http.StatusNotFound, 40109, "The model has not been downloaded yet")
	ErrProviderNotExist            = NewBcode(http.StatusNotFound, 40110, "The provider does not exist")
	ErrUnmarshalProviderProperties = NewBcode(http.StatusInternalServerError, 40111, "Failed to unmarshal provider properties")
	ErrMiddlewareHandle            = NewBcode(http.StatusInternalServerError, 40112, "Middleware handle error")
	ErrFlavorConvertRequest        = NewBcode(http.StatusUnprocessableEntity, 40113, "Flavor convert request error")
	ErrFlavorConvertResponse       = NewBcode(http.StatusInternalServerError, 40114, "Flavor convert response error")
	ErrReadResponseBody            = NewBcode(http.StatusInternalServerError, 40115, "Failed to read response body")
	ErrReadResponseChunk           = NewBcode(http.StatusInternalServerError, 40116, "Failed to read response chunk")
	ErrInvokeServiceProvider       = NewBcode(http.StatusBadGateway, 40117, "Failed to invoke service provider")

	// New error codes
	// Service-related errors
	ErrUnsupportedServiceType = NewBcode(http.StatusNotImplemented, 40130, "Unsupported service type")
	ErrPrepareRequest         = NewBcode(http.StatusBadRequest, 40131, "Failed to prepare request")
	ErrSendRequest            = NewBcode(http.StatusBadGateway, 40132, "Failed to send request to service")
	ErrReceiveResponse        = NewBcode(http.StatusBadGateway, 40133, "Failed to receive response from service")

	// WebSocket-related errors
	ErrWebSocketUpgradeFailed   = NewBcode(http.StatusBadRequest, 40140, "Failed to upgrade connection to WebSocket")
	ErrMissingWebSocketConnID   = NewBcode(http.StatusBadRequest, 40141, "Missing WebSocket connection ID")
	ErrWebSocketMessageFormat   = NewBcode(http.StatusBadRequest, 40142, "Unrecognized WebSocket message format")
	ErrWebSocketSendMessage     = NewBcode(http.StatusInternalServerError, 40143, "Failed to send message to WebSocket client")
	ErrWebSocketSessionCreation = NewBcode(http.StatusInternalServerError, 40144, "Failed to create WebSocket session")

	// Authentication-related errors
	ErrAuthInfoParsing      = NewBcode(http.StatusBadRequest, 40150, "Failed to parse authentication information")
	ErrAuthenticationFailed = NewBcode(http.StatusUnauthorized, 40151, "Authentication failed")

	// Data processing-related errors
	ErrJSONParsing         = NewBcode(http.StatusBadRequest, 40160, "Failed to parse JSON data")
	ErrParameterValidation = NewBcode(http.StatusBadRequest, 40161, "Parameter validation failed")

	// Task processing-related errors
	ErrUnknownTaskType = NewBcode(http.StatusBadRequest, 40170, "Unknown task type")
	ErrTaskProcessing  = NewBcode(http.StatusInternalServerError, 40171, "Task processing failed")

	// GRPC-related errors
	ErrGRPCStreamSend    = NewBcode(http.StatusBadGateway, 40180, "Failed to send data to GRPC stream")
	ErrGRPCStreamReceive = NewBcode(http.StatusBadGateway, 40181, "Failed to receive data from GRPC stream")
	ErrGRPCConnection    = NewBcode(http.StatusBadGateway, 40182, "Failed to establish GRPC connection")
)
