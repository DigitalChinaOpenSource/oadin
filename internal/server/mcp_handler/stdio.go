package mcp_handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"byze/internal/types"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type StdioTransport struct {
	clients map[string]*client.Client
	pending map[string]int
	mu      sync.Mutex
}

var (
	stdioTransportInstance = StdioTransport{
		clients: make(map[string]*client.Client),
		pending: make(map[string]int),
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
	try := func() (*client.Client, error) {
		var envVars []string
		for k, v := range config.Env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		stdioTransport := transport.NewStdio(
			config.Command,
			envVars,
			config.Args...,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := stdioTransport.Start(ctx); err != nil {
			log.Printf("failed to start stdio transport: %v", err)
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
		return c, nil
	}
	// 第一次尝试
	c, err := try()
	if err == nil {
		return c, nil
	}

	// 等待一小段时间后重试一次
	time.Sleep(1 * time.Second)
	log.Printf("initTransportClient: retrying after failure for server %s", config.Id)
	return try()
}

func (s *StdioTransport) Start(config *types.MCPServerConfig) error {
	serverKey := config.Id

	// 检查是否有正在初始化的客户端
	s.mu.Lock()
	if s.pending[serverKey] > 0 {
		s.mu.Unlock()
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
		delete(s.pending, serverKey)
	}
	s.mu.Unlock()

	// 异步初始化客户端
	s.pending[serverKey] = 1
	go func(s *StdioTransport) {
		cli, err := s.initTransportClient(config)
		if err != nil {
			log.Printf("[MCP] Failed to initialize client for server: %s, error: %v", config.Id, err)
			return
		}
		fmt.Printf("[MCP] Initialized client for server: %s\n", config.Id)
		s.mu.Lock()
		s.clients[serverKey] = cli
		s.mu.Unlock()
	}(s)

	return nil
}

func (s *StdioTransport) Stop(serverKey string) error {
	s.mu.Lock()
	cli, exists := s.clients[serverKey]
	if !exists {
		s.mu.Unlock()
		fmt.Printf("MCP Client for server %s does not exist, cannot stop\n", serverKey)
		return nil
	}
	// 先从map删除，防止并发重复关闭
	delete(s.clients, serverKey)
	delete(s.pending, serverKey)
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
	cli, exists := s.clients[serverKey]
	if !exists {
		return nil, errors.New("client not found")
	}

	toolsRequest := mcp.ListToolsRequest{}
	tools, err := cli.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}
	return tools.Tools, nil
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
		fmt.Printf("Failed to call the tool: %v\n", err)
		return nil, err
	}
	return result, nil
}
