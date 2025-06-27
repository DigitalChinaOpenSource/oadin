package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"oadin/internal/rpc"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"

	"github.com/gin-gonic/gin"
)

// GetMCPList
// @Summary 获取MCP列表
// @Description 根据关键词和标签检索MCP
// @Tags MCP
// @Param keyword query string false "搜索关键词"
// @Param category query []string false "分类列表"
// @Param page query int 1 "页码"
// @Param size query int 10 "每页数量"
// @Success 200 {object} models.MCPListResponse
// @Router /api/mcp [get]
func (t *OadinCoreServer) GetMCPList(c *gin.Context) {
	var req rpc.MCPListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := t.MCP.GetMCPList(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务暂时不可用"})
		return
	}

	c.JSON(http.StatusOK, resp.Data)
}

// GetMCPDetail @Summary 获取MCP详情
// @Description 根据ID获取MCP详细信息
// @Tags MCP
// @Param id path int true "MCP ID"
// @Success 200 {object} models.MCPDetailResponse
// @Router /api/mcp/{id} [get]
func (t *OadinCoreServer) GetMCPDetail(c *gin.Context) {
	id := c.Param("id")
	resp, err := t.MCP.GetMCP(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *OadinCoreServer) GetKits(c *gin.Context) {
	id := c.Param("id")
	var req rpc.ToolSearchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := t.MCP.GetKits(c, id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *OadinCoreServer) GetClients(c *gin.Context) {
	id := c.Param("id")
	resp, err := t.MCP.GetClients(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *OadinCoreServer) DownloadMCP(c *gin.Context) {
	id := c.Param("id")
	err := t.MCP.DownloadMCP(c, id)
	if err != nil {
		bcode.ReturnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (t *OadinCoreServer) AuthorizeMCP(c *gin.Context) {

	id := c.Param("id")

	auth := map[string]string{}

	err := c.BindJSON(&auth)
	authStr, err := json.Marshal(auth)

	err = t.MCP.AuthorizeMCP(c, id, string(authStr))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})
}

func (t *OadinCoreServer) GetCategories(c *gin.Context) {
	resp, err := t.MCP.GetCategories(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *OadinCoreServer) GetMyMCPList(c *gin.Context) {
	var req rpc.MCPListRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := t.MCP.GetMyMCPList(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)

}

func (t *OadinCoreServer) ReverseStatus(c *gin.Context) {
	id := c.Param("id")
	err := t.MCP.ReverseStatus(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (t *OadinCoreServer) SetupFunTool(c *gin.Context) {
	var req rpc.SetupFunToolRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := t.MCP.SetupFunTool(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})

}

func (t *OadinCoreServer) ClientMcpStart(c *gin.Context) {
	var req types.ClientMcpStartRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	var message string
	for _, id := range req.Ids {
		err := t.MCP.ClientMcpStart(c, id)
		if err != nil {
			message = message + fmt.Sprintf("MCP %s 启动失败: %s", id, err.Error())
		} else {
			message = message + fmt.Sprintf("MCP %s 启动成功", id)
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": "success", "data": message})
}

func (t *OadinCoreServer) ClientMcpStop(c *gin.Context) {
	var req types.ClientMcpStopRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	err := t.MCP.ClientMcpStop(c, req.Ids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "404", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "message": "success"})
}

func (t *OadinCoreServer) ClientRunTool(c *gin.Context) {
	var req types.ClientRunToolRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}
	resp, err := t.MCP.ClientRunTool(c, &req)
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
		if b, err := json.Marshal(resp.CallToolResult); err == nil {
			outputParamsStr = string(b)
		}
	}
	toolMessage := &types.ToolMessage{
		ID:           req.MessageId,
		McpId:        req.McpId,
		Name:         req.ToolName,
		InputParams:  inputParamsStr,
		OutputParams: outputParamsStr,
		Status:       !resp.IsError,
		Logo:         resp.Logo,
		Desc:         resp.ToolDesc,
	}
	t.Playground.UpdateToolCall(c, toolMessage)
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": resp})
}

func (t *OadinCoreServer) ClientGetTools(c *gin.Context) {
	var req types.ClientGetToolsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "400", "message": err.Error()})
		return
	}

	res := types.ClientGetToolsResponse{}
	mcpTools := make([]types.McpTool, 0, len(req.Ids))
	for _, id := range req.Ids {
		tools, err := t.MCP.ClientGetTools(c, id)
		if err != nil {
			continue
		}

		newTools := make([]types.Tool, 0, len(tools))
		for _, tool := range tools {
			newTools = append(newTools, types.Tool{Type: "function", Function: types.TypeFunction{Name: tool.Name, Description: tool.Description, Parameters: tool.InputSchema}})
		}
		mcpTools = append(mcpTools, types.McpTool{McpId: id, Tools: newTools})
	}
	res.Tools = mcpTools
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": res})
}

func (t *OadinCoreServer) ClientMAC(c *gin.Context) {
	err := t.MCP.ClientMAC(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "500", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "200", "data": nil})
}
