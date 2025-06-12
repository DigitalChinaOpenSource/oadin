package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"byze/internal/api/dto"
	"byze/internal/datastore"
	"byze/internal/provider/engine"
	"byze/internal/types"

	"github.com/google/uuid"
)

// 发送消息并流式返回响应
func (p *PlaygroundImpl) SendMessageStream(ctx context.Context, request *dto.SendStreamMessageRequest) (chan *types.ChatResponse, chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		// 获取会话
		session := &types.ChatSession{ID: request.SessionID}
		err := p.Ds.Get(ctx, session)
		if err != nil {
			slog.Error("Failed to get chat session", "error", err)
			errChan <- err
			return
		}

		// 获取会话中的所有消息，构建历史上下文
		messageQuery := &types.ChatMessage{SessionID: request.SessionID}
		messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
			SortBy: []datastore.SortOption{
				{Key: "msg_order", Order: datastore.SortOrderAscending},
			},
		})
		if err != nil {
			slog.Error("Failed to list chat messages", "error", err)
			errChan <- err
			return
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

		// 添加当前用户消息
		userMessage := map[string]string{
			"role":    "user",
			"content": enhancedContent,
		}
		history = append(history, userMessage)

		// 保存用户消息到数据库（保存原始消息，不含上下文）
		userMsg := &types.ChatMessage{
			ID:        uuid.New().String(),
			SessionID: request.SessionID,
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
			errChan <- err
			return
		}
		// 调用模型获取流式回复
		// 获取统一的聊天引擎
		modelEngine := engine.NewEngine()
		// 构建聊天请求
		chatRequest := &types.ChatRequest{
			Model:    session.ModelID,
			Messages: history,
			Stream:   true, // 启用流式输出
			Options:  make(map[string]any),
		}

		// 如果启用了思考模式，则添加thinking选项
		if session.ThinkingEnabled {
			chatRequest.Options["thinking"] = true
		}

		// Chat request (with tools)
		if request.Tools != nil {
			chatRequest.Tools = request.Tools
		}

		// 调用流式API
		responseStream, streamErrChan := modelEngine.ChatStream(ctx, chatRequest)

		// 创建会话消息ID
		assistantMsgID := uuid.New().String()
		var fullContent string
		var thoughts string

		// 转发流式响应
		for {
			select {
			case resp, ok := <-responseStream:
				if !ok {
					// 流结束
					// 保存完整的助手回复
					if fullContent != "" {
						assistantMsg := &types.ChatMessage{
							ID:        assistantMsgID,
							SessionID: request.SessionID,
							Role:      "assistant",
							Content:   fullContent,
							Order:     len(messages) + 1,
							CreatedAt: time.Now(),
							ModelID:   session.ModelID,
							ModelName: session.ModelName,
						}
						err = p.Ds.Add(ctx, assistantMsg)
						if err != nil {
							slog.Error("Failed to save assistant message", "error", err)
						} // 如果是第一条消息，更新会话标题
						if len(messages) == 0 {
							title := "新对话 " + time.Now().Format("2006-01-02")
							session.Title = title
							err = p.Ds.Put(ctx, session)
							if err != nil {
								slog.Error("Failed to update session title", "error", err)
							}
						}

						// 保存思考内容（如果有）
						if thoughts != "" && session.ThinkingEnabled {
							thoughtsMsg := &types.ChatMessage{
								ID:        uuid.New().String(),
								SessionID: request.SessionID,
								Role:      "system",
								Content:   "思考过程: " + thoughts,
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
					}
					return
				}
				msgType := "answer"
				if resp.Thoughts != "" {
					msgType = "thoughts"
				}
				resp.Type = msgType
				resp.Model = session.ModelID
				resp.ModelName = session.ModelName
				respChan <- resp

				// 累积完整内容
				fullContent += resp.Content
				if resp.Thoughts != "" {
					thoughts = resp.Thoughts
				}

			case err, ok := <-streamErrChan:
				if !ok {
					return
				}
				// 转发错误
				errChan <- err
				return
			case <-ctx.Done():
				// 上下文取消
				errChan <- ctx.Err()
				return
			}
		}
	}()

	return respChan, errChan
}
