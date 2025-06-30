package mcp_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"oadin/internal/types"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestMcpClientAndLLM(t *testing.T) {
	// 我需要用代码实现以下逻辑：第一步是初始化一个 MCP 客户端，使用 stdio 传输方式连接到 MCP 服务。
	// 第二步是发送一个http请求ollama的涉及工具调用的chat请求。
	// 第三步是处理返回的结果，打印出工具调用的结果。
	config := types.MCPServerConfig{
		Id:      "683ec88241fa614eb1531fc7",
		Command: "C:\\Users\\Intel\\AppData\\Roaming\\Oadin\\runtime\\bun.exe",
		Args:    []string{"x", "-y", "bing-cn-mcp"},
		Env:     nil,
	}
	mcpService := NewStdioTransport()
	ctx := context.Background()
	err := mcpService.Start(ctx, &config)
	if err != nil {
		log.Fatalf("Failed to initialize MCP client: %v", err)
	}

	tools, err := mcpService.FetchTools(ctx, config.Id)
	if err != nil {
		log.Fatalf("Failed to fetch tools: %v", err)
	}
	log.Printf("Available tools: %v", tools)

	// 一个json字符串,转为golang map接受
	url := "http://localhost:58380/api/chat"
	method := "POST"
	type toolOllama struct {
		Name     string `json:"type"`
		Function struct {
			Name        string              `json:"name"`
			Description string              `json:"description"`
			Parameters  mcp.ToolInputSchema `json:"parameters"`
		} `json:"function"`
	}

	var toolOllamas []toolOllama
	for _, tool := range tools {
		toolOllamas = append(toolOllamas, toolOllama{
			Name: tool.Name,
			Function: struct {
				Name        string              `json:"name"`
				Description string              `json:"description"`
				Parameters  mcp.ToolInputSchema `json:"parameters"`
			}{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}

	queryMap := map[string]interface{}{
		"model": "qwen2.5:7b",
		"messages": map[string]interface{}{
			"role":    "user",
			"content": "What's the weather like in Paris?",
		},
		"stream": false,
		"tools":  toolOllamas,
	}

	// finalText := []any
	// toolResults := []any

	client := &http.Client{}
	jsonData, err := json.Marshal(queryMap)
	if err != nil {
		fmt.Println("Failed to marshal queryMap:", err)
		return
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonData))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("User-Agent", "Apifox/1.0.0 (https://apifox.com)")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Host", "localhost:58380")
	req.Header.Add("Connection", "keep-alive")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))

	// 处理返回的结果
	response := struct {
		Message struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			ToolCalls []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments any    `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	}{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("Failed to unmarshal response: %v", err)
	}

	for _, toolCall := range response.Message.ToolCalls {
		_, _ = mcpService.CallTool(ctx, config.Id, mcp.CallToolParams{
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		})
	}
}

func TestMcpClientTest(t *testing.T) {
	config := types.MCPServerConfig{
		Command: "C:\\Users\\Intel\\AppData\\Roaming\\Oadin\\runtime\\bun.exe",
		Args:    []string{"x", "-y", "bing-cn-mcp"},
		Env:     nil,
	}
	fmt.Println("Initializing transport client with config:", config)
	if config.Command != "" {
		// command := "D:\\work_szsm\\20250603\\oadin\\internal\\hardware\\installer\\runtime\\bun.exe"
		// args := []string{"x", "-y", "bing-cn-mcp"}
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
			log.Printf("failed to start stdio transport: %v", err)
		}

		c := client.NewClient(stdioTransport)
		initRequest := mcp.InitializeRequest{}
		_, err := c.Initialize(ctx, initRequest)
		if err != nil {
			_ = stdioTransport.Close()
			log.Printf("failed to initialize stdio client: %v", err)
		}

		// List available tools
		toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			log.Printf("failed to list tools: %v", err)
		}

		log.Printf("Available tools: %v", toolsResult.Tools)
	}
}

// 这个可以跑通了
func TestMcpClientListTools(t *testing.T) {
	command := "D:\\work_szsm\\20250603\\oadin\\internal\\hardware\\installer\\runtime\\bun.exe"
	args := []string{"x", "-y", "bing-cn-mcp"}
	stdioTransport := transport.NewStdio(
		command,
		nil,
		args...,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := stdioTransport.Start(ctx); err != nil {
		log.Printf("failed to start stdio transport: %v", err)
	}

	c := client.NewClient(stdioTransport)
	initRequest := mcp.InitializeRequest{}
	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		_ = stdioTransport.Close()
		log.Printf("failed to initialize stdio client: %v", err)
	}

	// List available tools
	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("failed to list tools: %v", err)
	}

	log.Printf("Available tools: %v", toolsResult.Tools)
}

func TestMcpClientRunTool(t *testing.T) {
	command := "D:\\work_szsm\\20250603\\oadin\\internal\\hardware\\installer\\runtime\\bun.exe"
	args := []string{"x", "-y", "bing-cn-mcp"}
	stdioTransport := transport.NewStdio(
		command,
		nil,
		args...,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := stdioTransport.Start(ctx); err != nil {
		log.Printf("failed to start stdio transport: %v", err)
	}

	c := client.NewClient(stdioTransport)
	initRequest := mcp.InitializeRequest{}
	_, err := c.Initialize(ctx, initRequest)
	if err != nil {
		_ = stdioTransport.Close()
		log.Printf("failed to initialize stdio client: %v", err)
	}

	// List available tools
	toolsResult, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("failed to list tools: %v", err)
	}

	log.Printf("Available tools: %v", toolsResult.Tools)
	for _, tool := range toolsResult.Tools {
		log.Printf("Tool: %s, Description: %s", tool.Name, tool.Description)
		log.Printf("Tool Parameters: Type=%v, Properties=%v, Required=%v", tool.InputSchema.Type, tool.InputSchema.Properties, tool.InputSchema.Required)
	}

	toolCallRequestParams := make(map[string]any)
	if err := json.Unmarshal([]byte(`{"query": "hello"}`), &toolCallRequestParams); err != nil {
		log.Printf("failed to unmarshal tool call request params: %v", err)
	}

	callRequest := mcp.CallToolRequest{}
	// 大模型会返回这个参数
	callRequest.Params.Name = "bing_search"
	callRequest.Params.Arguments = toolCallRequestParams

	mcpResult, err := c.CallTool(ctx, callRequest)
	if err != nil {
		log.Printf("failed to call tool: %v", err)
	}

	result := &mcp.CallToolResult{
		IsError: mcpResult.IsError,
	}

	// Process content items
	if len(mcpResult.Content) > 0 {
		var validContents []mcp.Content

		for _, content := range mcpResult.Content {
			// Skip null content
			if content == nil {
				continue
			}

			// Try to get content type
			contentType := ""
			switch c := content.(type) {
			case *mcp.TextContent:
				contentType = "text"
				validContents = append(validContents, &mcp.TextContent{
					Type: "text",
					Text: c.Text,
				})
			case *mcp.ImageContent:
				contentType = "image"
				validContents = append(validContents, &mcp.ImageContent{
					Type:     "image",
					Data:     c.Data,
					MIMEType: c.MIMEType,
				})
			case *mcp.AudioContent:
				contentType = "audio"
				validContents = append(validContents, &mcp.AudioContent{
					Type:     "audio",
					Data:     c.Data,
					MIMEType: c.MIMEType,
				})
			default:
				// Try to parse from raw content
				rawContent, err := json.Marshal(content)
				if err == nil {
					var contentMap map[string]interface{}
					if json.Unmarshal(rawContent, &contentMap) == nil {
						if typ, ok := contentMap["type"].(string); ok {
							contentType = typ

							switch contentType {
							case "text":
								if text, ok := contentMap["text"].(string); ok {
									validContents = append(validContents, &mcp.TextContent{
										Type: "text",
										Text: text,
									})
								}
							case "image":
								data, _ := contentMap["data"].(string)
								mimeType, _ := contentMap["mimeType"].(string)
								validContents = append(validContents, &mcp.ImageContent{
									Type:     "image",
									Data:     data,
									MIMEType: mimeType,
								})
							case "audio":
								data, _ := contentMap["data"].(string)
								mimeType, _ := contentMap["mimeType"].(string)
								validContents = append(validContents, &mcp.AudioContent{
									Type:     "audio",
									Data:     data,
									MIMEType: mimeType,
								})
							}
						}
					}
				}
			}
		}

		if len(validContents) > 0 {
			result.Content = validContents
		}
	}

	log.Printf("Tool call result: %+v", result.Content)
}
