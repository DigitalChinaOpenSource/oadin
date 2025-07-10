package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"oadin/config"
	"oadin/internal/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/provider/engine"
	"oadin/internal/types"
	"oadin/internal/utils/bcode"

	"github.com/google/uuid"
)

type Playground interface {
	CreateSession(ctx context.Context, request *dto.CreateSessionRequest) (*dto.CreateSessionResponse, error)
	GetSessions(ctx context.Context) (*dto.GetSessionsResponse, error)
	SendMessage(ctx context.Context, request *dto.SendMessageRequest) (*dto.SendMessageResponse, error)
	GetMessages(ctx context.Context, request *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error)
	DeleteSession(ctx context.Context, request *dto.DeleteSessionRequest) (*dto.DeleteSessionResponse, error)
	ChangeSessionModel(ctx context.Context, req *dto.ChangeSessionModelRequest) (*dto.ChangeSessionModelResponse, error)
	ToggleThinking(ctx context.Context, req *dto.ToggleThinkingRequest) (*dto.ToggleThinkingResponse, error)

	SendMessageStream(ctx context.Context, request *dto.SendStreamMessageRequest) (chan *types.ChatResponse, chan error)
	UpdateToolCall(ctx context.Context, toolMessage *types.ToolMessage) error
	HandleToolCalls(ctx context.Context, sessionId string, messageId string) []map[string]string
	UpdateSessionTitle(ctx context.Context, sessionID string) error

	UploadFile(ctx context.Context, request *dto.UploadFileRequest, fileHeader io.Reader, filename string, filesize int64) (*dto.UploadFileResponse, error)
	GetFiles(ctx context.Context, request *dto.GetFilesRequest) (*dto.GetFilesResponse, error)
	DeleteFile(ctx context.Context, request *dto.DeleteFileRequest) (*dto.DeleteFileResponse, error)
	ProcessFile(ctx context.Context, request *dto.GenerateEmbeddingRequest) (*dto.GenerateEmbeddingResponse, error)
	CheckEmbeddingService(ctx context.Context, sessionID string) (bool, error)

	findRelevantContext(ctx context.Context, session *types.ChatSession, query string) (string, error)
	findRelevantContextWithVSS(ctx context.Context, session *types.ChatSession, query string, options RAGOptions) (string, error)
}

type PlaygroundImpl struct {
	Ds datastore.Datastore
	Js datastore.JsonDatastore
}

// 创建Playground服务实例
func NewPlayground() Playground {
	playground := &PlaygroundImpl{
		Ds: datastore.GetDefaultDatastore(),
		Js: datastore.GetDefaultJsonDatastore(),
	}
	go func() {
		ctx := context.Background()
		dbPath := config.GlobalOadinEnvironment.Datastore
		if err := InitPlaygroundVec(ctx, dbPath); err != nil {
			slog.Error("初始化VEC失败，将回退到标准向量搜索", "error", err)
		} else {
			if vecInitialized && vecDB != nil {
				slog.Info("VEC初始化成功，已启用向量相似度搜索优化")
			} else if UseVSSForPlayground() {
				slog.Info("VEC扩展未找到，将使用标准向量搜索")
			} else {
				slog.Info("VEC功能已通过环境变量禁用，将使用标准向量搜索")
			}
		}
	}()

	return playground
}

// 工具函数：根据modelId查找modelName
func (p *PlaygroundImpl) GetModelById(ctx context.Context, modelId string) *types.SupportModel {
	// If the modelId is empty, return empty string
	if modelId == "" {
		return &types.SupportModel{}
	}

	model := &types.SupportModel{Id: modelId}
	queryOpList := []datastore.FuzzyQueryOption{}
	queryOpList = append(queryOpList, datastore.FuzzyQueryOption{
		Key:   "id",
		Query: modelId,
	})
	res, err := p.Js.List(ctx, model, &datastore.ListOptions{FilterOptions: datastore.FilterOptions{Queries: queryOpList}})
	if err != nil {
		return &types.SupportModel{}
	}
	if len(res) == 0 {
		return &types.SupportModel{}
	}

	return res[0].(*types.SupportModel)
}

