package server

import (
	"byze/internal/datastore"
	"byze/internal/hardware"
	"byze/internal/hardware/installer"
	"byze/internal/rpc"
	"byze/internal/server/mcp_handler"
	"byze/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/mark3labs/mcp-go/mcp"
)

type MCPServer interface {
	GetMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error)
	GetMCP(ctx context.Context, id string) (*rpc.MCPDetailResponse, error)
	GetKits(ctx context.Context, id string, request *rpc.ToolSearchRequest) (*rpc.ToolSearchResponse, error)
	GetClients(ctx context.Context, id string) (*rpc.ClientListResponse, error)
	GetCategories(ctx context.Context) (*rpc.CategoryListResponse, error)
	GetMyMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error)
	DownloadMCP(ctx context.Context, id string) error
	AuthorizeMCP(ctx context.Context, id string, auth string) error
	ReverseStatus(c *gin.Context, id string) error
	SetupFunTool(c *gin.Context, req rpc.SetupFunToolRequest) error
	ClientMcpStart(c *gin.Context, id string) ([]mcp.Tool, error)
	ClientMcpStop(c *gin.Context, id string) error
	ClientRunTool(c *gin.Context, req *rpc.ClientRunToolRequest) (*mcp.CallToolResult, error)
}

type MCPServerImpl struct {
	Ds         datastore.Datastore
	Client     *resty.Client
	McpHandler *mcp_handler.StdioTransport
}

func NewMCPServer() MCPServer {
	return &MCPServerImpl{
		Ds:         datastore.GetDefaultDatastore(),
		Client:     rpc.GlobalClient,
		McpHandler: mcp_handler.NewMcpService(),
	}
}

func (M *MCPServerImpl) GetMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error) {

	res, err := rpc.GetMCPList(M.Client, *request)

	if err != nil {
		return nil, err
	}

	// 需要加载本地配置数据
	configs, err := M.Ds.List(ctx, &types.McpUserConfig{}, &datastore.ListOptions{})
	if err != nil {
		return nil, err
	}
	for mcp := range res.Data.List {
		for _, config := range configs {
			if res.Data.List[mcp].ID == config.(*types.McpUserConfig).MCPID {
				res.Data.List[mcp].Status = config.(*types.McpUserConfig).Status
				break
			}
		}

	}
	return res, err
}

func (M *MCPServerImpl) GetMCP(ctx context.Context, id string) (*rpc.MCPDetailResponse, error) {
	res, err := rpc.GetMCPDetail(M.Client, id)

	if err != nil {
		return nil, err
	}

	// 需要加载本地配置数据
	configs, err := M.Ds.List(ctx, &types.McpUserConfig{}, &datastore.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, config := range configs {
		if config.(*types.McpUserConfig).MCPID == id {
			// 插入本地配置数据到mcp中
			res.Data.Status = config.(*types.McpUserConfig).Status

			if config.(*types.McpUserConfig).Auth != "" {
				res.Data.Authorized = 1
			} else {
				res.Data.Authorized = 0
			}

			break
		}
	}
	return res, err
}

func (M *MCPServerImpl) GetKits(ctx context.Context, id string, request *rpc.ToolSearchRequest) (*rpc.ToolSearchResponse, error) {
	res, err := rpc.SearchTools(M.Client, id, request)
	if err != nil {
		return nil, err
	}

	// 需要加载本地配置数据
	con := new(types.McpUserConfig)
	con.MCPID = id

	_ = M.Ds.Get(ctx, con)
	if con == nil || con.ID == 0 || con.Kits == "" {
		for i, _ := range res.Data.List {
			res.Data.List[i].Enabled = true
		}
	} else {
		// 配置数据组合
		for i, tool := range res.Data.List {
			// 默认开启
			res.Data.List[i].Enabled = true
			for _, item := range strings.Split(con.Kits, ",") {
				if item == tool.Id {
					res.Data.List[i].Enabled = false
					break
				}
			}
		}
	}

	return res, err
}

func (M *MCPServerImpl) GetClients(ctx context.Context, id string) (*rpc.ClientListResponse, error) {
	res, err := rpc.GetRecommendedClients(M.Client)
	if err != nil {
		return nil, err
	}

	return res, err

}

func (M *MCPServerImpl) GetCategories(ctx context.Context) (*rpc.CategoryListResponse, error) {
	res, err := rpc.GetCategories(M.Client)
	if err != nil {
		return nil, err
	}

	return res, err
}

func (M *MCPServerImpl) GetMyMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error) {
	// 先获取属于我的mcp数据, 带入条件去查询列表数据
	configs, err := M.Ds.List(ctx, &types.McpUserConfig{}, &datastore.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.(*types.McpUserConfig).Status == 1 {
			request.MCPIds = append(request.MCPIds, config.(*types.McpUserConfig).MCPID)
		}
	}
	return M.GetMCPList(ctx, request)

}

