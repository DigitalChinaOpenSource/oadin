package api

import (
	"errors"
	"io"
	"net/http"

	"oadin/internal/api/dto"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

func (t *OadinCoreServer) CreateServiceProvider(c *gin.Context) {
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

	c.JSON(http.StatusOK, resp)
}

func (t *OadinCoreServer) DeleteServiceProvider(c *gin.Context) {
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

	c.JSON(http.StatusOK, resp)
}

func (t *OadinCoreServer) UpdateServiceProvider(c *gin.Context) {
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

	c.JSON(http.StatusOK, resp)
}

func (t *OadinCoreServer) GetServiceProvider(c *gin.Context) {
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

func (t *OadinCoreServer) GetServiceProviders(c *gin.Context) {
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

	c.JSON(http.StatusOK, resp)
}
