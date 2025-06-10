package mcp_handler

import (
	"byze/internal/hardware"
	"byze/internal/types"
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestMcpClientTest(t *testing.T) {
	server := types.MCPServerConfig{
		Name:    "bingcn",
		Args:    []string{"bing-cn-mcp"},
		Env:     map[string]string{},
		Command: "npx",
	}
	config := types.McpUserConfig{
		ID:        1,
		MCPID:     "1",
		Kits:      "",
		UserID:    "1",
		Status:    1,
		Auth:      "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	log.Printf("MCP Server Config: %+v", server)
	log.Printf("MCP User Config: %+v", config)

	commandBuilder := hardware.NewCommandBuilder(server.Command).WithArgs(server.Args...)
	cmd, _ := commandBuilder.GetRunCommand()
	log.Printf("Command to run: %s", cmd)
}

// 这个可以跑通了
func TestMcpClientListTools(t *testing.T) {
	command := "D:\\work_szsm\\20250603\\byze\\internal\\hardware\\installer\\runtime\\bun.exe"
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
	command := "D:\\work_szsm\\20250603\\byze\\internal\\hardware\\installer\\runtime\\bun.exe"
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
