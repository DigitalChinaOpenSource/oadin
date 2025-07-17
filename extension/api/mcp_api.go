package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oadin/extension/dto"
	"oadin/extension/entity"
	"oadin/extension/rpc"
	"oadin/extension/server"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

type McpApi struct {
	McpService server.MCPService
	Playground server.Playground
}

// NewMcpApi creates a new instance of McpApi
func NewMcpApi() *McpApi {
	return &McpApi{
		McpService: server.NewMcpService(),
		Playground: server.NewPlayground(),
	}
}

func (e *McpApi) InjectRoutes(api *gin.RouterGroup) {
	api.POST("/search", e.GetMCPList)
	api.GET("/:id", e.GetMCPDetail)
	api.POST("/:id/tools/search", e.GetKits)
	api.GET("/:id/clients", e.GetClients)
	api.GET("/categories", e.GetCategories)
	api.PUT("/:id/download", e.DownloadMCP)
	api.POST("/mine", e.GetMyMCPList)
	api.PUT("/:id/auth", e.AuthorizeMCP)
	api.PUT("/:id/reverse", e.ReverseStatus)
	api.PUT("/setup", e.SetupFunTool)
	api.POST("/client/start", e.ClientMcpStart)
	api.POST("/client/stop", e.ClientMcpStop)
	api.POST("/client/getTools", e.ClientGetTools)
	api.POST("/client/runTool", e.ClientRunTool)
}

func (e *McpApi) GetMCPList(c *gin.Context) {
	var req rpc.MCPListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := e.McpService.GetMCPList(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务暂时不可用"})
		return
	}

	c.JSON(http.StatusOK, resp.Data)
}

func (e *McpApi) GetMCPDetail(c *gin.Context) {
	id := c.Param("id")
	resp, err := e.McpService.GetMCP(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (e *McpApi) GetKits(c *gin.Context) {
	id := c.Param("id")
	var req rpc.ToolSearchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := e.McpService.GetKits(c, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (e *McpApi) GetClients(c *gin.Context) {
	id := c.Param("id")
	resp, err := e.McpService.GetClients(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (e *McpApi) DownloadMCP(c *gin.Context) {
	id := c.Param("id")
	err := e.McpService.DownloadMCP(c, id)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (e *McpApi) AuthorizeMCP(c *gin.Context) {

	id := c.Param("id")

	auth := map[string]string{}

	err := c.BindJSON(&auth)
	authStr, err := json.Marshal(auth)

	err = e.McpService.AuthorizeMCP(c, id, string(authStr))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})
}

func (e *McpApi) GetCategories(c *gin.Context) {
	resp, err := e.McpService.GetCategories(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (e *McpApi) GetMyMCPList(c *gin.Context) {
	var req rpc.MCPListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := e.McpService.GetMyMCPList(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Data)

}

func (e *McpApi) ReverseStatus(c *gin.Context) {
	id := c.Param("id")
	err := e.McpService.ReverseStatus(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (e *McpApi) SetupFunTool(c *gin.Context) {
	var req rpc.SetupFunToolRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := e.McpService.SetupFunTool(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (e *McpApi) ClientMcpStart(c *gin.Context) {
	var req dto.ClientMcpStartRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	var message string
	var successIds []string
	var errorIds []string
	for _, id := range req.Ids {
		err := e.McpService.ClientMcpStart(c, id)
		if err != nil {
			message = message + fmt.Sprintf("MCP %s 启动失败: %s", id, err.Error())
			errorIds = append(errorIds, id)
		} else {
			message = message + fmt.Sprintf("MCP %s 启动成功", id)
			successIds = append(successIds, id)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": message, "data": map[string][]string{"successIds": successIds, "errorIds": errorIds}})
}

func (e *McpApi) ClientMcpStop(c *gin.Context) {
	var req dto.ClientMcpStopRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	err := e.McpService.ClientMcpStop(c, req.Ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "404", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": "success"})
}

func (e *McpApi) ClientRunTool(c *gin.Context) {
	var req dto.ClientRunToolRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	resp, err := e.McpService.ClientRunTool(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "500", "message": err.Error()})
		return
	}
	inputParamsStr := ""
	if req.ToolArgs != nil {
		if b, err := json.Marshal(map[string]interface{}{"name": req.ToolName, "arguments": req.ToolArgs}); err == nil {
			inputParamsStr = string(b)
		}
	}
	outputParamsStr := ""
	if resp != nil {
		if b, err := json.Marshal(struct {
			Content any  `json:"content"`
			IsError bool `json:"isError"`
		}{
			Content: resp.Content,
			IsError: resp.IsError,
		}); err == nil {
			outputParamsStr = string(b)
		}
	}
	toolMessage := &entity.ToolMessage{
		ID:           req.MessageId,
		McpId:        req.McpId,
		Name:         req.ToolName,
		InputParams:  inputParamsStr,
		OutputParams: outputParamsStr,
		Status:       !resp.IsError,
		Logo:         resp.Logo,
		Desc:         resp.ToolDesc,
	}
	e.Playground.UpdateToolCall(c, toolMessage)
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": resp})
}

func (e *McpApi) ClientGetTools(c *gin.Context) {
	var req dto.ClientGetToolsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}

	res := dto.ClientGetToolsResponse{}
	mcpTools := make([]dto.McpTool, 0, len(req.Ids))
	for _, id := range req.Ids {
		tools, err := e.McpService.ClientGetTools(c, id)
		if err != nil {
			continue
		}

		newTools := make([]dto.Tool, 0, len(tools))
		for _, tool := range tools {
			newTools = append(newTools, dto.Tool{Type: "function", Function: dto.TypeFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema}})
		}
		mcpTools = append(mcpTools, dto.McpTool{McpId: id, Tools: newTools})
	}
	res.Tools = mcpTools
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": res})
}
