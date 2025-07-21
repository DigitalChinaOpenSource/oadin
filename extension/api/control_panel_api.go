package api

import (
	"errors"
	"io"
	"net/http"

	"oadin/extension/dto"
	"oadin/extension/server"
	"oadin/extension/utils/bcode"

	"github.com/gin-gonic/gin"
)

type ControlPanelApi struct {
}

func NewControlPanelApi() *ControlPanelApi {
	return &ControlPanelApi{}
}

func (e *ControlPanelApi) InjectRoutes(api *gin.RouterGroup) {
	api.GET("/model/filepath", e.GetModelFilePathHandler)
	api.POST("/model/filepath", e.ModifyModelFilePathHandler)
	api.GET("/path/space", e.GetPathDiskSizeHandler)
	api.GET("/model/square", e.GetSupportModelListCombine)
}

func (e *ControlPanelApi) GetModelFilePathHandler(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := server.GetModelFilePath(ctx)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (e *ControlPanelApi) ModifyModelFilePathHandler(c *gin.Context) {
	request := new(dto.ModifyModelFilePathRequest)
	if err := c.Bind(request); err != nil {
		if !errors.Is(err, io.EOF) {
			bcode.ReturnError(c, bcode.ErrModelBadRequest)
			return
		}
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := server.ModifyModelFilePath(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (e *ControlPanelApi) GetPathDiskSizeHandler(c *gin.Context) {
	request := new(dto.GetPathDiskSizeInfoRequest)
	if err := c.Bind(request); err != nil {
		if !errors.Is(err, io.EOF) {
			bcode.ReturnError(c, bcode.ErrModelBadRequest)
			return
		}
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := server.GetFilePathSize(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (e *ControlPanelApi) GetSupportModelListCombine(c *gin.Context) {
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
