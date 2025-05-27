package api

import (
	"byze/internal/rpc"
	"github.com/gin-gonic/gin"
	"net/http"
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
func (t *ByzeCoreServer) GetMCPList(c *gin.Context) {
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
func (t *ByzeCoreServer) GetMCPDetail(c *gin.Context) {
	id := c.Param("id")
	resp, err := t.MCP.GetMCP(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *ByzeCoreServer) GetKits(c *gin.Context) {
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

func (t *ByzeCoreServer) GetClients(c *gin.Context) {
	id := c.Param("id")
	resp, err := t.MCP.GetClients(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *ByzeCoreServer) DownloadMCP(c *gin.Context) {
	id := c.Param("id")
	resp, err := t.MCP.DownloadMCP(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp.Data)

}

func (t *ByzeCoreServer) AuthorizeMCP(c *gin.Context) {

	id := c.Param("id")

	auth := ""
	err := c.BindJSON("auth")
	resp, err := t.MCP.AuthorizeMCP(c, id, auth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *ByzeCoreServer) GetCategories(c *gin.Context) {
	resp, err := t.MCP.GetCategories(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, resp.Data)
}

func (t *ByzeCoreServer) GetMyMCPList(c *gin.Context) {
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

func (t *ByzeCoreServer) ReverseStatus(c *gin.Context) {
	id := c.Param("id")
	err := t.MCP.ReverseStatus(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, nil)

}

func (t *ByzeCoreServer) SetupFunTool(c *gin.Context) {
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
	c.JSON(http.StatusOK, nil)

}
