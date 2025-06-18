package engine

import (
	"byze/internal/schedule"
	"byze/internal/types"
	"context"
	"encoding/json"
	"fmt"
)

func (e *Engine) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan *types.ChatResponse, <-chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)
	body, err := json.Marshal(req)
	if err != nil {
		go func() {
			errChan <- err
			close(respChan)
			close(errChan)
		}()
		return respChan, errChan
	}

	// Convert model ID to model name if needed
	originalModel := req.Model
	modelName := getModelNameById(req.Model)

	// Debug log to trace model conversion
	fmt.Printf("[ChatStream] Model conversion: %s -> %s\n", originalModel, modelName)

	// 打印即将发往Ollama的请求体内容，重点关注think参数
	fmt.Printf("[ChatStream] Final request body to Ollama: %s\n", string(body))

	serviceReq := &types.ServiceRequest{
		Service:       "chat",
		Model:         modelName, // 使用模型名
		FromFlavor:    "ollama",  // 使用Ollama风格
		AskStreamMode: true,      // 启用流式输出
		Think:         req.Think,
		HTTP: types.HTTPContent{
			Body: body,
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
		var thoughts string

		// 处理流式响应
		for result := range ch {
			if result.Error != nil {
				errChan <- result.Error
				return
			}

			// 如果chunk为空，则跳过
			if len(result.HTTP.Body) == 0 {
				// fmt.Printf("[ChatStream] 收到空块，跳过\n")
				continue
			}
			// Debug输出
			// fmt.Printf("[ChatStream] 收到块，长度: %d\n", len(result.HTTP.Body))
			// debugLogJSON("[ChatStream] 原始响应内容", result.HTTP.Body)
			// debugLogJSON("[ChatStream] 收到块内容", result.HTTP.Body) // 调用调试日志函数

			// 每个块都是一个完整的JSON对象
			var ollamaResp ollamaAPIResponse
			var content string
			isComplete := false
			model := ""

			// 首先尝试解析为Ollama API标准响应
			if err := json.Unmarshal(result.HTTP.Body, &ollamaResp); err == nil {
				// 提取模型名称
				model = ollamaResp.Model

				// 检查是否完成
				isComplete = ollamaResp.Done

				// 按优先级提取content
				// 1. 如果存在message字段且有内容，使用message.content (优先 /api/chat 格式)
				if ollamaResp.Message != nil && ollamaResp.Message.Content != "" {
					content = ollamaResp.Message.Content
					// fmt.Printf("[ChatStream] 从message.content提取内容，长度: %d\n", len(content))
				} else if ollamaResp.Response != "" {
					// 2. 如果存在response字段且有内容，使用response (/api/generate 格式)
					content = ollamaResp.Response
					fmt.Printf("[ChatStream] 从response提取内容，长度: %d\n", len(content))
				} else if ollamaResp.Content != "" {
					// 3. 如果存在content字段且有内容，使用content (备用)
					content = ollamaResp.Content
					fmt.Printf("[ChatStream] 从content提取内容，长度: %d\n", len(content))
				}

				var thoughts string
				if ollamaResp.Message != nil && ollamaResp.Message.Thinking != "" {
					thoughts = ollamaResp.Message.Thinking
					fmt.Printf("[ChatStream] 提取到思考内容，长度: %d\n", len(thoughts))
				}

				// 提取工具调用(如果有)
				if ollamaResp.Message != nil && ollamaResp.Message.ToolCalls != nil && len(ollamaResp.Message.ToolCalls) > 0 {
					toolCalls = ollamaResp.Message.ToolCalls
					fmt.Printf("[ChatStream] 提取到工具调用，数量: %d\n", len(toolCalls))
				} else if ollamaResp.ToolCalls != nil && len(ollamaResp.ToolCalls) > 0 {
					toolCalls = ollamaResp.ToolCalls
					fmt.Printf("[ChatStream] 提取到工具调用，数量: %d\n", len(toolCalls))
				}
			} else {
				// 如果标准解析失败，尝试使用通用map解析
				fmt.Printf("[ChatStream] 标准解析失败: %v，尝试通用解析\n", err)

				var data map[string]interface{}
				if err := json.Unmarshal(result.HTTP.Body, &data); err == nil {
					// 提取消息内容
					extractedContent, found := extractContent(data)
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
			} // 处理提取到的内容
			if content != "" {
				accumulatedContent += content
				// fmt.Printf("[ChatStream] 累积内容，当前长度: %d\n", len(accumulatedContent))
			} // 创建响应对象
			resp := &types.ChatResponse{
				Content:       content, // 只发送当前块的内容，而不是累积的内容
				Model:         model,
				IsComplete:    isComplete,
				ToolCalls:     toolCalls,
				Object:        "chat.completion.chunk",
				TotalDuration: ollamaResp.TotalDuration, // 使用HTTP响应的持续时间
				Thoughts:      thoughts,                 // 添加思考内容
			} // 发送响应
			// 只发送有内容或是最后一个块的响应
			if resp.Content != "" || resp.IsComplete {
				// 如果是最后一个块，发送完整累积的内容
				if resp.IsComplete {
					// 确保最后一块还包含之前累积的内容
					resp.Content = accumulatedContent
					resp.Object = "chat.completion"
				}
				respChan <- resp
			}
		}
	}()
	return respChan, errChan
}

// 辅助函数：从数据中提取内容
func extractContent(data map[string]interface{}) (string, bool) {

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

	if text, ok := data["text"].(string); ok && text != "" {
		return text, true
	}

	if msg, ok := data["message"].(map[string]interface{}); ok {
		if toolCalls, ok := msg["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
			return "", true
		}
	}

	if toolCalls, ok := data["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
		return "", true
	}

	return "", false
}

func debugLogJSON(prefix string, data []byte) {
	maxLen := 1000
	content := string(data)
	if len(content) > maxLen {
		content = content[:maxLen] + "... (truncated)"
	}
	fmt.Printf("%s: %s\n", prefix, content)
}
