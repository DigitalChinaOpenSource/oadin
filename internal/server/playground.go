package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"time"

	"byze/config"
	"byze/internal/api/dto"
	"byze/internal/datastore"
	"byze/internal/provider"
	"byze/internal/rpc"
	"byze/internal/server/mcp_handler"
	"byze/internal/types"
	"byze/internal/utils/bcode"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
)

type PlaygroundImpl struct {
	Ds datastore.Datastore
}

// 创建Playground服务实例
func NewPlayground() *PlaygroundImpl {
	playground := &PlaygroundImpl{
		Ds: datastore.GetDefaultDatastore(),
	}
	go func() {
		ctx := context.Background()
		dbPath := config.GlobalByzeEnvironment.Datastore
		if err := InitPlaygroundVSS(ctx, dbPath); err != nil {
			slog.Error("初始化VSS失败，将回退到标准向量搜索", "error", err)
		} else {
			slog.Info("VSS初始化成功，已启用向量相似度搜索优化")
		}
	}()

	return playground
}

// 工具函数：根据modelId查找modelName
func getModelNameById(modelId string) string {
	// 1. 先查本地模型json
	data, err := ioutil.ReadFile("internal/provider/template/local_model.json")
	if err == nil {
		var local struct {
			Chat []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"chat"`
			Embed []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"embed"`
		}
		if json.Unmarshal(data, &local) == nil {
			for _, m := range local.Chat {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
			for _, m := range local.Embed {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
		}
	}
	// 2. 查support_model.json
	data, err = ioutil.ReadFile("internal/datastore/jsonds/data/support_model.json")
	if err == nil {
		var arr []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &arr) == nil {
			for _, m := range arr {
				if m.ID == modelId || m.Name == modelId {
					return m.Name
				}
			}
		}
	}
	return ""
}

// 创建对话会话
func (p *PlaygroundImpl) CreateSession(ctx context.Context, request *dto.CreateSessionRequest) (*dto.CreateSessionResponse, error) {
	session := &types.ChatSession{
		ID:              uuid.New().String(),
		Title:           request.Title,
		ModelID:         request.ModelId,
		ModelName:       getModelNameById(request.ModelId),
		EmbedModelID:    request.EmbedModelId,
		ThinkingEnabled: request.ThinkingEnabled,
	}

	err := p.Ds.Add(ctx, session)
	if err != nil {
		slog.Error("Failed to create chat session", "error", err)
		return nil, err
	}
	return &dto.CreateSessionResponse{
		Bcode: bcode.SuccessCode,
		Data: dto.Session{
			Id:              session.ID,
			Title:           session.Title,
			ModelId:         session.ModelID,
			ModelName:       session.ModelName,
			EmbedModelId:    session.EmbedModelID,
			ThinkingEnabled: session.ThinkingEnabled,
			CreatedAt:       session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetSessions 获取所有会话
func (p *PlaygroundImpl) GetSessions(ctx context.Context) (*dto.GetSessionsResponse, error) {
	sessionQuery := &types.ChatSession{}
	sessions, err := p.Ds.List(ctx, sessionQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "created_at", Order: datastore.SortOrderDescending},
		},
	})
	if err != nil {
		slog.Error("Failed to list chat sessions", "error", err)
		return nil, err
	}

	sessionDTOs := make([]dto.Session, 0, len(sessions))
	for _, s := range sessions {
		session := s.(*types.ChatSession)
		sessionDTOs = append(sessionDTOs, dto.Session{
			Id:              session.ID,
			Title:           session.Title,
			ModelId:         session.ModelID,
			ModelName:       session.ModelName,
			EmbedModelId:    session.EmbedModelID,
			ThinkingEnabled: session.ThinkingEnabled,
			CreatedAt:       session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
		})
	}

	return &dto.GetSessionsResponse{
		Bcode: bcode.SuccessCode,
		Data:  sessionDTOs,
	}, nil
}

// 发送消息并获取回复
func (p *PlaygroundImpl) SendMessage(ctx context.Context, request *dto.SendMessageRequest) (*dto.SendMessageResponse, error) {
	// 获取会话
	session := &types.ChatSession{ID: request.SessionId}
	err := p.Ds.Get(ctx, session)
	if err != nil {
		slog.Error("Failed to get chat session", "error", err)
		return nil, err
	}

	// 获取会话中的所有消息，构建历史上下文
	messageQuery := &types.ChatMessage{SessionID: request.SessionId}
	messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "order", Order: datastore.SortOrderAscending},
		},
	})
	if err != nil {
		slog.Error("Failed to list chat messages", "error", err)
		return nil, err
	}

	// 构建历史对话
	history := make([]map[string]string, 0, len(messages)+1)
	for _, m := range messages {
		msg := m.(*types.ChatMessage)
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	// 添加RAG上下文
	enhancedContent := request.Content
	relevantContext, err := p.findRelevantContext(ctx, session, request.Content)
	if err != nil {
		slog.Warn("查找相关上下文失败", "error", err)
	}

	if relevantContext != "" {
		// 添加RAG上下文到用户消息中
		slog.Info("找到相关上下文，使用RAG增强对话", "session_id", session.ID, "context_length", len(relevantContext))
		enhancedContent = fmt.Sprintf("我的问题是: %s\n\n参考以下信息回答我的问题:\n\n%s",
			request.Content, relevantContext)
	} else {
		slog.Info("未找到相关上下文，使用通用对话模式", "session_id", session.ID)
	}

	// 是否启用了mcp服务器
	mcpResults, err := p.HandleMCPToolInvocation(ctx, session.ModelID, request.Content, request.McpTools)
	if len(mcpResults) > 0 && err == nil {
		// 将MCP工具调用结果添加到历史记录中
		for _, result := range mcpResults {
			toolResultText := fmt.Sprintf("MCP工具【%s】执行结果:%s", result.McpTool.Tool.Name, result.Result.Content)
			history = append(history, map[string]string{
				"role":    "system",
				"content": toolResultText,
			})
		}
	}

	// 添加当前用户消息
	userMessage := map[string]string{
		"role":    "user",
		"content": enhancedContent,
	}
	history = append(history, userMessage)

	// 保存用户消息到数据库
	userMsg := &types.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: request.SessionId,
		Role:      "user",
		Content:   request.Content,
		Order:     len(messages),
		CreatedAt: time.Now(),
		ModelID:   session.ModelID,
		ModelName: session.ModelName,
	}
	err = p.Ds.Add(ctx, userMsg)
	if err != nil {
		slog.Error("Failed to save user message", "error", err)
		return nil, err
	} // 调用模型获取回复
	engineName := "ollama" // 默认使用Ollama引擎

	// 获取当前模型的引擎
	modelEngine := provider.GetModelEngine(engineName)
	// 构建聊天请求
	chatRequest := &types.ChatRequest{
		Model:    session.ModelID,
		Messages: history,
		Options:  make(map[string]any),
	}

	// 如果启用了思考模式，则添加thinking选项
	if session.ThinkingEnabled {
		chatRequest.Options["thinking"] = true
	}
	// 调用模型API
	chatResp, err := modelEngine.Chat(ctx, chatRequest)
	if err != nil {
		slog.Error("Failed to call model API", "error", err)
		return nil, err
	}
	response := chatResp.Content
	// 保存模型回复
	assistantMsg := &types.ChatMessage{
		ID:        uuid.New().String(),
		SessionID: request.SessionId,
		Role:      "assistant",
		Content:   response,
		Order:     len(messages) + 1,
		CreatedAt: time.Now(),
		ModelID:   session.ModelID,
		ModelName: session.ModelName,
	}
	err = p.Ds.Add(ctx, assistantMsg)
	if err != nil {
		slog.Error("Failed to save assistant message", "error", err)
		return nil, err
	}
	// 保存思考内容
	if chatResp.Thoughts != "" && session.ThinkingEnabled {
		thoughtsMsg := &types.ChatMessage{
			ID:        uuid.New().String(),
			SessionID: request.SessionId,
			Role:      "system",
			Content:   "思考过程: " + chatResp.Thoughts,
			Order:     len(messages) + 2,
			CreatedAt: time.Now(),
			ModelID:   session.ModelID,
			ModelName: session.ModelName,
		}
		err = p.Ds.Add(ctx, thoughtsMsg)
		if err != nil {
			slog.Error("Failed to save thoughts message", "error", err)
			// 非致命错误，继续执行
		}
	}

	// 如果是第一条消息，更新会话标题
	if len(messages) == 0 {
		title := "新对话 " + time.Now().Format("2006-01-02")
		session.Title = title
		err = p.Ds.Put(ctx, session)
		if err != nil {
			slog.Error("Failed to update session title", "error", err)
		}
	}
	// 返回响应
	resultMessages := []dto.Message{
		{
			Id:        assistantMsg.ID,
			SessionId: assistantMsg.SessionID,
			Role:      assistantMsg.Role,
			Content:   assistantMsg.Content,
			CreatedAt: assistantMsg.CreatedAt.Format(time.RFC3339),
			Thoughts:  chatResp.Thoughts,
			Type:      "answer",
			ModelId:   session.ModelID,
			ModelName: session.ModelName,
		},
	}
	if chatResp.Thoughts != "" && session.ThinkingEnabled {
		resultMessages = append(resultMessages, dto.Message{
			Id:        "thoughts-" + assistantMsg.ID,
			SessionId: assistantMsg.SessionID,
			Role:      "system",
			Content:   chatResp.Thoughts,
			CreatedAt: assistantMsg.CreatedAt.Format(time.RFC3339),
			Type:      "thoughts",
			ModelId:   session.ModelID,
			ModelName: session.ModelName,
		})
	}
	return &dto.SendMessageResponse{
		Bcode:      bcode.SuccessCode,
		Data:       resultMessages,
		McpResults: mcpResults,
	}, nil
}

