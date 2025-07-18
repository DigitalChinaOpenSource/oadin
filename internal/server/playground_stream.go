package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"oadin/internal/api/dto"
	"oadin/internal/datastore"
	"oadin/internal/provider/engine"
	"oadin/internal/types"

	"github.com/google/uuid"
)

// 发送消息并流式返回响应
func (p *PlaygroundImpl) SendMessageStream(ctx context.Context, request *dto.SendStreamMessageRequest) (chan *types.ChatResponse, chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		sendError := func(err error) {
			errChan <- err
		}

		// 获取会话
		session := &types.ChatSession{ID: request.SessionID}
		err := p.Ds.Get(ctx, session)
		if err != nil {
			slog.Error("Failed to get chat session", "error", err)
			sendError(err)
			return
		}

		// 增加会话标题
		if session.Title == "" {
			p.AddSessionTitle(ctx, request)
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
			sendError(err)
			return
		}

		// 构建历史对话
		history := make([]map[string]string, 0, len(messages)+1)
		for _, m := range messages {
			msg := m.(*types.ChatMessage)
			if msg.Role == "assistant" {
				// 把thinking内容清理掉
				if strings.Contains(msg.Content, "<think>") && strings.Contains(msg.Content, "</think>") {
					re := regexp.MustCompile(`(?s)<think>.*?</think>\s*`)
					msg.Content = re.ReplaceAllString(msg.Content, "")
					msg.Content = strings.TrimSpace(msg.Content)
				}

				if msg.Content == "" {
					continue
				}
			}
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

		// 保存用户消息到数据库（保存原始消息，不含上下文）
		var userMsg *types.ChatMessage
		if request.Content != "" {
			// 添加当前用户消息
			userMessage := map[string]string{
				"role":    "user",
				"content": enhancedContent,
			}
			history = append(history, userMessage)

			userMsg = &types.ChatMessage{
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
				sendError(err)
				return
			}
		}

		// 处理多轮工具调用
		if request.ToolGroupID != "" {
			toolList := p.HandleToolCalls(ctx, request.SessionID, request.ToolGroupID)
			history = append(history, toolList...)
		}

		// 调用模型获取流式回复
		// 获取统一的聊天引擎
		modelEngine := engine.NewEngine() // 构建聊天请求
		chatRequest := &types.ChatRequest{
			Model:    session.ModelName,
			Messages: history,
			Stream:   true,  // 启用流式输出
			Think:    false, // 默认不启用深度思考
		}
		// 如果模型支持深度思考且用户启用了思考模式，则添加thinking选项
		if session.ThinkingEnabled && session.ThinkingActive {
			chatRequest.Think = true
		}

		// Chat request (with tools)
		if len(request.Tools) > 0 {
			chatRequest.Tools = request.Tools
		}

		// 调用流式API
		responseStream, streamErrChan := modelEngine.ChatStream(ctx, chatRequest)

		// 创建会话消息ID
		assistantMsgID := uuid.New().String()
		var fullContent string
		var thoughts string
		var totalDuration int64
		// 转发流式响应
		for {
			select {
			case resp, ok := <-responseStream:
				if !ok {
					return
				}

				// 保留原始的 Content
				originalContent := resp.Content
				resp.Type = "answer"
				resp.Model = session.ModelName
				resp.ModelName = session.ModelName

				if resp.IsComplete {
					fmt.Printf("[PlaygroundStream] 收到流式输出完成标记，原始内容长度: %d，累积内容长度: %d\n", len(originalContent), len(fullContent))
					fmt.Printf("[PlaygroundStream] 原始内容: '%s'，累积内容: '%s'\n", originalContent, fullContent)
					slog.Info("收到流式输出完成标记",
						"is_complete", resp.IsComplete,
						"content_length", len(originalContent),
						"accumulated_content_length", len(fullContent))

					// 收集思考内容（如果有）
					if resp.Thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
						thoughts = thoughts + resp.Thoughts
					}

					// 确保使用累积的完整内容
					if len(fullContent) == 0 && len(originalContent) > 0 {
						fullContent = originalContent
						fmt.Printf("[PlaygroundStream] 使用原始内容作为完整内容: '%s'\n", fullContent)
					}

					resp.TotalDuration = resp.TotalDuration / int64(time.Second)
					totalDuration = resp.TotalDuration

					// 将思考内容包装在<think></think>标签中并添加到assistant响应
					finalContent := fullContent
					if thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
						// 在正文前添加思考内容，使用<think>标签包装
						finalContent = fmt.Sprintf("<think>%s\n</think>\n%s", thoughts, fullContent)
					}

					// 确保完整内容被保存和返回给客户端
					fmt.Printf("[PlaygroundStream] 准备保存数据库，finalContent长度: %d, 内容: '%s'\n", len(finalContent), finalContent)
					if finalContent != "" {
						assistantMsg := &types.ChatMessage{
							ID:            assistantMsgID,
							SessionID:     request.SessionID,
							Role:          "assistant",
							Content:       finalContent,
							Order:         len(messages) + 1,
							CreatedAt:     time.Now(),
							ModelID:       session.ModelID,
							ModelName:     session.ModelName,
							TotalDuration: totalDuration,
						}
						err = p.Ds.Add(ctx, assistantMsg)
						if err != nil {
							slog.Error("Failed to save assistant message on complete", "error", err)
						} else {
							fmt.Printf("[PlaygroundStream] 成功保存assistant消息到数据库，ID: %s, 内容长度: %d\n", assistantMsgID, len(finalContent))
						}
					} else {
						fmt.Printf("[PlaygroundStream] 警告：finalContent为空，不保存到数据库！fullContent='%s', thoughts='%s'\n", fullContent, thoughts)
					}

					// 处理多轮工具调用
					if len(resp.ToolCalls) > 0 {
						if request.ToolGroupID == "" {
							if userMsg != nil {
								p.AddToolCall(ctx, request.SessionID, userMsg.ID, assistantMsgID, resp.TotalDuration)
							}
						} else {
							p.AddToolCall(ctx, request.SessionID, request.ToolGroupID, assistantMsgID, resp.TotalDuration)
						}
					}

					// 发送全部内容作为响应
					resp.Content = fullContent // 注意：UI显示仍使用原始内容，不包含思考部分
					resp.ID = assistantMsgID
					if userMsg != nil && len(resp.ToolCalls) > 0 && request.ToolGroupID == "" {
						resp.ToolGroupID = userMsg.ID
					}
					if userMsg == nil && len(resp.ToolCalls) > 0 && request.ToolGroupID != "" {
						resp.ToolGroupID = request.ToolGroupID
					}
					if len(resp.ToolCalls) == 0 && request.ToolGroupID != "" {
						p.MarkToolCallEnd(ctx, request.SessionID, request.ToolGroupID, assistantMsgID)
					}

					fmt.Println("流式输出完成，保存助手回复", resp)
					slog.Info("流式输出完成，保存助手回复", resp)
					respChan <- resp
				} else if len(originalContent) > 0 {
					// fmt.Printf("[PlaygroundStream] 累积内容块: '%s', 当前fullContent长度: %d\n", originalContent, len(fullContent))
					fullContent += resp.Content
					// fmt.Printf("[PlaygroundStream] 累积后fullContent长度: %d\n", len(fullContent))
					respChan <- resp
				} else if resp.Thoughts != "" {
					// 收集思考内容，但不再单独存储
					thoughts += resp.Thoughts
					if resp.TotalDuration > 0 {
						resp.TotalDuration = resp.TotalDuration / int64(time.Second)
					}
					respChan <- resp
				} else {
					slog.Debug("跳过空内容块")
					continue
				}

			case err, ok := <-streamErrChan:
				if !ok {
					return
				}
				slog.Error("[PlaygroundStream] 收到streamErrChan错误", "error", err.Error())
				respChan <- &types.ChatResponse{
					Type:       "error",
					Content:    err.Error(),
					IsComplete: true,
					ID:         assistantMsgID,
				}
				return
			case err, ok := <-errChan:
				if !ok {
					return
				}
				slog.Error("[PlaygroundStream] 收到errChan错误", "error", err.Error())
				respChan <- &types.ChatResponse{
					Type:       "error",
					Content:    err.Error(),
					IsComplete: true,
					ID:         assistantMsgID,
				}
				return
			case <-ctx.Done():
				slog.Info("上下文取消，继续处理剩余流式数据", "session_id", request.SessionID)

				go func() {
					for {
						select {
						case resp, ok := <-responseStream:
							if !ok {
								fmt.Printf("[PlaygroundStream] 流式数据处理完成，context已取消\n")
								return
							}

							// 处理完成块以保存数据
							if resp.IsComplete {
								fmt.Printf("[PlaygroundStream] [Context取消] 收到完成块，保存数据\n")

								// 收集思考内容
								if resp.Thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
									thoughts = thoughts + resp.Thoughts
								}

								// 使用累积内容
								if len(fullContent) == 0 && len(resp.Content) > 0 {
									fullContent = resp.Content
								}

								resp.TotalDuration = resp.TotalDuration / int64(time.Second)
								totalDuration = resp.TotalDuration

								// 保存到数据库
								finalContent := fullContent
								if thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
									finalContent = fmt.Sprintf("<think>%s\n</think>\n%s", thoughts, fullContent)
								}

								if finalContent != "" {
									assistantMsg := &types.ChatMessage{
										ID:            assistantMsgID,
										SessionID:     request.SessionID,
										Role:          "assistant",
										Content:       finalContent,
										Order:         len(messages) + 1,
										CreatedAt:     time.Now(),
										ModelID:       session.ModelID,
										ModelName:     session.ModelName,
										TotalDuration: totalDuration,
									}
									saveCtx := context.Background()
									saveCtxWithTimeout, cancel := context.WithTimeout(saveCtx, 10*time.Second)
									defer cancel()

									err = p.Ds.Add(saveCtxWithTimeout, assistantMsg)
									if err != nil {
										fmt.Printf("[PlaygroundStream] [Context取消] 保存失败: %v\n", err)
									} else {
										fmt.Printf("[PlaygroundStream] [Context取消] 成功保存到数据库\n")
									}
								}
								return
							} else if len(resp.Content) > 0 {
								// 继续累积内容
								fullContent += resp.Content
							}
						}
					}
				}()

				respChan <- &types.ChatResponse{
					Type:       "error",
					Content:    ctx.Err().Error(),
					IsComplete: true,
					ID:         assistantMsgID,
				}
				return
			}
		}
	}()

	return respChan, errChan
}