// 创建对话会话
func (p *PlaygroundImpl) CreateSession(ctx context.Context, request *dto.CreateSessionRequest) (*dto.CreateSessionResponse, error) {
	supportModel := p.GetModelById(ctx, request.ModelId)
	if supportModel == nil || supportModel.Id == "" {
		slog.Error("Model not found", "model_id", request.ModelId)
		// ModelName用请求中的值
		if request.ModelName == "" {
			return nil, fmt.Errorf("模型未找到或未指定，请检查模型ID: %s", request.ModelId)
		}
		supportModel = &types.SupportModel{
			Id:   request.ModelId,
			Name: request.ModelName,
		}
	}
	session := &types.ChatSession{
		ID:           uuid.New().String(),
		Title:        request.Title,
		ModelID:      request.ModelId,
		ModelName:    supportModel.Name,
		EmbedModelID: request.EmbedModelId,
	}

	session.ThinkingEnabled = supportModel.Think
	session.ThinkingActive = supportModel.Think // 默认情况下，如果支持深度思考，则启用它

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
			ThinkingActive:  session.ThinkingActive,
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
			ThinkingActive:  session.ThinkingActive,
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
		SortBy: []datastore.SortOption{{Key: "msg_order", Order: datastore.SortOrderAscending}},
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

	// 检查当前会话是否有上传文件，只有有文件时才查找RAG上下文
	fileQuery := &types.File{SessionID: request.SessionId}
	files, err := p.Ds.List(ctx, fileQuery, nil)
	if err != nil {
		slog.Warn("查找会话文件失败，跳过RAG上下文查找", "error", err, "session_id", session.ID)
	}

	enhancedContent := request.Content
	if err == nil && len(files) > 0 {
		slog.Info("开始查找相关RAG上下文", "session_id", session.ID, "question", request.Content)
		relevantContext, err := p.findRelevantContext(ctx, session, request.Content)
		if err != nil {
			slog.Warn("查找相关上下文失败", "error", err, "session_id", session.ID)
		}

		if relevantContext != "" {
			// 添加RAG上下文到用户消息中
			slog.Info("找到相关上下文，使用RAG增强对话", "session_id", session.ID, "context_length", len(relevantContext))
			enhancedContent = fmt.Sprintf("我的问题是: %s\n\n参考以下信息回答我的问题:\n\n%s", request.Content, relevantContext)
		} else {
			slog.Info("未找到相关上下文，使用通用对话模式", "session_id", session.ID)
		}
	} else {
		slog.Info("当前会话无上传文件，跳过RAG上下文查找", "session_id", session.ID)
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
	}

	// 直接调用统一 Engine 层
	chatRequest := &types.ChatRequest{
		Model:    session.ModelName,
		Messages: history,
		Think:    false,
	}
	if session.ThinkingEnabled && session.ThinkingActive {
		chatRequest.Think = true
	}

	if len(request.Tools) > 0 {
		chatRequest.Tools = request.Tools
	}

	slog.Info("发送非流式请求到引擎", "model", session.ModelName)
	chatResp, err := engine.NewEngine().Chat(ctx, chatRequest)
	if err != nil {
		slog.Error("Failed to call model API", "error", err)
		return nil, err
	}

	slog.Info("收到非流式响应",
		"content_length", len(chatResp.Content),
		"model", chatResp.Model,
		"is_complete", chatResp.IsComplete,
		"tool_calls_count", len(chatResp.ToolCalls),
	)

	response := chatResp.Content

	// 如果没有内容但有工具调用，构建提示信息
	if response == "" && len(chatResp.ToolCalls) > 0 {
		slog.Info("模型未生成内容，但有工具调用，构建提示信息", "tool_calls_count", len(chatResp.ToolCalls), chatResp.ToolCalls)
		for _, toolCall := range chatResp.ToolCalls {
			// toolCall.Function.Argument 是map[string]interface{}, 转为json字符串
			arguments, err := json.Marshal(toolCall.Function.Arguments)
			if err != nil {
				slog.Error("工具调用参数序列化失败", "error", err, "arguments", toolCall.Function.Arguments)
			}
			response += fmt.Sprintf("<tool_use>\n  <name>%s</name>\n  <arguments>%s</arguments>\n</tool_use>\n", toolCall.Function.Name, arguments)
		}
	}

	// 保存模型回复
	assistantMsg := &types.ChatMessage{
		ID:            uuid.New().String(),
		SessionID:     request.SessionId,
		Role:          "assistant",
		Content:       response,
		Order:         len(messages) + 1,
		CreatedAt:     time.Now(),
		ModelID:       session.ModelID,
		ModelName:     session.ModelName,
		TotalDuration: chatResp.TotalDuration,
	}
	err = p.Ds.Add(ctx, assistantMsg)
	if err != nil {
		slog.Error("Failed to save assistant message", "error", err)
		return nil, err
	}
	// 保存思考内容
	if chatResp.Thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
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
	resultMessages := []dto.Message{{
		Id:            assistantMsg.ID,
		SessionId:     assistantMsg.SessionID,
		Role:          assistantMsg.Role,
		Content:       assistantMsg.Content,
		CreatedAt:     assistantMsg.CreatedAt.Format(time.RFC3339),
		Thoughts:      chatResp.Thoughts,
		Type:          "answer",
		ModelId:       session.ModelID,
		ModelName:     session.ModelName,
		TotalDuration: chatResp.TotalDuration,
	}}
	if chatResp.Thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
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
		Bcode: bcode.SuccessCode,
		Data:  resultMessages,
	}, nil
}