func (M *MCPServerImpl) DownloadMCP(ctx context.Context, id string) error {
	mcp, err := rpc.GetMCPDetail(M.Client, id)
	if err != nil {
		return err
	}
	if mcp == nil || mcp.Code != 200 || mcp.Data.ID == "" {
		return err
	}

	// 初始化个人配置
	config := new(types.McpUserConfig)
	config.MCPID = id
	config.Status = 0
	config.Auth = ""
	// TODO: 后续优化通过json 数组的方式读取数据库的数据
	config.Kits = ""
	// 数据库不存在则初始化一条
	err = M.Ds.Get(ctx, config)
	if err != nil || config == nil || config.ID == 0 {
		err := M.Ds.Add(ctx, config)
		if err != nil {
			return err
		}
	}

	// bun和uv环境检测与下载安装逻辑
	installer.SetupBunAndUv()

	// 执行mcp安装命令
	installCMD := mcp.Data.ServerConfig

	for _, item := range installCMD {
		for _, y := range item.McpServers {
			// 传递args
			commandBuilder := hardware.NewCommandBuilder(y.Command).WithArgs(y.Args...)
			// 从用户配置的auth字段获取环境变量
			if config.Auth != "" {
				var authMap map[string]string
				err := json.Unmarshal([]byte(config.Auth), &authMap)
				if err == nil {
					for key, value := range authMap {
						commandBuilder.WithEnv(key, value)
					}
				} else {
					// 如果解析失败，则回退到使用AUTH_TOKEN
					commandBuilder.WithEnv("AUTH_TOKEN", config.Auth)
				}
			} else if len(y.Env) > 0 {
				// 如果auth为空，则回退到使用y.Env
				for key, value := range y.Env {
					commandBuilder.WithEnv(key, value)
				}
			}
			// 执行安装命令
			output, errOut, err := commandBuilder.WithTimeout(time.Minute).Execute()
			if err != nil {
				return err
			}
			fmt.Printf("output of command execution: %s", output)
			fmt.Printf("error output of command execution: %s", errOut)

		}
	}

	// 下载完毕后, 则添加成功
	config.Status = 1
	M.Ds.Put(ctx, config)

	return nil

}

func (M *MCPServerImpl) AuthorizeMCP(ctx context.Context, id string, auth string) error {
	con := new(types.McpUserConfig)
	con.MCPID = id
	err := M.Ds.Get(ctx, con)

	if err != nil || con == nil || con.ID == 0 {
		// 初始化个人配置
		con.MCPID = id
		con.Status = 0
		con.Auth = auth
		con.Kits = ""
		M.Ds.Add(ctx, con)
	}

	// 保存授权配置项
	con.Auth = auth
	err = M.Ds.Put(ctx, con)
	if err != nil {
		return err
	}

	return nil
}

func (M *MCPServerImpl) ReverseStatus(c *gin.Context, id string) error {

	con := new(types.McpUserConfig)
	con.MCPID = id
	err := M.Ds.Get(c.Request.Context(), con)

	if err != nil {
		return err
	}
	if con == nil || con.ID == 0 {
		return err
	}

	// 保存授权配置项
	if con.Status == 1 {
		con.Status = 0
	}

	err = M.Ds.Put(c.Request.Context(), con)
	if err != nil {
		return err
	}

	return nil
}

