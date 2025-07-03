package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"oadin/internal/datastore"
	"oadin/internal/schedule"
	"oadin/internal/types"
	"strings"
)

func (e *Engine) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, <-chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)

	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		go func() {
			errChan <- err
			close(respChan)
			close(errChan)
		}()
		return respChan, errChan
	}

	modelName := req.Model

	// 打印即将发往服务的请求体内容，重点关注think参数
	fmt.Printf("[ChatStream] Final request body: %s\n", string(body))
	slog.Info("[ChatStream] Final request body: ", string(body))

	hybridPolicy := "default"
	ds := datastore.GetDefaultDatastore()
	sp := &types.Service{
		Name:   "chat",
		Status: 1,
	}
	err = ds.Get(context.Background(), sp)
	if err != nil {
		slog.Error("[Schedule] Failed to get service", "error", err, "service", "embed")
	} else {
		hybridPolicy = sp.HybridPolicy
	}
	hybridPolicy = sp.HybridPolicy

	serviceReq := &types.ServiceRequest{
		Service:       "chat",
		Model:         modelName,
		FromFlavor:    "oadin",
		HybridPolicy:  hybridPolicy,
		AskStreamMode: true,
		Think:         req.Think,
		HTTP: types.HTTPContent{
			Header: http.Header{},
			Body:   body,
		},
	}

	// 请求调度器执行任务
	_, ch := schedule.GetScheduler().Enqueue(serviceReq)

	go func() {
		defer close(respChan)
		defer close(errChan)

		// 跟踪流的状态
		accumulatedContent := ""
		var toolCalls []types.ToolCall
		var totalDuration int64 // 跟踪总处理时间

		// 处理流式响应
		for result := range ch {
			if result.Error != nil {
				errChan <- result.Error
				return
			}

			// 如果chunk为空，则跳过
			if len(result.HTTP.Body) == 0 {
				continue
			}

			// 处理前缀
			bodyStr := string(result.HTTP.Body)
			if strings.HasPrefix(bodyStr, "data: ") {
				bodyStr = strings.TrimPrefix(bodyStr, "data: ")
				bodyStr = strings.TrimSpace(bodyStr)
			}
			// 转回[]byte
			cleanBody := []byte(bodyStr)
			fmt.Println("[ChatStream] 收到块内容:", bodyStr)

			// 每个块都是一个完整的JSON对象
			var content string
			isComplete := false
			var thoughts string
			model := ""
			parseSucceeded := false

			type oadinStreamChunk struct {
				CreatedAt    string `json:"created_at"`
				FinishReason string `json:"finish_reason,omitempty"`
				Finished     bool   `json:"finished"`
				Id           string `json:"id"`
				Message      struct {
					Content   string           `json:"content"`
					Role      string           `json:"role"`
					Thinking  string           `json:"thinking"`
					ToolCalls []types.ToolCall `json:"tool_calls,omitempty"`
				} `json:"message"`
				Model string `json:"model"`
			}

			var streamChunk oadinStreamChunk
			if err := json.Unmarshal(cleanBody, &streamChunk); err == nil {
				// 成功解析为直接流式格式
				parseSucceeded = true
				// fmt.Printf("[ChatStream] 解析为直接流式格式成功\n")

				// 提取模型名称
				model = streamChunk.Model

				// 提取内容
				if streamChunk.Message.Content != "" {
					content = streamChunk.Message.Content
					// fmt.Printf("[ChatStream] 从直接流式格式提取内容，长度: %d\n", len(content))
				}

				if streamChunk.Message.Thinking != "" {
					thoughts = streamChunk.Message.Thinking
					// fmt.Printf("[ChatStream] 从直接流式格式提取内容，长度: %d\n", len(content))
				}

				if len(streamChunk.Message.ToolCalls) > 0 {
					toolCalls = append(toolCalls, streamChunk.Message.ToolCalls...)
				}

				// 检查是否完成
				isComplete = streamChunk.Finished || streamChunk.FinishReason != ""
				if isComplete {
					fmt.Printf("[ChatStream] 检测到流式输出完成标记 finished=%v, finish_reason=%s\n",
						streamChunk.Finished, streamChunk.FinishReason)
				}
			}

			if !parseSucceeded {
				var oadinResp types.OadinAPIResponse
				if err := json.Unmarshal(cleanBody, &oadinResp); err == nil && oadinResp.BusinessCode == 10000 {
					dataBytes, err := json.Marshal(oadinResp.Data)
					if err == nil {
						var streamResp types.OadinChatStreamResponse
						if err := json.Unmarshal(dataBytes, &streamResp); err == nil {
							parseSucceeded = true
							fmt.Printf("[ChatStream] 解析为标准Oadin响应格式成功\n")

							model = streamResp.Model

							if len(streamResp.Choices) > 0 {
								if streamResp.Choices[0].Delta.Content != "" {
									content = streamResp.Choices[0].Delta.Content
									fmt.Printf("[ChatStream] 从Oadin delta提取内容，长度: %d\n", len(content))
								}

								if streamResp.Choices[0].Delta.Thinking != "" {
									thoughts = streamResp.Choices[0].Delta.Thinking
									fmt.Printf("[ChatStream] 从Oadin delta提取思考内容，长度: %d\n", len(thoughts))
								}

								isComplete = streamResp.Choices[0].FinishReason != ""

								if len(streamResp.Choices[0].Delta.ToolCalls) > 0 {
									toolCalls = streamResp.Choices[0].Delta.ToolCalls
									fmt.Printf("[ChatStream] 从Oadin delta提取到工具调用，数量: %d\n", len(toolCalls))
									slog.Info("[ChatStream] 从Oadin delta提取到工具调用", "数量", len(toolCalls))
								}
							}
						}
					}
				}
			}

			if !parseSucceeded {
				fmt.Printf("[ChatStream] Oadin API响应解析失败，尝试通用格式解析\n")

				var data map[string]interface{}
				if err := json.Unmarshal(cleanBody, &data); err == nil {
					extractContentLocal := func(data map[string]interface{}) (string, bool) {
						if msg, ok := data["message"].(map[string]interface{}); ok {
							if content, ok := msg["content"].(string); ok && content != "" {
								return content, true
							}
						}

						if response, ok := data["response"].(string); ok && response != "" {
							return response, true
						}

						if content, ok := data["content"].(string); ok && content != "" {
							return content, true
						}

						return "", false
					}

					// 提取消息内容
					extractedContent, found := extractContentLocal(data)
					if found {
						content = extractedContent
						fmt.Printf("[ChatStream] 提取到内容，长度: %d\n", len(content))
					}

					// 检查是否完成
					if done, ok := data["done"].(bool); ok {
						isComplete = done
					}

					// 提取模型名称
					if m, ok := data["model"].(string); ok {
						model = m
					}

					// 提取思考内容
					if msg, ok := data["message"].(map[string]interface{}); ok {
						if th, ok := msg["thinking"].(string); ok && th != "" {
							thoughts = th
							fmt.Printf("[ChatStream] 从通用格式message.thinking中提取到思考内容，长度: %d\n", len(thoughts))
						}
					}
					// 如果没有在message中找到thinking，尝试从顶层查找
					if thoughts == "" {
						if th, ok := data["thinking"].(string); ok && th != "" {
							thoughts = th
							fmt.Printf("[ChatStream] 从顶层thinking中提取到思考内容，长度: %d\n", len(thoughts))
						}
					}

					// 提取工具调用
					if msg, ok := data["message"].(map[string]interface{}); ok {
						if tc, ok := msg["tool_calls"].([]types.ToolCall); ok && len(tc) > 0 {
							toolCalls = tc
							fmt.Printf("[ChatStream] 提取到工具调用，数量: %d\n", len(toolCalls))
							slog.Info("[ChatStream] 从message中提取到工具调用", "数量", len(toolCalls))
						}
					}

					// 如果没有在message中找到，尝试从顶层查找
					if len(toolCalls) == 0 {
						if tc, ok := data["tool_calls"].([]types.ToolCall); ok && len(tc) > 0 {
							toolCalls = tc
							fmt.Printf("[ChatStream] 从顶层提取到工具调用，数量: %d\n", len(toolCalls))
						}
					}
				} else {
					fmt.Printf("[ChatStream] JSON解析完全失败: %v，跳过此块\n", err)
					continue // 如果连JSON都解析不了，就跳过这个块
				}
			}

			// 处理提取到的内容
			if content != "" {
				accumulatedContent += content
			}

			// 创建响应对象
			resp := &types.ChatResponse{
				Content:       content, // 只发送当前块的内容，而不是累积的内容
				Model:         model,
				IsComplete:    isComplete,
				ToolCalls:     toolCalls,
				Object:        "chat.completion.chunk",
				TotalDuration: totalDuration,
				Thoughts:      thoughts,
			}

			if isComplete {
				fmt.Printf("[ChatStream] 收到完成标记，当前块内容长度: %d，累积内容长度: %d\n",
					len(content), len(accumulatedContent))
			}

			// 发送响应
			// 只发送有内容或是最后一个块的响应
			if (resp.Content != "" || resp.Thoughts != "") || resp.IsComplete {
				// 如果是最后一个块，发送完整累积的内容
				if resp.IsComplete {
					var tempData map[string]interface{}
					if err := json.Unmarshal(cleanBody, &tempData); err == nil {
						if duration, ok := tempData["total_duration"].(float64); ok {
							totalDuration = int64(duration)
							resp.TotalDuration = totalDuration
							fmt.Printf("[ChatStream] 提取到总时长: %dms\n", totalDuration)
						}
					}

					// 确保最后一块还包含之前累积的内容
					resp.Content = accumulatedContent
					resp.Object = "chat.completion"
					fmt.Printf("[ChatStream] 发送最终完整响应，内容长度: %d\n", len(accumulatedContent))

					if len(accumulatedContent) == 0 {
						fmt.Printf("[ChatStream] 警告：累积内容为空，尝试使用最后接收的非空内容\n")
					}
				}
				respChan <- resp
			}
		}
	}()
	return respChan, errChan
}
