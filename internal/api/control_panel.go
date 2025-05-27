package api

import (
	"byze/internal/api/dto"
	"byze/internal/server"
	"byze/internal/utils/bcode"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func (t *ByzeCoreServer) GetModelFilePathHandler(c *gin.Context) {
	ctx := c.Request.Context()
	resp, err := server.GetModelFilePath(ctx)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *ByzeCoreServer) GetPathDiskSizeHandler(c *gin.Context) {
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

func (t *ByzeCoreServer) ModifyModelFilePathHandler(c *gin.Context) {
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

func (t *ByzeCoreServer) GetAllModels(c *gin.Context) {
	//ctx := c.Request.Context()

}
