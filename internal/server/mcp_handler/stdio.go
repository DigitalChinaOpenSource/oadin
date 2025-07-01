package mcp_handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	ConfigRoot "oadin/config"
	"oadin/internal/types"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type StdioTransport struct {
	clients map[string]*client.Client
	Pending map[string]*types.MCPServerConfig
	mu      sync.Mutex
}

var (
	stdioTransportInstance = StdioTransport{
		clients: make(map[string]*client.Client),
		Pending: make(map[string]*types.MCPServerConfig),
	}
	stdioTransportOnce sync.Once
)

func NewStdioTransport() *StdioTransport {
	return &stdioTransportInstance
}

func (s *StdioTransport) Start(ctx context.Context, config *types.MCPServerConfig) error {
	serverKey := config.Id

	// 检查是否有正在初始化的客户端
	s.mu.Lock()
	pendingConfig, exists := s.Pending[serverKey]
	if exists && pendingConfig.StartCacheTime.Add(10*time.Second).After(time.Now()) {
		s.mu.Unlock()
		log.Printf("[MCP] Client for server %s is already pending initialization", serverKey)
		fmt.Printf("[MCP] Client for server %s is already pending initialization\n", serverKey)
		return nil
	}

	// 检查是否已有客户端实例
	fmt.Printf("[MCP] check Starting client for server: %s\n", config.Id)
	if cli, exists := s.clients[serverKey]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cli.Ping(ctx); err == nil {
			s.mu.Unlock()
			return nil
		} else {
			s.Stop(serverKey)
		}
	}

	config.StartCacheTime = time.Now()
	s.Pending[serverKey] = config
	s.mu.Unlock()

	cli, err := s.ClientStart(ctx, config)
	if err != nil {
		fmt.Printf("[MCP] Failed to initialize client for server: %s, error: %v", config.Id, err)
		delete(s.Pending, serverKey)
		return err
	}
	fmt.Printf("[MCP] Initialized client for server: %s\n", config.Id)
	slog.Info("[MCP] Initialized client for server", "server_id", config.Id)
	s.mu.Lock()
	s.clients[serverKey] = cli
	s.mu.Unlock()

	return nil
}

func (s *StdioTransport) Stop(serverKey string) error {
	s.mu.Lock()
	delete(s.Pending, serverKey)
	cli, exists := s.clients[serverKey]
	if !exists {
		s.mu.Unlock()
		fmt.Printf("MCP Client for server %s does not exist, cannot stop\n", serverKey)
		return nil
	}
	// 先从map删除，防止并发重复关闭
	delete(s.clients, serverKey)
	s.mu.Unlock()

	// 关闭客户端
	if err := cli.Close(); err != nil {
		fmt.Printf("MCP Error closing client for server %s: %v\n", serverKey, err)
	} else {
		fmt.Printf("MCP Closed server: %s\n", serverKey)
	}
	return nil
}

func (s *StdioTransport) FetchTools(ctx context.Context, serverKey string) ([]mcp.Tool, error) {
	config, exists := s.Pending[serverKey]
	if !exists && config != nil {
		cli, exists := s.clients[serverKey]
		if !exists {
			return nil, errors.New("client not found")
		}

		toolsRequest := mcp.ListToolsRequest{}
		tools, err := cli.ListTools(ctx, toolsRequest)
		if err != nil {
			return nil, err
		}

		if tools == nil {
			return nil, errors.New("no tools found")
		}

		config.Tools = tools.Tools
		return tools.Tools, nil
	}

	if len(config.Tools) == 0 {
		cli, exists := s.clients[serverKey]
		if !exists {
			return nil, errors.New("client not found")
		}

		toolsRequest := mcp.ListToolsRequest{}
		tools, err := cli.ListTools(ctx, toolsRequest)
		if err != nil {
			return nil, err
		}
		config.Tools = tools.Tools
		return tools.Tools, nil
	}

	return config.Tools, nil
}

func (s *StdioTransport) CallTool(ctx context.Context, mcpId string, params mcp.CallToolParams) (*mcp.CallToolResult, error) {
	cli, exist := s.clients[mcpId]
	if !exist {
		return nil, fmt.Errorf("client for MCP %s not found", mcpId)
	}
	fetchRequest := mcp.CallToolRequest{}
	fetchRequest.Params.Name = params.Name
	fetchRequest.Params.Arguments = params.Arguments
	result, err := cli.CallTool(ctx, fetchRequest)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *StdioTransport) ClientStart(ctx context.Context, config *types.MCPServerConfig) (*client.Client, error) {
	var envVars []string
	for k, v := range config.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	envVars = append(envVars, fmt.Sprintf("%s=%s", "NPM_CONFIG_REGISTRY", ConfigRoot.ConfigRootInstance.Registry.Npm))
	envVars = append(envVars, fmt.Sprintf("%s=%s", "PIP_INDEX_URL", ConfigRoot.ConfigRootInstance.Registry.Pip))
	envVars = append(envVars, fmt.Sprintf("%s=%s", "UV_DEFAULT_INDEX", ConfigRoot.ConfigRootInstance.Registry.Pip))
	stdioTransport := transport.NewStdio(
		config.Command,
		envVars,
		config.Args...,
	)

	fmt.Println("[MCP] params", config.Id, "command:", config.Command, "args:", config.Args, "env:", envVars)

	if err := stdioTransport.Start(ctx); err != nil {
		fmt.Printf("failed to start stdio transport: %v", err)
		return nil, err
	}

	c := client.NewClient(stdioTransport)
	initRequest := mcp.InitializeRequest{}
	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		_ = stdioTransport.Close()
		fmt.Printf("failed to initialize stdio client for server %s: %v\n", config.Id, err)
		return nil, err
	}
	toolsRequest := mcp.ListToolsRequest{}
	tools, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		fmt.Println("failed to ListTools err", err)
		return nil, err
	}
	fmt.Println("tools", tools.Tools)
	config.Tools = tools.Tools
	return c, nil
}
