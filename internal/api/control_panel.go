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
	// "errors"
	// "io"
	"net/http"

	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/server"
	"intel.com/aog/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

func (t *AOGCoreServer) GetDashBoardHandler(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := server.GetDashboard(ctx)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (t *AOGCoreServer) GetSupportModelListCombine(c *gin.Context) {
	request := new(dto.GetSupportModelRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := server.GetSupportModelListCombine(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (t *AOGCoreServer) SetDefaultModelHandler(c *gin.Context) {
	request := new(dto.SetDefaultModelRequest)
	if err := c.ShouldBindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}
	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}
	ctx := c.Request.Context()
	err := server.SetDefaultModel(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (t *AOGCoreServer) GetProductInfoHandler(c *gin.Context) {
	resp, err := server.GetProductInfo(c.Request.Context())
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (t *AOGCoreServer) GetModelkeyHandler(c *gin.Context) {
	request := new(dto.GetModelkeyRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := server.GetModelkey(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
