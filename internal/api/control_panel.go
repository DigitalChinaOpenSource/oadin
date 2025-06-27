package api

import (
	"errors"
	"io"
	"net/http"

	"oadin/internal/api/dto"
	"oadin/internal/server"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

func (t *OadinCoreServer) GetModelFilePathHandler(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := server.GetModelFilePath(ctx)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *OadinCoreServer) GetPathDiskSizeHandler(c *gin.Context) {
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

func (t *OadinCoreServer) ModifyModelFilePathHandler(c *gin.Context) {
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
