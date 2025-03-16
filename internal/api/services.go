package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"intel.com/aog/internal/api/dto"
	"intel.com/aog/internal/utils/bcode"
)

func (t *AOGCoreServer) CreateAIGCService(c *gin.Context) {
	request := new(dto.CreateAIGCServiceRequest)
	if err := c.BindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrAIGCServiceBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.AIGCService.CreateAIGCService(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *AOGCoreServer) UpdateAIGCService(c *gin.Context) {
	request := new(dto.UpdateAIGCServiceRequest)
	if err := c.BindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrAIGCServiceBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.AIGCService.UpdateAIGCService(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *AOGCoreServer) GetAIGCService(c *gin.Context) {
}

func (t *AOGCoreServer) GetAIGCServices(c *gin.Context) {
	request := new(dto.GetAIGCServicesRequest)
	if err := c.BindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrAIGCServiceBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.AIGCService.GetAIGCServices(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *AOGCoreServer) ExportService(c *gin.Context) {
	request := new(dto.ExportServiceRequest)
	if err := c.BindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrAIGCServiceBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.AIGCService.ExportService(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (t *AOGCoreServer) ImportService(c *gin.Context) {
	request := new(dto.ImportServiceRequest)
	if err := c.BindJSON(request); err != nil {
		bcode.ReturnError(c, bcode.ErrAIGCServiceBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	resp, err := t.AIGCService.ImportService(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
