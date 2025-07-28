package api

import (
	"github.com/gin-gonic/gin"
	"oadin/extension/api/dto"
	"oadin/extension/server"
	"oadin/extension/utils/bcode"
)

type EngineApi struct {
	EngineManageService server.EngineManageService
}

func NewEngineApi() *EngineApi {
	return &EngineApi{
		EngineManageService: server.NewEngineManageService(),
	}
}

func (e *EngineApi) InjectRoutes(api *gin.RouterGroup) {
	api.GET("/exist", e.exist)
}

// exist 检查引擎是否存在
func (e *EngineApi) exist(c *gin.Context) {
	req := dto.EngineManageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidFailure(c, err.Error())
		return

	}
	// 检查引擎的安装状态
	exist := e.EngineManageService.Exist(c, req)

	dto.Success(c, exist)
}

// 执行引擎下载逻辑
func (e *EngineApi) install(c *gin.Context) {
	req := dto.EngineManageRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ValidFailure(c, err.Error())
		return
	}

	// 执行安装逻辑
	err := e.EngineManageService.Install(c, req)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}

	dto.Success(c, "引擎安装成功")
}