// 更新会话的标题
func (p *PlaygroundImpl) UpdateSessionTitle(ctx context.Context, sessionID string) error {
	sessionCheck := &types.ChatSession{ID: sessionID}
	err := p.Ds.Get(ctx, sessionCheck)
	if err != nil || sessionCheck == nil || sessionCheck.ID == "" {
		fmt.Println("Failed to get chat session", "error", err)
		return err
	}

	// 检查会话标题开头是否是指定的格式
	if !strings.HasPrefix(sessionCheck.Title, "新对话 ") {
		return nil
	}

	// 获取会话中的所有消息，构建历史上下文
	messageQuery := &types.ChatMessage{SessionID: sessionID}
	messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "created_at", Order: datastore.SortOrderAscending},
		},
	})
	if err != nil {
		fmt.Println("Failed to list chat messages", "error", err)
		return err
	}

	// 构建历史对话
	history := make([]map[string]string, 0, len(messages)+1)
	for _, m := range messages {
		msg := m.(*types.ChatMessage)
		if msg.Role == "assistant" {
			// 把thinking内容清理掉
			if strings.Contains(msg.Content, "<think>") && strings.Contains(msg.Content, "</think>") {
				re := regexp.MustCompile(`(?s)<think>.*?</think>\s*`)
				msg.Content = re.ReplaceAllString(msg.Content, "")
				msg.Content = strings.TrimSpace(msg.Content)
			}

			if msg.Content == "" {
				continue
			}
		}
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	genTitlePrompt := "请用一句话为本次对话生成一个不超过10字的简洁标题，仅输出标题本身"
	history = append(history, map[string]string{
		"role":    "user",
		"content": genTitlePrompt,
	})

	modelEngine := engine.NewEngine() // 构建聊天请求
	currentModelName := sessionCheck.ModelName

	if currentModelName == "" {

		if len(messages) > 0 {
			if latestMsg := messages[len(messages)-1].(*types.ChatMessage); latestMsg.ModelName != "" {
				currentModelName = latestMsg.ModelName
			}
		}
	}

	chatRequest := &types.ChatRequest{
		Model:    currentModelName,
		Messages: history,
		Stream:   false,
		Think:    false,
	}

	// fmt.Println("generate title request:", chatRequest)

	res, err := modelEngine.Chat(ctx, chatRequest)
	if err != nil {
		fmt.Println("Failed to generate title", "error", err)
		slog.Error("Failed to generate title", "error", err)
		return err
	}

	// fmt.Printf("generate title res: %v\n", res)

	title := res.Content
	if strings.Contains(title, "<think>") && strings.Contains(title, "</think>") {
		re := regexp.MustCompile(`(?s)<think>.*?</think>\s*`)
		title = re.ReplaceAllString(title, "")
		title = strings.TrimSpace(title)
	}
	if title != "" {
		runes := []rune(title)
		title = string(runes)
		sessionCheck.Title = title
		err = p.Ds.Put(ctx, sessionCheck)
		if err != nil {
			fmt.Println("Failed to generate title put", "error", err)
			slog.Error("Failed to generate title put", "error", err)
			return err
		}
	}

	return nil
}

