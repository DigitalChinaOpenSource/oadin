package mcp_handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"byze/internal/types"

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

func (s *StdioTransport) initTransportClient(config *types.MCPServerConfig) (*client.Client, error) {
	if config.Command == "" {
		return nil, errors.New("command must be provided")
	}
	// try := func() (*client.Client, error) {
	// 	var envVars []string
	// 	for k, v := range config.Env {
	// 		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	// 	}
	// 	stdioTransport := transport.NewStdio(
	// 		config.Command,
	// 		envVars,
	// 		config.Args...,
	// 	)

	// 	slog.Info("[MCP] Initializing stdio transport for server", "server_id", config.Id, "command", config.Command, "args", config.Args, "env", envVars)
	// 	fmt.Println("[MCP] Initializing stdio transport for server:", config.Id, "command:", config.Command, "args:", config.Args, "env:", envVars)
	// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()

	// 	if err := stdioTransport.Start(ctx); err != nil {
	// 		log.Printf("failed to start stdio transport: %v", err)
	// 		return nil, err
	// 	}

	// 	c := client.NewClient(stdioTransport)
	// 	initRequest := mcp.InitializeRequest{}
	// 	_, err := c.Initialize(ctx, initRequest)
	// 	if err != nil {
	// 		_ = stdioTransport.Close()
	// 		fmt.Printf("failed to initialize stdio client for server %s: %v\n", config.Id, err)
	// 		return nil, err
	// 	}
	// 	return c, nil
	// }

	for i := range 3 {
		c, err := s.ClientMAC(config)
		if err == nil {
			return c, nil
		}
		log.Printf("initTransportClient: retrying after failure for server %s, attempt %d", config.Id, i+1)
		time.Sleep(1 * time.Second)
	}

	// 如果重试失败，删除Pending状态
	c, err := s.ClientMAC(config)
	if err == nil {
		return c, nil
	} else {
		delete(s.Pending, config.Id)
		return nil, err
	}
}

func (s *StdioTransport) Start(config *types.MCPServerConfig) error {
	serverKey := config.Id

	// 检查是否有正在初始化的客户端
	s.mu.Lock()
	_, exists := s.Pending[serverKey]
	if exists {
		s.mu.Unlock()
		log.Printf("[MCP] Client for server %s is already pending initialization", serverKey)
		return nil
	}

	// 检查是否已有客户端实例
	if cli, exists := s.clients[serverKey]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := cli.Ping(ctx); err == nil {
			s.mu.Unlock()
			return nil
		}
		s.Stop(serverKey)
		delete(s.clients, serverKey)
		delete(s.Pending, serverKey)
	}

	s.Pending[serverKey] = config
	s.mu.Unlock()

	// 异步初始化客户端
	cli, err := s.initTransportClient(config)
	if err != nil {
		log.Printf("[MCP] Failed to initialize client for server: %s, error: %v", config.Id, err)
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
	if !exists {
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

	try := func() (*mcp.CallToolResult, error) {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		fetchRequest := mcp.CallToolRequest{}
		fetchRequest.Params.Name = params.Name
		fetchRequest.Params.Arguments = params.Arguments
		result, err := cli.CallTool(ctx, fetchRequest)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	for i := range 10 {
		result, err := try()
		if err == nil {
			return result, nil
		}
		fmt.Printf("Retrying to call tool for MCP %s, attempt %d\n", mcpId, i+1)
		slog.Info("Retrying to call tool for MCP, attempt ", mcpId, i+1)
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("failed to call tool after 10 attempts for MCP %s", mcpId)
}

func (s *StdioTransport) ClientMAC(config *types.MCPServerConfig) (*client.Client, error) {
	if config == nil {
		config = &types.MCPServerConfig{
			Id:      "mac",
			Command: "/Users/aipc/Library/Application Support/Byze/runtime/bun",
			Args:    []string{"x", "-y", "@amap/amap-maps-mcp-server"},
			Env:     map[string]string{"AMAP_MAPS_API_KEY": "486fe8946aa80aa2baf26d840b6fa6a0"},
		}
	}
	var envVars []string
	for k, v := range config.Env {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}
	stdioTransport := transport.NewStdio(
		config.Command,
		envVars,
		config.Args...,
	)

	fmt.Println("[MCP] params", config.Id, "command:", config.Command, "args:", config.Args, "env:", envVars)
	ctx := context.Background()

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
	return c, nil
}
