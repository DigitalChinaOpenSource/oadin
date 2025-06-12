package rpc

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/mark3labs/mcp-go/mcp"
)

// 1. MCP 列表检索
type MCPListRequest struct {
	Keyword    string   `json:"keyword"`
	Category   []string `json:"category"`
	Tag        []string `json:"tags"`
	Deployment string   `json:"deployment"`
	MCPIds     []string `json:"id"`
	Page       int      `json:"page"`
	Size       int      `json:"size"`
}

type MCPListResponse struct {
	Code int `json:"code"`
	Data struct {
		Total int `json:"total"`
		List  []struct {
			ID   string `json:"id"`
			Name struct {
				Src string `json:"src"`
				Zh  string `json:"zh"`
			} `json:"name"`
			Abstract struct {
				Src string `json:"src"`
				Zh  string `json:"zh"`
			} `json:"abstract"`
			Supplier    string   `json:"supplier"`
			Logo        string   `json:"logo"`
			Popularity  int      `json:"popularity"`
			Tags        []string `json:"tags"`
			Hosted      bool     `json:"hosted"`
			Status      int      `json:"status"`
			Authorized  int      `json:"authorized"`
			EnvRequired int      `json:"envRequired"`
			UpdatedAt   int64    `json:"updatedAt"`
		} `json:"list"`
	} `json:"data"`
}

func GetMCPList(client *resty.Client, req MCPListRequest) (*MCPListResponse, error) {
	var resp MCPListResponse
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&resp).
		Post("/api/mcp/search")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("API错误: %d", resp.Code)
	}
	return &resp, nil
}

// 2. 获取 MCP 详情
type MCPDetailResponse struct {
	Code int `json:"code"`
	Data struct {
		ID   string `json:"id"`
		Name struct {
			Src string `json:"src"`
			Zh  string `json:"zh"`
		} `json:"name"`
		ServerName  string   `json:"serverName"`
		Supplier    string   `json:"supplier"`
		Authorized  int      `json:"authorized"`
		EnvRequired int      `json:"envRequired"`
		Tags        []string `json:"tags"`
		Hosted      bool     `json:"hosted"`
		Status      int      `json:"status"`
		UpdatedAt   int64    `json:"updatedAt"`
		Summary     struct {
			Src string `json:"src"`
			Zh  string `json:"zh"`
		} `json:"summary"`

		Abstract struct {
			Src string `json:"src"`
			Zh  string `json:"zh"`
		} `json:"abstract"`

		EnvSchema struct {
			Type       string                 `json:"type"`
			Required   []string               `json:"required"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"envSchema"`
		Logo         string `json:"logo"`
		ServerConfig []struct {
			McpServers map[string]struct {
				Args    []string          `json:"args"`
				Command string            `json:"command"`
				Env     map[string]string `json:"env"` // 使用map存储动态环境变量
			} `json:"mcpServers"`
		} `json:"serverConfig"`
	} `json:"data"`
}

func GetMCPDetail(client *resty.Client, id string) (*MCPDetailResponse, error) {
	var resp MCPDetailResponse
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("id", id).
		SetResult(&resp).
		Get("/api/mcp/search/{id}")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("API错误: %d", resp.Code)
	}
	return &resp, nil
}

// 3. 工具函数列表检索
type ToolSearchRequest struct {
	Keyword string `json:"keyword"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

type ToolSearchResponse struct {
	Code int `json:"code"`
	Data struct {
		Total int `json:"total"`
		List  []struct {
			Description string   `json:"description"`
			Name        string   `json:"name"`
			Server      string   `json:"server"`
			Tool        string   `json:"tool"`
			Id          string   `json:"id"`
			Tags        []string `json:"tags"`    // 使用字符串切片处理空数组
			Enabled     bool     `json:"enabled"` // 默认启用状态

			InputSchema struct {
				Schema               string   `json:"$schema"`  // 注意$符号的特殊处理
				RequiredProps        []string `json:"required"` // 单独处理required字段
				AdditionalProperties bool     `json:"additionalProperties"`

				Properties struct {
					Command struct {
						Type     string `json:"type"`
						Required bool   `json:"required"` // 显式标注必填
					} `json:"command"`

					TimeoutMs struct {
						Type     string `json:"type"`
						Required bool   `json:"required"`
					} `json:"timeout_ms"`
				} `json:"properties"`

				Type string `json:"type"`
			} `json:"inputSchema"`
		} `json:"list"`
	} `json:"data"`
}

func SearchTools(client *resty.Client, id string, req *ToolSearchRequest) (*ToolSearchResponse, error) {
	var resp ToolSearchResponse
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("id", id).
		SetBody(req).
		SetResult(&resp).
		Post("/api/mcp/tools/search/{id}")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("API错误: %d", resp.Code)
	}
	return &resp, nil
}

// 4. 获取推荐客户端
type ClientItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	LinkCommand string   `json:"linkCommand"`
	Description string   `json:"description"`
	RelatedTags []string `json:"relatedTags"`
	SortWeight  int      `json:"sortWeight"`
	CreateBy    string   `json:"createBy"`
	UpdateBy    string   `json:"updateBy"`
}

type ClientListResponse struct {
	Code int          `json:"code"`
	Data []ClientItem `json:"data"`
}

func GetRecommendedClients(client *resty.Client) (*ClientListResponse, error) {
	var resp ClientListResponse
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resp).
		Get("/api/clients")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	return &resp, nil
}

// 5. 获取分类与子标签组
type CategoryItem struct {
	Category string `json:"category"`
	Tags     []Tag  `json:"tags"`
}
type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type CategoryListResponse struct {
	Code int            `json:"code"`
	Data []CategoryItem `json:"data"`
}

func GetCategories(client *resty.Client) (*CategoryListResponse, error) {
	var resp CategoryListResponse
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resp).
		Get("/api/mcp/categories")

	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	return &resp, nil
}

type SetupFunToolRequest struct {
	MCPId   string `json:"mcpId"`
	Enabled bool   `json:"enabled"`
	ToolId  string `json:"toolId"`
}

type ClientMcpStartRequest struct {
	Id string `json:"id"`
}

type ClientMcpStartResponse struct {
	Id string `json:"id"`
}

type ClientMcpStopRequest struct {
	Ids []string `json:"ids"`
}

type ClientRunToolRequest struct {
	MCPId    string         `json:"mcpId"`
	ToolName string         `json:"toolName"`
	ToolArgs map[string]any `json:"toolArgs"`
}

type ClientGetToolsRequest struct {
	Ids []string `json:"ids"`
}

type ClientGetToolsResponse struct {
	Tools []McpTool `json:"mcpTools"`
}

type McpTool struct {
	McpId string `json:"mcpId"`
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function TypeFunction `json:"function"`
}
type TypeFunction struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  mcp.ToolInputSchema `json:"parameters"`
}
