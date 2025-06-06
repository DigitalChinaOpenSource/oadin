package mcp_handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"byze/internal/hardware"
	"byze/internal/types"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// 缓存服务
type CacheService struct {
	cache map[string]struct {
		data      interface{}
		expiresAt time.Time
	}
	mu sync.RWMutex
}

func NewCacheService() *CacheService {
	return &CacheService{
		cache: make(map[string]struct {
			data      interface{}
			expiresAt time.Time
		}),
	}
}

func (cs *CacheService) Has(key string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	entry, exists := cs.cache[key]
	if !exists {
		return false
	}
	if time.Now().After(entry.expiresAt) {
		delete(cs.cache, key)
		return false
	}
	return true
}

func (cs *CacheService) Get(key string) interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	entry, exists := cs.cache[key]
	if !exists {
		return nil
	}
	if time.Now().After(entry.expiresAt) {
		delete(cs.cache, key)
		return nil
	}
	return entry.data
}

func (cs *CacheService) Set(key string, data interface{}, ttl time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.cache[key] = struct {
		data      interface{}
		expiresAt time.Time
	}{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

func (cs *CacheService) Remove(key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.cache, key)
}

// McpService 结构体
type McpClientService struct {
	clients        map[string]*client.Client
	pendingClients map[string]chan *client.Client
	cacheService   *CacheService
	mu             sync.Mutex
}

func NewMcpService() *McpClientService {
	return &McpClientService{
		clients:        make(map[string]*client.Client),
		pendingClients: make(map[string]chan *client.Client),
		cacheService:   NewCacheService(),
	}
}

func (s *McpClientService) getServerKey(server types.MCPServerConfig) string {
	return server.Id
}

func (s *McpClientService) initTransportClient(server types.MCPServerConfig, config types.McpUserConfig) (*client.Client, error) {
	// 命令执行路径
	// 环境变量
	if server.Command != "" {
		// command := "D:\\work_szsm\\20250603\\byze\\internal\\hardware\\installer\\runtime\\bun.exe"
		// args := []string{"x", "-y", "bing-cn-mcp"}
		commandBuilder := hardware.NewCommandBuilder(server.Command).WithArgs(server.Args...)
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
		} else if len(server.Env) > 0 {
			// 如果auth为空，则回退到使用y.Env
			for key, value := range server.Env {
				commandBuilder.WithEnv(key, value)
			}
		}

		cmd, err := commandBuilder.GetRunCommand()
		if err != nil {
			return nil, err
		}
		stdioTransport := transport.NewStdio(
			cmd.Path,
			cmd.Env,
			cmd.Args...,
		)
		fmt.Printf("Command to run: %s\n", cmd.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := stdioTransport.Start(ctx); err != nil {
			log.Printf("failed to start stdio transport: %v", err)
		}

		c := client.NewClient(stdioTransport)
		initRequest := mcp.InitializeRequest{}
		_, err = c.Initialize(ctx, initRequest)
		if err != nil {
			_ = stdioTransport.Close()
			log.Printf("failed to initialize stdio client: %v", err)
		}
		return c, nil
	}
	return nil, errors.New("either baseUrl or command must be provided")
}

func (s *McpClientService) Start(server types.MCPServerConfig, config types.McpUserConfig) (*client.Client, error) {
	serverKey := s.getServerKey(server)

	// 检查是否有正在初始化的客户端
	s.mu.Lock()
	if ch, exists := s.pendingClients[serverKey]; exists {
		s.mu.Unlock()
		client, ok := <-ch
		if !ok {
			return nil, errors.New("client initialization failed")
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

		cli, err := s.initTransportClient(server, config)
		if err != nil {
			log.Printf("[MCP] Failed to initialize client for server %s: %v", server.Name, err)
			return
		}

		s.mu.Lock()
		s.clients[serverKey] = cli
		s.mu.Unlock()
		ch <- cli
	}()

	client, ok := <-ch
	if !ok {
		return nil, errors.New("client initialization failed")
	}
	return client, nil
}

func (s *McpClientService) Stop(serverKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cli, exists := s.clients[serverKey]; exists {
		if err := cli.Close(); err != nil {
			return err
		}
		delete(s.clients, serverKey)
		s.cacheService.Remove(fmt.Sprintf("mcp:list_tool:%s", serverKey))
		log.Printf("[MCP] Closed server: %s", serverKey)
		log.Printf("[MCP] Cleared cache for server: %s", serverKey)
	} else {
		log.Printf("[MCP] No client found for server: %s", serverKey)
	}
	return nil
}

func (s *McpClientService) listTools(server types.MCPServerConfig, config types.McpUserConfig) ([]mcp.Tool, error) {
	cacheKey := fmt.Sprintf("mcp:list_tool:%s", s.getServerKey(server))
	if s.cacheService.Has(cacheKey) {
		if tools, ok := s.cacheService.Get(cacheKey).([]mcp.Tool); ok {
			log.Printf("[MCP] Tools from %s loaded from cache", server.Name)
			return tools, nil
		}
	}
	log.Printf("[MCP] Listing tools for server: %s", server.Name)
	cli, err := s.Start(server, config)
	if err != nil {
		return nil, err
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

func (s *McpClientService) CallTool(server types.MCPServerConfig, config types.McpUserConfig, name string, args interface{}) (*mcp.CallToolResult, error) {
	log.Printf("[MCP] Calling: %s %s %v", server.Name, name, args)
	cli, err := s.Start(server, config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fetchRequest := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
	}
	fetchRequest.Params.Name = name
	fetchRequest.Params.Arguments = args
	result, err := cli.CallTool(ctx, fetchRequest)
	if err != nil {
		log.Fatalf("😡 Failed to call the tool: %v", err)
	}
	return result, nil
}