// 增加会话的标题
func (p *PlaygroundImpl) AddSessionTitle(ctx context.Context, request *dto.SendStreamMessageRequest) error {
	sessionCheck := &types.ChatSession{ID: request.SessionID}
	err := p.Ds.Get(ctx, sessionCheck)
	if err == nil && sessionCheck.Title == "" && request.Content != "" {
		content := request.Content
		runes := []rune(content)
		if len(runes) > 10 {
			content = string(runes[:10])
		}
		title := "新对话 " + content
		sessionCheck.Title = title
		err = p.Ds.Put(ctx, sessionCheck)
	} else {
		slog.Info("AddSessionTitle err", err)
		fmt.Println("AddSessionTitle err", err)
	}

	return err
}

// 处理工具调用，作为历史消息请求大模型
func (p *PlaygroundImpl) HandleToolCalls(ctx context.Context, sessionId string, messageId string) []map[string]string {
	messageQuery := &types.ToolMessage{SessionID: sessionId, MessageId: messageId}
	messages, err := p.Ds.List(ctx, messageQuery, &datastore.ListOptions{
		SortBy: []datastore.SortOption{
			{Key: "updated_at", Order: 1},
		},
	})
	if err != nil {
		slog.Error("Failed to list chat messages", "error", err)
		return nil
	}
	con := new(types.ChatMessage)
	con.ID = messageId
	_ = p.Ds.Get(ctx, con)

	history := make([]map[string]string, 0)
	for _, m := range messages {
		msg := m.(*types.ToolMessage)
		if msg.MessageId == messageId && msg.SessionID == sessionId {
			// 反转义 InputParams 和 OutputParams 字符串
			inputParams := msg.InputParams
			outputParams := msg.OutputParams

			if inputParams != "" && outputParams != "" {
				// 尝试将转义的 JSON 字符串还原为原始 JSON
				var inputObj interface{}
				if err := json.Unmarshal([]byte(inputParams), &inputObj); err == nil {
					// 重新格式化为缩进后的 JSON 字符串
					if pretty, err := json.MarshalIndent(inputObj, "", "  "); err == nil {
						inputParams = string(pretty)
					}
				}

				// 解析 outputParams，获取嵌套的 text 字段
				type ContentItem struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}
				type OutputContent struct {
					Content []ContentItem `json:"content"`
				}
				var outputObj OutputContent
				if err := json.Unmarshal([]byte(outputParams), &outputObj); err == nil {
					// 提取所有 text 字段拼接
					var texts []string
					for _, item := range outputObj.Content {
						texts = append(texts, item.Text)
					}
					outputParams = strings.Join(texts, "\n")
				} else {
					fmt.Println(err)
				}

				history = append(history, map[string]string{
					"role":    "assistant",
					"content": fmt.Sprintf("<tool_use>\n  <name>%s</name>\n  <arguments>%s</arguments>\n</tool_use>\n", msg.Name, inputParams),
				})
				history = append(history, map[string]string{
					"role":    "user",
					"content": outputParams,
				})
			}
		}
	}
	return history
}