// 获取会话中的消息
func (p *PlaygroundImpl) GetMessages(ctx context.Context, request *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error) {
	messageQuery := &types.ChatMessage{SessionID: request.SessionId}
	messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "order", Order: datastore.SortOrderAscending},
		},
	})
	if err != nil {
		slog.Error("Failed to list chat messages", "error", err)
		return nil, err
	}

	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, m := range messages {
		msg := m.(*types.ChatMessage)
		typeStr := ""
		if msg.Role == "assistant" {
			typeStr = "answer"
		} else if msg.Role == "system" && len(msg.Content) > 0 && (msg.Content[:12] == "思考过程: " || msg.Content[:9] == "Thoughts:") {
			typeStr = "thoughts"
		}
		messageDTOs = append(messageDTOs, dto.Message{
			Id:        msg.ID,
			SessionId: msg.SessionID,
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
			Type:      typeStr,
			ModelId:   msg.ModelID,
			ModelName: msg.ModelName,
		})
	}

	return &dto.GetMessagesResponse{
		Bcode: bcode.SuccessCode,
		Data:  messageDTOs,
	}, nil
}

// 切换会话模型
func (p *PlaygroundImpl) ChangeSessionModel(ctx context.Context, req *dto.ChangeSessionModelRequest) (*dto.ChangeSessionModelResponse, error) {
	session := &types.ChatSession{ID: req.SessionId}
	err := p.Ds.Get(ctx, session)
	if err != nil {
		return nil, err
	}
	if req.ModelId != "" {
		session.ModelID = req.ModelId
		session.ModelName = getModelNameById(req.ModelId)
	}
	if req.EmbedModelId != "" {
		session.EmbedModelID = req.EmbedModelId
	}
	if req.ThinkingEnabled != nil {
		session.ThinkingEnabled = *req.ThinkingEnabled
	}
	session.UpdatedAt = time.Now()
	err = p.Ds.Put(ctx, session)
	if err != nil {
		return nil, err
	}
	return &dto.ChangeSessionModelResponse{
		Bcode: bcode.SuccessCode,
		Data: &dto.Session{
			Id:              session.ID,
			Title:           session.Title,
			ModelId:         session.ModelID,
			ModelName:       session.ModelName,
			EmbedModelId:    session.EmbedModelID,
			ThinkingEnabled: session.ThinkingEnabled,
			CreatedAt:       session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// 处理消息中的mcp工具调用
func (p *PlaygroundImpl) HandleMCPToolInvocation(ctx context.Context, model, query string, mcpTools []dto.McpTool) ([]dto.McpToolResult, error) {
	if len(mcpTools) == 0 {
		return nil, nil
	}
	var results []dto.McpToolResult

	engineName := "ollama" // 默认使用Ollama引擎
	modelEngine := provider.GetModelEngine(engineName)

	userMessage := map[string]string{
		"role":    "user",
		"content": query,
	}
	chatRequest := &types.ChatRequest{
		Model:    model,
		Messages: []map[string]string{userMessage},
		Tools:    make([]map[string]any, 0, len(mcpTools)),
	}
	for _, mcpTool := range mcpTools {
		chatRequest.Tools = append(chatRequest.Tools, map[string]any{
			"type":     "function",
			"function": mcpTool.Tool,
		})
	}

	// 发起带tools的chat请求
	chatResp, err := modelEngine.Chat(ctx, chatRequest)
	if err != nil {
		return nil, err
	}

	// 解析chatResp中的工具调用结果（假设返回内容中包含tool_calls字段）
	var toolCalls []struct {
		Function struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		} `json:"function"`
	}
	// 兼容不同模型返回格式
	if chatResp != nil && chatResp.Content != "" {
		var respMap map[string]interface{}
		// 这里的 Content 对应api的 message
		if err := json.Unmarshal([]byte(chatResp.Content), &respMap); err == nil {
			if tc, ok := respMap["tool_calls"]; ok {
				if tcArr, ok := tc.([]interface{}); ok {
					for _, t := range tcArr {
						b, _ := json.Marshal(t)
						var call struct {
							Function struct {
								Name      string         `json:"name"`
								Arguments map[string]any `json:"arguments"`
							} `json:"function"`
						}
						if err := json.Unmarshal(b, &call); err == nil {
							toolCalls = append(toolCalls, call)
						}
					}
				}
			}
		}
	}

	// 将toolCalls转为rpc.ClientRunToolRequest格式
	if len(toolCalls) == 0 {
		slog.Warn("No tool calls found in chat response", "response", chatResp.Content)
		return nil, nil
	}
	clientRunToolRequests := make([]rpc.ClientRunToolRequest, 0, len(toolCalls))
	for _, call := range toolCalls {
		for _, mcpTool := range mcpTools {
			if mcpTool.Tool.Name == call.Function.Name {
				clientRunToolRequests = append(clientRunToolRequests, rpc.ClientRunToolRequest{
					MCPId:    mcpTool.MCPId,
					ToolName: call.Function.Name,
					ToolArgs: call.Function.Arguments,
				})
				break
			}
		}
	}

	mcpHandler := mcp_handler.NewMcpService()
	for _, req := range clientRunToolRequests {
		// 调用MCP服务器的工具
		mcpResult, err := mcpHandler.CallTool(req.MCPId, mcp.CallToolParams{
			Name:      req.ToolName,
			Arguments: req.ToolArgs,
		})
		if err != nil {
			slog.Error("Failed to call MCP tool", "error", err, "mcpId", req.MCPId, "toolName", req.ToolName)
			continue // 继续处理其他工具调用
		}

		// 构建结果
		for _, mcpTool := range mcpTools {
			if mcpTool.MCPId == req.MCPId {
				results = append(results, dto.McpToolResult{
					McpTool:  mcpTool,
					ToolArgs: req.ToolArgs,
					Result:   *mcpResult,
				})
				break
			}
		}
	}
	return results, nil
}