func (M *MCPServerImpl) SetupFunTool(c *gin.Context, req rpc.SetupFunToolRequest) error {
	con := new(types.McpUserConfig)
	con.MCPID = req.MCPId
	err := M.Ds.Get(c.Request.Context(), con)

	if err != nil {
		return err
	}
	if con == nil || con.ID == 0 {
		return err
	}

	// 将逗号分隔的字符串转换为map，便于处理
	toolMap := make(map[string]bool)
	if con.Kits != "" {
		for _, id := range strings.Split(con.Kits, ",") {
			if id != "" { // 忽略空字符串
				toolMap[id] = true
			}
		}
	}

	// 更新工具状态
	if !req.Enabled {
		// 要禁用工具：添加到禁用列表
		toolMap[req.ToolId] = true
	} else {
		// 要启用工具：从禁用列表中移除
		delete(toolMap, req.ToolId)
	}

	// 将map转回逗号分隔的字符串
	var toolIds []string
	for id := range toolMap {
		toolIds = append(toolIds, id)
	}

	// 重置为新的字符串
	if len(toolIds) > 0 {
		con.Kits = strings.Join(toolIds, ",")
	} else {
		con.Kits = "" // 如果没有禁用的工具，设为空字符串
	}

	err = M.Ds.Put(c.Request.Context(), con)
	if err != nil {
		return err
	}

	return nil
}

func (M *MCPServerImpl) ClientMcpStart(ctx *gin.Context, id string) ([]mcp.Tool, error) {
	mcpUserConfig := new(types.McpUserConfig)
	mcpUserConfig.MCPID = id

	err := M.Ds.Get(ctx, mcpUserConfig)
	if err != nil {
		return nil, err
	}

	mcpConfig, err := rpc.GetMCPDetail(M.Client, id)
	if err != nil {
		return nil, err
	}
	var env map[string]string
	if mcpUserConfig.Auth != "" {
		err := json.Unmarshal([]byte(mcpUserConfig.Auth), &env)
		if err != nil {
			return nil, err
		}
	}

	serverConfig := mcpConfig.Data.ServerConfig[0]
	mcpServers := serverConfig.McpServers
	config := mcpServers[mcpConfig.Data.ServerName]
	mcpServerConfig := types.MCPServerConfig{
		Id:      id,
		Name:    mcpConfig.Data.ServerName,
		Args:    config.Args,
		Command: config.Command,
		Env:     env,
	}

	_, err = M.McpHandler.Start(mcpServerConfig)
	if err != nil {
		return nil, err
	}

	searchTools, err := rpc.SearchTools(M.Client, id, &rpc.ToolSearchRequest{Size: 100, Page: 1})
	if err != nil {
		return nil, err
	}
	if mcpUserConfig == nil || mcpUserConfig.ID == 0 || mcpUserConfig.Kits == "" {
		for i := range searchTools.Data.List {
			searchTools.Data.List[i].Enabled = true
		}
	} else {
		// 配置数据组合
		for i, tool := range searchTools.Data.List {
			// 默认开启
			searchTools.Data.List[i].Enabled = true
			for _, item := range strings.Split(mcpUserConfig.Kits, ",") {
				if item == tool.Id {
					searchTools.Data.List[i].Enabled = false
					break
				}
			}
		}
	}

	fetchTools, err := M.McpHandler.FetchTools(mcpServerConfig)
	if err != nil {
		return nil, err
	}

	var tools []mcp.Tool
	for i, tool := range fetchTools {
		for _, item := range searchTools.Data.List {
			if item.Name == tool.Name && item.Enabled {
				tools = append(tools, fetchTools[i])
				break
			}
		}
	}
	return tools, nil
}

func (M *MCPServerImpl) ClientMcpStop(ctx *gin.Context, id string) error {
	err := M.McpHandler.Stop(id)
	if err != nil {
		return err
	}
	return nil
}

// RunTools 运行单个mcp的工具
func (M *MCPServerImpl) ClientRunTool(c *gin.Context, req *rpc.ClientRunToolRequest) (*mcp.CallToolResult, error) {
	params := mcp.CallToolParams{
		Name:      req.ToolName,
		Arguments: req.ToolArgs,
	}
	data, err := M.McpHandler.CallTool(req.MCPId, params)
	if err != nil {
		return nil, err
	}
	return data, nil
}
