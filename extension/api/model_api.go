package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/internal/utils/bcode"
)

type ModelApi struct {
	Playground server.Playground
	MCP        server.MCPService
}

func NewModelApi() *ModelApi {
	return &ModelApi{}
}

func (e *ModelApi) InjectRoutes(api *gin.RouterGroup) {
	api.GET("/support/smartvision", e.GetSmartVisionSupportModelList)
}

func (e *ModelApi) GetSmartVisionSupportModelList(c *gin.Context) {
	request := new(dto.SmartVisionSupportModelRequest)
	if err := c.Bind(request); err != nil {
		bcode.ReturnError(c, bcode.ErrModelBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		bcode.ReturnError(c, err)
		return
	}

	ctx := c.Request.Context()
	data, err := server.GetSupportSmartVisionModels(ctx, request)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, data)
}
