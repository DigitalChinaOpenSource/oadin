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

					assistantMsg := &types.ChatMessage{
						ID:        assistantMsgID,
						SessionID: request.SessionID,
						Role:      "assistant",
						Content:   fullContent, // 即使为空也保存
						Order:     len(messages) + 1,
						CreatedAt: time.Now(),
						ModelID:   session.ModelID,
						ModelName: session.ModelName,
					}
					err = p.Ds.Add(ctx, assistantMsg)
					if err != nil {
						slog.Error("Failed to save assistant message", "error", err)
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
									title = string(runes)
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

					// 如果没有内容但有工具调用，构建提示信息
					if fullContent == "" && len(resp.ToolCalls) > 0 {
						for _, toolCall := range resp.ToolCalls {
							// toolCall.Function.Argument 是map[string]interface{}, 转为json字符串
							arguments, err := json.Marshal(toolCall.Function.Arguments)
							if err != nil {
								slog.Error("工具调用参数序列化失败", "error", err, "arguments", toolCall.Function.Arguments)
							}
							fullContent += fmt.Sprintf("<tool_use>\n  <name>%s</name>\n  <arguments>%s</arguments>\n</tool_use>\n", toolCall.Function.Name, arguments)
						}
					}

					// 确保完整内容被保存和返回给客户端
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

					// 发送全部内容作为响应
					resp.Content = fullContent
					resp.ID = assistantMsgID
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
