package api

import (
	"net/http"

	"oadin/extension/server"
	"oadin/extension/dto"
	"oadin/internal/utils/bcode"
	"github.com/gin-gonic/gin"
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