// 获取会话中的消息
func (p *PlaygroundImpl) GetMessages(ctx context.Context, request *dto.GetMessagesRequest) (*dto.GetMessagesResponse, error) {
	slog.Info("GetMessages called", "session_id", request.SessionId)

	sessionQuery := &types.ChatSession{ID: request.SessionId}
	sessionResults, err := p.Ds.List(ctx, sessionQuery, &datastore.ListOptions{PageSize: 1})
	if err != nil {
		slog.Error("Failed to get session", "error", err, "session_id", request.SessionId)
		return nil, err
	}

	var thinkingActive bool
	if len(sessionResults) > 0 {
		if session, ok := sessionResults[0].(*types.ChatSession); ok {
			thinkingActive = session.ThinkingActive
			slog.Info("Found session", "session_id", request.SessionId, "thinking_active", thinkingActive)
		}
	} else {
		slog.Warn("Session not found", "session_id", request.SessionId)
	}

	messageQuery := &types.ChatMessage{SessionID: request.SessionId}
	messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "msg_order", Order: datastore.SortOrderAscending},
		},
	})
	if err != nil {
		slog.Error("Failed to list chat messages", "error", err)
		return nil, err
	}

	slog.Info("Found messages in database", "session_id", request.SessionId, "count", len(messages))

	messageDTOs := make([]dto.Message, 0, len(messages))
	for _, m := range messages {
		msg := m.(*types.ChatMessage)
		typeStr := ""
		if msg.Role == "assistant" {
			typeStr = "answer"
		} else if msg.Role == "think" && len(msg.Content) > 0 {
			typeStr = "thoughts"
		}

		var toolMessages []types.ToolMessage
		if msg.IsToolGroupID {
			typeStr = "mcp"
			toolMsgQuery := new(types.ToolMessage)
			toolMsgQuery.AssistantMsgID = msg.ID
			toolMsgResults, err := p.Ds.List(ctx, toolMsgQuery, &datastore.ListOptions{
				SortBy: []datastore.SortOption{
					{Key: "updated_at", Order: datastore.SortOrderAscending},
				},
			})
			if err != nil {
				slog.Error("Failed to list tool messages", "error", err)
				return nil, err
			}
			for _, tm := range toolMsgResults {
				if tmsg, ok := tm.(*types.ToolMessage); ok {
					if tmsg.AssistantMsgID == msg.ID {
						toolMessages = append(toolMessages, *tmsg)
					}
				}
			}
		}
		messageDTOs = append(messageDTOs, dto.Message{
			Id:            msg.ID,
			SessionId:     msg.SessionID,
			Role:          msg.Role,
			Content:       msg.Content,
			CreatedAt:     msg.CreatedAt.Format(time.RFC3339),
			Type:          typeStr,
			ModelId:       msg.ModelID,
			ModelName:     msg.ModelName,
			TotalDuration: msg.TotalDuration,
			ToolCalls:     toolMessages,
		})
	}

	return &dto.GetMessagesResponse{
		Bcode: bcode.SuccessCode,
		Data: dto.GetMessagesData{
			Messages:       messageDTOs,
			ThinkingActive: thinkingActive,
		},
	}, nil
}

