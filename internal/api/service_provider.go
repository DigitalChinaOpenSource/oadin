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

package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"oadin/internal/api/dto"
	"oadin/internal/logger"
	"oadin/internal/utils/bcode"
)

func (t *OADINCoreServer) CreateServiceProvider(c *gin.Context) {
	logger.ApiLogger.Debug("[API] CreateServiceProvider request params:", c.Request.Body)
	request := new(dto.CreateServiceProviderRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrServiceProviderBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.ServiceProvider.CreateServiceProvider(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	logger.ApiLogger.Debug("[API] CreateServiceProvider response:", resp)
	c.JSON(http.StatusOK, resp)
}

func (t *OADINCoreServer) DeleteServiceProvider(c *gin.Context) {
	logger.ApiLogger.Debug("[API] DeleteServiceProvider request params:", c.Request.Body)
	request := new(dto.DeleteServiceProviderRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrServiceProviderBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.ServiceProvider.DeleteServiceProvider(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	logger.ApiLogger.Debug("[API] DeleteServiceProvider response:", resp)
	c.JSON(http.StatusOK, resp)
}

func (t *OADINCoreServer) UpdateServiceProvider(c *gin.Context) {
	logger.ApiLogger.Debug("[API] UpdateServiceProvider request params:", c.Request.Body)
	request := new(dto.UpdateServiceProviderRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrServiceProviderBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.ServiceProvider.UpdateServiceProvider(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	logger.ApiLogger.Debug("[API] UpdateServiceProvider response:", resp)
	c.JSON(http.StatusOK, resp)
}

func (t *OADINCoreServer) GetServiceProvider(c *gin.Context) {
	request := &dto.GetServiceProviderRequest{}
	if err := c.Bind(request); err != nil {
		if !errors.Is(err, io.EOF) {
			bcode.ReturnError(c, bcode.ErrServiceProviderBadRequest)
			return
		}
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.ServiceProvider.GetServiceProvider(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *OADINCoreServer) GetServiceProviders(c *gin.Context) {
	logger.ApiLogger.Debug("[API] GetServiceProviders request params:", c.Request.Body)
	request := &dto.GetServiceProvidersRequest{}
	if err := c.ShouldBindJSON(request); err != nil {
		if !errors.Is(err, io.EOF) {
			bcode.ReturnError(c, bcode.ErrServiceProviderBadRequest)
			return
		}
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.ServiceProvider.GetServiceProviders(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	logger.ApiLogger.Debug("[API] GetServiceProviders response:", resp)
	c.JSON(http.StatusOK, resp)
}
