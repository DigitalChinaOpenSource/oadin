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

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"intel.com/aog/internal/datastore"
	"intel.com/aog/internal/logger"
)

const (
	// Common HTTP status codes
	HTTPStatusOK                  = 200
	HTTPStatusBadRequest          = 400
	HTTPStatusUnauthorized        = 401
	HTTPStatusForbidden           = 403
	HTTPStatusNotFound            = 404
	HTTPStatusInternalServerError = 500
	HTTPStatusBadGateway          = 502
)

// Error Code of AOG contains 5 digits, the first 3 digits should be reversed and indicates the category of concept
// the last two digits indicates the error number
// For example, business code 11001 should split to 110 and 01, it means the code belongs to the 011 category env, and it's the 01 number error.

// SuccessCode a success code
var SuccessCode = NewBcode(HTTPStatusOK, HTTPStatusOK, "success")

// ErrServer an unexpected mistake.
var ErrServer = NewBcode(HTTPStatusInternalServerError, HTTPStatusInternalServerError, "The service has lapsed.")

// ErrForbidden check user perms failure
var ErrForbidden = NewBcode(HTTPStatusForbidden, HTTPStatusForbidden, "403 Forbidden")

// ErrUnauthorized check user auth failure
var ErrUnauthorized = NewBcode(HTTPStatusUnauthorized, HTTPStatusUnauthorized, "401 Unauthorized")

// ErrNotFound the request resource is not found
var ErrNotFound = NewBcode(HTTPStatusNotFound, HTTPStatusNotFound, "404 Not Found")

// ErrUpstreamNotFound the proxy upstream is not found
var ErrUpstreamNotFound = NewBcode(HTTPStatusBadGateway, HTTPStatusBadGateway, "Upstream not found")

// Bcode business error code
type Bcode struct {
	HTTPCode     int32  `json:"-"`
	BusinessCode int32  `json:"business_code"`
	Message      string `json:"message"`
}

func (b *Bcode) Error() string {
	switch {
	case b.Message != "":
		return b.Message
	default:
		return "something went wrong, please see the aog server logs for details"
	}
}

// SetMessage set new message and return a new error instance
func (b *Bcode) SetMessage(message string) *Bcode {
	return &Bcode{
		HTTPCode:     b.HTTPCode,
		BusinessCode: b.BusinessCode,
		Message:      message,
	}
}

var bcodeMap map[int32]*Bcode

// NewBcode new error code
func NewBcode(httpCode, businessCode int32, message string) *Bcode {
	if bcodeMap == nil {
		bcodeMap = make(map[int32]*Bcode)
	}
	if _, exit := bcodeMap[businessCode]; exit {
		panic("error business code is exist")
	}
	bcode := &Bcode{HTTPCode: httpCode, BusinessCode: businessCode, Message: message}
	bcodeMap[businessCode] = bcode
	return bcode
}

// ReturnHTTPError Unified handling of all types of errors, generating a standard return structure.
func ReturnHTTPError(c *gin.Context, err error) {
	c.SetAccepted(gin.MIMEJSON)
	ReturnError(c, err)
}

// ReturnError Unified handling of all types of errors, generating a standard return structure.
func ReturnError(c *gin.Context, err error) {
	var bcode *Bcode
	if errors.As(err, &bcode) {
		c.JSON(int(bcode.HTTPCode), err)
		return
	}

	if errors.Is(err, datastore.ErrRecordNotExist) {
		c.JSON(http.StatusNotFound, err)
		return
	}

	var validErr validator.ValidationErrors
	if errors.As(err, &validErr) {
		c.JSON(http.StatusBadRequest, Bcode{
			HTTPCode:     http.StatusBadRequest,
			BusinessCode: HTTPStatusBadRequest,
			Message:      err.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Bcode{
		HTTPCode:     http.StatusInternalServerError,
		BusinessCode: HTTPStatusInternalServerError,
		Message:      err.Error(),
	})
}

// WrapError wraps a Bcode error with the original error's message
// This preserves the error code while providing more context
// It also prevents error nesting if the original error is already a Bcode
func WrapError(bcodeErr *Bcode, originalErr error) error {
	if originalErr == nil {
		return bcodeErr
	}

	// Check if originalErr is already a Bcode error
	var existingBcode *Bcode
	if errors.As(originalErr, &existingBcode) {
		// Error is already a Bcode, don't nest errors
		return originalErr
	}

	return fmt.Errorf("%w: %v", bcodeErr, originalErr)
}

// LogAndReturnError logs the detailed error and returns it
// This is useful for server errors that should be logged but also returned to the client
func LogAndReturnError(bcodeErr *Bcode, originalErr error, logFields ...interface{}) error {
	if originalErr != nil {
		logger.LogicLogger.Error(bcodeErr.Message, append([]interface{}{"error", originalErr}, logFields...)...)
	} else {
		logger.LogicLogger.Error(bcodeErr.Message, logFields...)
	}
	// Return the wrapped error so the client gets the context
	return WrapError(bcodeErr, originalErr)
}
