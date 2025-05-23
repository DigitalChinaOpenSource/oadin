package server

import (
	"byze/internal/datastore"
	"byze/internal/hardware"
	"byze/internal/rpc"
	"byze/internal/types"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"strings"
)

type MCPServer interface {
	GetMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error)
	GetMCP(ctx context.Context, id string) (*rpc.MCPDetailResponse, error)
	GetKits(ctx context.Context, id string, request *rpc.ToolSearchRequest) (*rpc.ToolSearchResponse, error)
	GetClients(ctx context.Context, id string) (*rpc.ClientListResponse, error)
	GetCategories(ctx context.Context) (*rpc.CategoryListResponse, error)
	GetMyMCPList(ctx context.Context, request *rpc.MCPListRequest) (*rpc.MCPListResponse, error)
	DownloadMCP(ctx context.Context, id string) (*rpc.MCPListResponse, error)
	AuthorizeMCP(ctx context.Context, id string, auth string) (*rpc.MCPListResponse, error)
	ReverseStatus(c *gin.Context, id string) error
	SetupFunTool(c *gin.Context, req rpc.SetupFunToolRequest) error
}

type MCPServerImpl struct {
	Ds     datastore.Datastore
	Client *resty.Client
}

func NewMCPServer() MCPServer {
	return &MCPServerImpl{
		Ds:     datastore.GetDefaultDatastore(),
		Client: rpc.GlobalClient,
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
		for mcp := range res.Data.List {
			for _, config := range configs {
				if res.Data.List[mcp].ID == config.(*types.McpUserConfig).MCPID {
					res.Data.List[mcp].Status = config.(*types.McpUserConfig).Status
					break
				}
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
		for _, item := range res.Data.List {
			item.Enabled = true
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

func (M *MCPServerImpl) DownloadMCP(ctx context.Context, id string) (*rpc.MCPListResponse, error) {
	//安装 npx
	err := hardware.InstallNpxEnvironment()
	if err != nil {
		return nil, err
	}

	//安装 uv
	envMgr := hardware.NewEnvManager()
	err = envMgr.InstallUVWithPython()
	if err != nil {
		return nil, err
	}

	// 执行mcp安装命令
	_, err = rpc.GetMCPDetail(M.Client, id)
	if err != nil {
		return nil, err
	}

	hardware.NewCommandBuilder("").Execute()

	return nil, nil

}

func (M *MCPServerImpl) AuthorizeMCP(ctx context.Context, id string, auth string) (*rpc.MCPListResponse, error) {
	con := new(types.McpUserConfig)
	con.MCPID = id
	err := M.Ds.Get(ctx, con)

	if err != nil {
		return nil, err
	}
	if con == nil || con.ID == 0 {
		return nil, err
	}

	// 保存授权配置项
	con.Auth = auth
	err = M.Ds.Put(ctx, con)
	if err != nil {
		return nil, err
	}

	return nil, nil
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

	// 保存授权配置项
	if req.Enabled == 0 {
		con.Kits += "," + req.ToolId
	} else {
		con.Kits = strings.Replace(con.Kits, ","+req.ToolId, "", -1)
	}

	err = M.Ds.Put(c.Request.Context(), con)
	if err != nil {
		return err
	}

	return nil
}