// 删除会话
func (p *PlaygroundImpl) DeleteSession(ctx context.Context, request *dto.DeleteSessionRequest) (*dto.DeleteSessionResponse, error) {
	// 1. 获取会话记录
	session := &types.ChatSession{ID: request.SessionId}
	err := p.Ds.Get(ctx, session)
	if err != nil {
		slog.Error("Failed to get chat session", "error", err, "session_id", request.SessionId)
		return nil, err
	}

	// 2. 删除会话相关的消息记录
	messageQuery := &types.ChatMessage{SessionID: request.SessionId}
	messages, err := p.Ds.List(ctx, messageQuery, nil)
	if err != nil {
		slog.Error("Failed to list chat messages", "error", err)
	} else {
		for _, m := range messages {
			msg := m.(*types.ChatMessage)
			err = p.Ds.Delete(ctx, msg)
			if err != nil {
				slog.Error("Failed to delete chat message", "error", err, "message_id", msg.ID)
			}
		}
	}

	// 3. 删除会话相关的文件记录
	fileQuery := &types.File{SessionID: request.SessionId}
	files, err := p.Ds.List(ctx, fileQuery, nil)
	if err != nil {
		slog.Error("Failed to list files", "error", err)
	} else {
		for _, f := range files {
			file := f.(*types.File)

			chunkQuery := &types.FileChunk{FileID: file.ID}
			chunks, err := p.Ds.List(ctx, chunkQuery, nil)
			if err != nil {
				slog.Error("Failed to list file chunks", "error", err)
			} else {
				if vecInitialized {
					if vecDB != nil {
						var chunkIDs []string
						for _, c := range chunks {
							chunk := c.(*types.FileChunk)
							chunkIDs = append(chunkIDs, chunk.ID)
						}
						if len(chunkIDs) > 0 {
							err := vecDB.DeleteChunks(ctx, chunkIDs)
							if err != nil {
								slog.Error("从VEC删除文件块失败", "error", err, "file_id", file.ID)
							}
						}
					}
				}

				for _, c := range chunks {
					chunk := c.(*types.FileChunk)
					// 删除文件块记录
					err = p.Ds.Delete(ctx, chunk)
					if err != nil {
						slog.Error("Failed to delete file chunk", "error", err)
					}
				}
			}

			// 删除文件记录
			err = p.Ds.Delete(ctx, file)
			if err != nil {
				slog.Error("Failed to delete file", "error", err)
			}
		}
	}

	// 4. 删除会话记录
	err = p.Ds.Delete(ctx, session)
	if err != nil {
		slog.Error("Failed to delete chat session", "error", err)
		return nil, err
	}

	return &dto.DeleteSessionResponse{
		Bcode: bcode.SuccessCode,
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
		modelInfo := p.GetModelById(ctx, req.ModelId)
		session.ModelName = modelInfo.Name
		// 根据 modelId 查询模型属性，自动赋值 thinkingEnabled
		model := &types.Model{ModelName: modelInfo.Name}
		err := p.Ds.Get(ctx, model)
		if err == nil {
			session.ThinkingEnabled = modelInfo.Think
			// 切换模型时，如果新模型支持思考，则保持当前思考状态；如果不支持，则关闭思考
			if !modelInfo.Think {
				session.ThinkingActive = false
			}
		}
	}
	if req.EmbedModelId != "" {
		session.EmbedModelID = req.EmbedModelId
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
			ThinkingActive:  session.ThinkingActive,
			CreatedAt:       session.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       session.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// 切换思考状态
func (p *PlaygroundImpl) ToggleThinking(ctx context.Context, req *dto.ToggleThinkingRequest) (*dto.ToggleThinkingResponse, error) {
	// 查找会话
	session := &types.ChatSession{ID: req.SessionId}
	err := p.Ds.Get(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("会话不存在: %v", err)
	}
	// 首先检查模型是否支持深度思考
	if !session.ThinkingEnabled {
		return nil, fmt.Errorf("当前模型不支持深度思考功能")
	}

	// 根据请求切换深度思考状态
	if req.Enabled != nil {
		// 如果提供了明确的启用/禁用值，则使用该值
		session.ThinkingActive = *req.Enabled
	} else {
		// 如果没有提供值，则切换当前状态
		session.ThinkingActive = !session.ThinkingActive
	}

	// 更新会话
	session.UpdatedAt = time.Now()
	err = p.Ds.Put(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("更新会话失败: %v", err)
	}
	// 返回更新后的状态
	return &dto.ToggleThinkingResponse{
		Bcode:          bcode.SuccessCode,
		ThinkingActive: session.ThinkingActive,
	}, nil
}

func (p *PlaygroundImpl) findRelevantContextWithVSS(ctx context.Context, session *types.ChatSession, query string, options RAGOptions) (string, error) {
	return "", nil
}

func (p *PlaygroundImpl) CheckEmbeddingService(ctx context.Context, sessionID string) (bool, error) {
	return true, nil
}

func InitPlaygroundVec(ctx context.Context, dbPath string) error {

	return initVecDB(dbPath)
}

func UseVSSForPlayground() bool {
	return false
}
