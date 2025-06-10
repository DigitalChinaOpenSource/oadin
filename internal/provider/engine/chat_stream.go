package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"byze/internal/types"
	"byze/internal/utils"

	"github.com/google/uuid"
)


func (o *OllamaProvider) ChatStream(ctx context.Context, req *types.ChatRequest) (chan *types.ChatResponse, chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		c := o.GetDefaultClient()
		// 确保设置流式请求标志
		reqCopy := *req
		reqCopy.Stream = true

		// 确保选项存在
		if reqCopy.Options == nil {
			reqCopy.Options = make(map[string]any)
		}
		// 准备思考模式
		var thoughts strings.Builder

		// 检查是否启用了thinking选项
		thinkingEnabled := false
		if reqCopy.Options != nil {
			if thinking, ok := reqCopy.Options["thinking"].(bool); ok && thinking {
				thinkingEnabled = true
			}
		}

		if thinkingEnabled {
			// 注意：此处思考模式基于Ollama特性，如果使用不支持thinking的模型会触发错误
			slog.Info("启用思考模式", "model", req.Model)

			// 创建一个单独的思考请求
			thinkingReq := types.ChatRequest{
				Model:    req.Model,
				Messages: req.Messages,
				Options: map[string]any{"thinking": true},
			}

			// 执行思考请求，使用上下文超时控制
			thinkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			thinkResp, err := o.Chat(thinkCtx, &thinkingReq)
			if err != nil {
				slog.Warn("思考请求失败，将继续执行主请求", "error", err)
				// 思考失败不阻止主请求
			} else if thinkResp != nil {
				// 获取思考内容
				thoughts.WriteString(thinkResp.Content)
				slog.Debug("获取思考内容成功", "length", len(thinkResp.Content))
			}
		}

		// 处理流式输出
		url := "/api/chat"
		if strings.HasPrefix(req.Model, "openai/") {
			url = "/v1/chat/completions"
		}

		// 使用StreamResponse方法获取流式响应
		dataChan, streamErrChan := c.StreamResponse(ctx, http.MethodPost, url, reqCopy)

		responseID := uuid.New().String()
		var fullContent strings.Builder

		// 处理流式数据
		for {
			select {
			case data, ok := <-dataChan:
				if !ok {
					// 流结束
					if fullContent.Len() > 0 {
						respChan <- &types.ChatResponse{
							ID:         responseID,
							Object:     "chat.completion.chunk",
							Model:      req.Model,
							Content:    fullContent.String(),
							IsComplete: true,
							Thoughts:   thoughts.String(),
						}					}
					return
				} 
				
				// 解析JSON响应
				var chunk map[string]interface{}
				if err := json.Unmarshal(data, &chunk); err != nil {
					slog.Debug("流响应解析失败", "error", err, "data", string(data))
					errChan <- fmt.Errorf("failed to decode response: %v", err)
					continue
				}

				// 提取内容片段
				var content string

				// 检查错误响应
				if errMsg, hasErr := chunk["error"].(string); hasErr && errMsg != "" {
					slog.Error("模型返回错误", "error", errMsg)
					errChan <- fmt.Errorf("model error: %s", errMsg)
					continue
				}

				// 尝试Ollama格式
				message, ok := chunk["message"].(map[string]interface{})
				if ok {
					content, _ = message["content"].(string)
				} else {
					// 尝试OpenAI格式
					choices, ok := chunk["choices"].([]interface{})
					if ok && len(choices) > 0 {
						choice := choices[0].(map[string]interface{})
						delta, ok := choice["delta"].(map[string]interface{})
						if ok {
							content, _ = delta["content"].(string)
						}
					} else {
						// 尝试直接从内容字段获取
						if directContent, ok := chunk["content"].(string); ok {
							content = directContent
						}
					}
				}

				if content != "" {
					fullContent.WriteString(content)

					respChan <- &types.ChatResponse{
						ID:         responseID,
						Object:     "chat.completion.chunk",
						Model:      req.Model,
						Content:    content,
						IsComplete: false,
						Thoughts:   thoughts.String(),
					}
				}

			case err, ok := <-streamErrChan:
				if !ok {
					return
				}
				errChan <- err
				return

			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}
		}
	}()

	return respChan, errChan
}


func (o *OpenvinoProvider) ChatStream(ctx context.Context, req *types.ChatRequest) (chan *types.ChatResponse, chan error) {
	respChan := make(chan *types.ChatResponse)
	errChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errChan)

		// 非流式实现，使用模拟流式输出
		resp, err := o.Chat(ctx, req)
		if err != nil {
			errChan <- err
			return
		}

		// 如果内容很短，直接返回
		if len(resp.Content) < 100 {
			resp.IsComplete = true
			respChan <- resp
			return
		}
		// 模拟流式输出
		chunks := utils.SplitIntoChunks(resp.Content, 20)
		for i, chunk := range chunks {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
				respChan <- &types.ChatResponse{
					ID:         resp.ID,
					Object:     resp.Object,
					Model:      resp.Model,
					Content:    chunk,
					IsComplete: i == len(chunks)-1,
					Thoughts:   resp.Thoughts,
				}
				// 模拟真实流式输出
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
	return respChan, errChan
}
