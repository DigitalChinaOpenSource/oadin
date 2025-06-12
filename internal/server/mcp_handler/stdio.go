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

// McpService 结构体
type StdioTransport struct {
	clients        map[string]*client.Client
	pendingClients map[string]chan *client.Client
	mu             sync.Mutex
}

var (
	stdioTransportInstance StdioTransport
	stdioTransportOnce     sync.Once
)

func NewMcpService() *StdioTransport {
	stdioTransportOnce.Do(func() {
		stdioTransportInstance = StdioTransport{
			clients:        make(map[string]*client.Client),
			pendingClients: make(map[string]chan *client.Client),
		}
	})
	return &stdioTransportInstance
}

func (s *StdioTransport) initTransportClient(config types.MCPServerConfig) (*client.Client, error) {
	fmt.Println("Initializing transport client with config:", config)
	if config.Command == "" {
		return nil, errors.New("either baseUrl or command must be provided")
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := stdioTransport.Start(ctx); err != nil {
		fmt.Printf("failed to start stdio transport: %v", err)
		return nil, err
	}

	c := client.NewClient(stdioTransport)
	initRequest := mcp.InitializeRequest{}
	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		_ = stdioTransport.Close()
		fmt.Printf("failed to initialize stdio client: %v", err)
		return nil, err
	}
	return c, nil
}

func (s *StdioTransport) Start(config types.MCPServerConfig) (*client.Client, error) {
	serverKey := config.Id

	// 检查是否有正在初始化的客户端
	s.mu.Lock()
	if ch, exists := s.pendingClients[serverKey]; exists {
		s.mu.Unlock()
		client, ok := <-ch
		if !ok {
			return nil, errors.New("pendingClients client initialization failed")
		}
		return client, nil
	}

	// 检查是否已有客户端实例
	if cli, exists := s.clients[serverKey]; exists {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := cli.Ping(ctx); err == nil {
			s.mu.Unlock()
			return cli, nil
		}
		delete(s.clients, serverKey)
	}

	// 创建一个通道来等待客户端初始化
	ch := make(chan *client.Client, 1)
	s.pendingClients[serverKey] = ch
	s.mu.Unlock()

	// 异步初始化客户端
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.pendingClients, serverKey)
			s.mu.Unlock()
			close(ch)
		}()

		cli, err := s.initTransportClient(config)
		if err != nil {
			fmt.Println("[MCP] Failed to initialize client for server : ", config.Name, err)
			return
		}
		s.mu.Lock()
		s.clients[serverKey] = cli
		s.mu.Unlock()
		ch <- cli
	}()

	// client, ok := <-ch
	// if !ok {
	// 	return nil, errors.New("client initialization failed")
	// }
	return nil, nil
}

func (s *StdioTransport) Stop(serverKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cli, exists := s.clients[serverKey]; exists {
		if err := cli.Close(); err != nil {
			return err
		}
		delete(s.clients, serverKey)
		fmt.Printf("[MCP] Closed server: %s", serverKey)
		return nil
	} else {
		fmt.Printf("[MCP] No client found for server: %s", serverKey)
		return errors.New("client not found")
	}
}

func (s *StdioTransport) FetchTools(serverKey string) ([]mcp.Tool, error) {
	cli, exists := s.clients[serverKey]
	if !exists {
		return nil, errors.New("client not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	toolsRequest := mcp.ListToolsRequest{}
	tools, err := cli.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, err
	}
	return tools.Tools, nil
}

func (s *StdioTransport) CallTool(serverKey string, params mcp.CallToolParams) (*mcp.CallToolResult, error) {
	cli, exists := s.clients[serverKey]
	if !exists {
		return nil, errors.New("client not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fetchRequest := mcp.CallToolRequest{}
	fetchRequest.Params.Name = params.Name
	fetchRequest.Params.Arguments = params.Arguments
	result, err := cli.CallTool(ctx, fetchRequest)
	if err != nil {
		log.Fatalf("Failed to call the tool: %v", err)
		return nil, err
	}
	return result, nil
}