// 更新工具调用的逻辑
func (p *PlaygroundImpl) UpdateToolCall(ctx context.Context, toolMessage *types.ToolMessage) error {
	con := new(types.ToolMessage)
	con.ID = toolMessage.ID
	err := p.Ds.Get(ctx, con)

	if err != nil {
		return err
	}
	if con == nil || con.ID == "" {
		return err
	}

	// 保存工具调用状态
	con.McpId = toolMessage.McpId
	con.Logo = toolMessage.Logo
	con.Name = toolMessage.Name
	con.Desc = toolMessage.Desc
	con.InputParams = toolMessage.InputParams
	con.OutputParams = toolMessage.OutputParams
	con.Status = toolMessage.Status
	con.UpdatedAt = time.Now()

	err = p.Ds.Put(ctx, con)
	if err != nil {
		return err
	}

	return nil
}

// 新增工具调用的逻辑
func (p *PlaygroundImpl) AddToolCall(ctx context.Context, sessionId, messageId string, id string, totalDuration int64) (string, error) {
	// 1 判断McpIds以及MessageID 先存储一次工具调用逻辑
	toolMessage := &types.ToolMessage{
		ID:            id,
		SessionID:     sessionId,
		MessageId:     messageId,
		ExecutionTime: totalDuration,
	}

	err := p.Ds.Add(ctx, toolMessage)
	if err != nil {
		fmt.Println("Failed to save tool message", "error", err, "session_id", sessionId, "message_id", messageId, "tool_id", id)
		return "", fmt.Errorf("failed to save tool message: %w", err)
	}

	return toolMessage.ID, nil
}

// 标识一轮工具调用结束，批量更新所有匹配 SessionID 和 MessageId 的 ToolMessage 的 AssistantMsgID
func (p *PlaygroundImpl) MarkToolCallEnd(ctx context.Context, sessionId, messageId, assistantMsgID string) error {
	// 查询所有匹配的 ToolMessage
	query := &types.ToolMessage{
		SessionID: sessionId,
		MessageId: messageId,
	}
	messages, err := p.Ds.List(ctx, query, nil)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return nil // 没有需要更新的记录
	}

	for _, m := range messages {
		toolMsg := m.(*types.ToolMessage)
		if toolMsg.MessageId == messageId {
			toolMsg.AssistantMsgID = assistantMsgID
			if err := p.Ds.Put(ctx, toolMsg); err != nil {
				return err
			}
		}
	}

	con := new(types.ChatMessage)
	con.ID = assistantMsgID
	err = p.Ds.Get(ctx, con)

	if err != nil {
		return err
	}
	if con == nil || con.ID == "" {
		return fmt.Errorf("chat message not found")
	}
	con.IsToolGroupID = true // 标记为工具组ID
	con.UpdatedAt = time.Now()
	err = p.Ds.Put(ctx, con)

	return nil
}
