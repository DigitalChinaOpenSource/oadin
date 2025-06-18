package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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
				errChan <- err
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
		// 转发流式响应
		for {
			select {
			case resp, ok := <-responseStream:
				if !ok { // 流结束
					// 因为有可能最后一个块是完成标记但内容为空
					slog.Info("流式输出结束，准备保存助手回复",
						"content_length", len(fullContent))

					// 显示预览（如果有内容）
					if len(fullContent) > 0 {
						previewLen := min(100, len(fullContent))
						slog.Info("回复内容预览",
							"content_preview", fullContent[:previewLen])
					} else {
						slog.Warn("助手回复内容为空！")
					}

					// 如果是第一条消息，更新会话标题
					if len(messages) == 0 {
						genTitlePrompt := fmt.Sprintf("请为以下用户问题生成一个简洁的标题（不超过10字），用于在首轮对话时生成对话标题，该标题应描述用户提问的主题或意图，而不是问题的答案本身：%s", request.Content)
						// 构造 HTTP 请求体
						payload := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"%s"}],"stream":false}`,
							session.ModelName, genTitlePrompt)
						slog.Info("[DEBUG] TitleGen HTTP payload", "payload", payload)
						client := &http.Client{Timeout: 15 * time.Second}
						req, err := http.NewRequest("POST", "http://localhost:16688/byze/v0.2/services/chat", strings.NewReader(payload))
						if err == nil {
							req.Header.Set("Content-Type", "application/json")
							resp, err := client.Do(req)
							title := "新对话 " + time.Now().Format("2006-01-02")
							if err == nil && resp != nil {
								defer resp.Body.Close()
								var result struct {
									Content string `json:"content"`
									Message struct {
										Content string `json:"content"`
									} `json:"message"`
								}
								decodeErr := json.NewDecoder(resp.Body).Decode(&result)
								slog.Info("[DEBUG] TitleGen HTTP resp", "decodeErr", decodeErr, "respContent", result.Content, "msgContent", result.Message.Content)
								if decodeErr == nil {
									if len(result.Content) > 0 {
										title = result.Content
									} else if len(result.Message.Content) > 0 {
										title = result.Message.Content
									}
									runes := []rune(title)
									if len(runes) > 10 {
										title = string(runes[:10])
									} else {
										title = string(runes)
									}
								}
							}
							slog.Info("[DEBUG] TitleGen final title", "title", title)
							session.Title = title
							err = p.Ds.Put(context.Background(), session)
							if err != nil {
								slog.Error("Failed to update session title", "error", err)
							}
						}
					} // 保存思考内容（如果有）
					if thoughts != "" && session.ThinkingEnabled && session.ThinkingActive {
						thoughtsMsg := &types.ChatMessage{
							ID:            uuid.New().String(),
							SessionID:     request.SessionID,
							Role:          "think",
							Content:       thoughts,
							Order:         len(messages) + 2,
							CreatedAt:     time.Now(),
							ModelID:       session.ModelID,
							ModelName:     session.ModelName,
							TotalDuration: resp.TotalDuration / int64(time.Second),
						}
						err = p.Ds.Add(ctx, thoughtsMsg)
						if err != nil {
							slog.Error("Failed to save thoughts message", "error", err)
							// 非致命错误，继续执行
						}
					}
					return
				}
				msgType := "answer"
				if resp.Thoughts != "" {
					msgType = "thoughts"
				} // 保留原始的 Content
				originalContent := resp.Content
				resp.Type = msgType
				resp.Model = session.ModelName
				resp.ModelName = session.ModelName

				if resp.IsComplete {
					slog.Info("收到流式输出完成标记",
						"is_complete", resp.IsComplete,
						"content_length", len(originalContent),
						"accumulated_content_length", len(fullContent))

					// 如果最后一块有内容，需要累积
					if len(originalContent) > 0 {
						fullContent += resp.Content
					}

					// 确保完整内容被保存和返回给客户端
					if fullContent != "" {
						assistantMsg := &types.ChatMessage{
							ID:            assistantMsgID,
							SessionID:     request.SessionID,
							Role:          "assistant",
							Content:       fullContent,
							Order:         len(messages) + 1,
							CreatedAt:     time.Now(),
							ModelID:       session.ModelID,
							ModelName:     session.ModelName,
							TotalDuration: resp.TotalDuration / int64(time.Second),
						}
						err = p.Ds.Add(ctx, assistantMsg)
						if err != nil {
							slog.Error("Failed to save assistant message on complete", "error", err)
						}
					}

					// 处理多轮工具调用
					if fullContent == "" && len(resp.ToolCalls) > 0 {
						if request.ToolGroupID == "" {
							if userMsg != nil {
								p.AddToolCall(ctx, request.SessionID, userMsg.ID, assistantMsgID)
							}
						} else {
							p.AddToolCall(ctx, request.SessionID, request.ToolGroupID, assistantMsgID)
						}
					}

					// 发送全部内容作为响应
					resp.Content = fullContent
					resp.ID = assistantMsgID
					if userMsg != nil && len(resp.ToolCalls) > 0 && request.ToolGroupID == "" {
						resp.ToolGroupID = userMsg.ID
					}
					if userMsg == nil && len(resp.ToolCalls) > 0 && request.ToolGroupID != "" {
						resp.ToolGroupID = request.ToolGroupID
					}
					respChan <- resp
				} else if len(originalContent) > 0 {
					slog.Info("收到非空流式输出块",
						"content_length", len(originalContent),
						"is_complete", resp.IsComplete)

					fullContent += resp.Content

					respChan <- resp
				} else {
					slog.Debug("跳过空内容块")
					continue
				}
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
		// 反转义 InputParams 和 OutputParams 字符串
		inputParams := msg.InputParams
		outputParams := msg.OutputParams

		// 尝试将转义的 JSON 字符串还原为原始 JSON
		var inputObj interface{}
		var outputObj interface{}
		if err := json.Unmarshal([]byte(inputParams), &inputObj); err == nil {
			// 重新格式化为缩进后的 JSON 字符串
			if pretty, err := json.MarshalIndent(inputObj, "", "  "); err == nil {
				inputParams = string(pretty)
			}
		}
		if err := json.Unmarshal([]byte(outputParams), &outputObj); err == nil {
			if pretty, err := json.MarshalIndent(outputObj, "", "  "); err == nil {
				outputParams = string(pretty)
			}
		}

		history = append(history, map[string]string{
			"role":    "assistant",
			"content": fmt.Sprintf("调用工具: %s, 参数: %s", msg.Name, inputParams),
		})
		history = append(history, map[string]string{
			"role":    "user",
			"content": fmt.Sprintf("工具调用结果: %s", outputParams),
		})
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
	con.ExecutionTime = toolMessage.ExecutionTime

	err = p.Ds.Put(ctx, con)
	if err != nil {
		return err
	}

	return nil
}

// 新增工具调用的逻辑
func (p *PlaygroundImpl) AddToolCall(ctx context.Context, sessionId, messageId string, id string) (string, error) {
	// 1 判断McpIds以及MessageID 先存储一次工具调用逻辑
	toolMessage := &types.ToolMessage{
		ID:        id,
		SessionID: sessionId,
		MessageId: messageId,
	}

	err := p.Ds.Add(ctx, toolMessage)
	if err != nil {
		fmt.Println("Failed to save tool message", "error", err, "session_id", sessionId, "message_id", messageId, "tool_id", id)
		return "", fmt.Errorf("failed to save tool message: %w", err)
	}

	con := new(types.ChatMessage)
	con.ID = messageId
	err = p.Ds.Get(ctx, con)

	if err != nil {
		return "", err
	}
	if con == nil || con.ID == "" {
		return "", fmt.Errorf("chat message not found")
	}
	con.IsToolGroupID = true // 标记为工具组ID
	con.UpdatedAt = time.Now()
	err = p.Ds.Put(ctx, con)

	return toolMessage.ID, nil

	// 1.1 当客户端发消息时，请求参数只有McpIds，没有messageId时，证明是第一次工具调用，此时生成mcpMessageId为后续工具调用的问题消息标识
	// 1.2 当客户发消息时，请求参数只有mcpMessageId,就证明工具调用已经结束了
	/*
			   {
				"SessionID": "917b68be-93b5-4893-a134-56bfc0fd2a00",
				"content": "武汉中建星光城到大悦城的路线规划",
				"mcpMessageId": ["683ec88241fa614eb1531fc4"]
		       }

			   {
			    "SessionID": "917b68be-93b5-4893-a134-56bfc0fd2a00",
				"content": "",
				"mcpIds": ["683ec88241fa614eb1531fc4"]
				"mcpMessageId": "0"
			   }

			   {
			    "SessionID": "917b68be-93b5-4893-a134-56bfc0fd2a00",
				"content": "",
				"mcpMessageId": "0"
			   }

	*/

	// 调用工具时补充
	// 2.1 如果是第一次调用，生成一个新的工具调用ID
	/*
		{
				ID            string `json:"id"`
				SessionID     string `json:"session_id"`
				MessageId     string `json:"message_id"`     // 关联的消息ID
				McpId         string `json:"mcp_id"`         // 工具调用的MCP ID
				Logo          string `json:"logo"`      // 工具调用的MCP图标
				Name          string `json:"name"`           // 工具名称
				Desc          string `json:"desc"`           // 工具描述
				InputParams   string `json:"input_params"`   // 输入参数
				OutputParams  string `json:"output_params"`  // 输出参数
				Status        string `json:"status"`         // 工具调用状态

		}
	*/

	// 请求大模型问答的时候，将工具调用的信息作为历史消息传入

	// return nil
}
