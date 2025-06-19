import { useState, useCallback, useRef, useEffect } from 'react';
import { fetchEventSource, FetchEventSourceInit } from '@microsoft/fetch-event-source';
import { baseHeaders } from '@/utils';
import { API_PREFIX } from '@/constants';
import { MessageType } from '@res-utiles/ui-components';
import useChatStore from '@/components/chat-container/store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useSelectedModelStore from '@/store/useSelectedModel';
import { getIdByFunction } from '../..//select-mcp/lib/useSelectMcpHelper';
import { httpRequest } from '@/utils/httpRequest';
import { generateUniqueId } from './utils';
import { findExistingToolMessage, findProgressToolMessage, extractToolCallData, createMcpContent, updateContentListWithMcp, buildToolCallData, handleToolCallErrorMessage } from './mcpContentUtils';
import { buildMessageWithThinkContent, handleTextContent } from './thinkContentUtils';
import { IRunToolParams, IToolCall, StreamCallbacks, ChatRequestParams, ChatResponseData } from './types';

// 错误消息常量
const ERROR_MESSAGES = {
  NO_MODEL_SELECTED: '未选择模型，请先选择一个模型',
  NO_DATA_TIMEOUT: '请求超时：60秒未收到数据',
  TOTAL_TIMEOUT: '请求超时：120秒内未收到任何数据',
  PARSE_ERROR: '解析响应数据失败',
  CONNECTION_INTERRUPTED: '\n\n[消息传输中断]',
  EMPTY_RESPONSE: '服务器响应异常，未能获取到有效数据',
  NON_STREAMING: '服务器未返回流式数据',
  TOOL_EXECUTION_FAILED: '工具执行失败',
  RESPONSE_INTERRUPTED: '\n\n[回复已中断]',
};

// 超时设置（毫秒）
const TIMEOUT_CONFIG = {
  NO_DATA: 60000, // 60秒无数据超时
  TOTAL: 120000, // 120秒总超时
};

/**
 * 流式聊天Hook
 */
export function useChatStream(currentSessionId: string) {
  const { addMessage, setCurrentSessionId } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();

  const [streamingContent, setStreamingContent] = useState<string>('');
  const [streamingThinking, setStreamingThinking] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 请求控制
  const abortControllerRef = useRef<AbortController | null>(null);

  const requestState = useRef({
    content: {
      response: '',
      thinking: '',
    },
    status: {
      hasReceivedData: false,
      isToolCallActive: false,
    },
    timers: {
      noDataTimer: null as NodeJS.Timeout | null,
      totalTimer: null as NodeJS.Timeout | null,
    },
  });

  // 上下文保存
  const lastUserMessageRef = useRef<string | null>(null);
  const toolCallHandlersRef = useRef<{
    continueConversation: ((result: string) => Promise<void>) | null;
  }>({ continueConversation: null });

  /**
   * 清除所有定时器
   */
  const clearTimers = useCallback(() => {
    if (requestState.current.timers.noDataTimer) {
      clearTimeout(requestState.current.timers.noDataTimer);
      requestState.current.timers.noDataTimer = null;
    }

    if (requestState.current.timers.totalTimer) {
      clearTimeout(requestState.current.timers.totalTimer);
      requestState.current.timers.totalTimer = null;
    }
  }, []);

  /**
   * 重置无数据超时定时器
   */
  const resetNoDataTimer = useCallback(() => {
    if (requestState.current.timers.noDataTimer) {
      clearTimeout(requestState.current.timers.noDataTimer);
    }

    requestState.current.timers.noDataTimer = setTimeout(() => {
      console.log('无数据超时触发');
      setError(ERROR_MESSAGES.NO_DATA_TIMEOUT);
      cancelRequest('无数据超时');
    }, TIMEOUT_CONFIG.NO_DATA);
  }, []);

  // 清除流式状态
  const clearStreamingState = useCallback(() => {
    setStreamingContent('');
    setStreamingThinking('');
    setError(null);

    // 重置请求状态
    requestState.current = {
      content: {
        response: '',
        thinking: '',
      },
      status: {
        hasReceivedData: false,
        isToolCallActive: false,
      },
      timers: {
        noDataTimer: null,
        totalTimer: null,
      },
    };
  }, []);

  const cleanupResources = useCallback(() => {
    // 取消网络请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    // 清理定时器
    if (requestState.current.timers.noDataTimer) {
      clearTimeout(requestState.current.timers.noDataTimer);
      requestState.current.timers.noDataTimer = null;
    }

    if (requestState.current.timers.totalTimer) {
      clearTimeout(requestState.current.timers.totalTimer);
      requestState.current.timers.totalTimer = null;
    }
  }, []);

  // 取消请求并保存已生成内容
  const cancelRequest = useCallback(
    (reason?: string) => {
      // 如果有生成的内容，保存为消息
      if (isLoading && (streamingContent || requestState.current.content.response)) {
        const finalContent = (requestState.current.content.response || streamingContent) + ERROR_MESSAGES.RESPONSE_INTERRUPTED;
        const aiMessage = buildMessageWithThinkContent(finalContent);
        addMessage(aiMessage);
      }

      // 调用纯清理函数
      cleanupResources();
      // 清除流式状态
      clearStreamingState();
      // 更新加载状态
      setIsLoading(false);
    },
    [isLoading, streamingContent, addMessage, cleanupResources, clearStreamingState],
  );

  // 创建流式请求
  const createStreamRequest = useCallback(async (url: string, options: FetchEventSourceInit, callbacks: StreamCallbacks) => {
    requestState.current.status.hasReceivedData = false;
    const enhancedOptions: FetchEventSourceInit = {
      ...options,
      async onopen(response) {
        // 检查响应状态
        if (!response.ok) {
          const errorText = await response.text().catch(() => 'Unknown error');
          console.error('流式请求打开失败:', response.status, response.statusText, errorText);
          throw new Error(`服务器返回错误: ${response.status} ${response.statusText} - ${errorText}`);
        }
        // 检查Content-Type是否为EventStream
        const contentType = response.headers.get('Content-Type');
        if (!contentType || !contentType.includes('text/event-stream')) {
          console.warn(`服务器未返回正确的流式数据类型，当前Content-Type: ${contentType}`);
          await callbacks.onFallbackResponse(response);
          throw new Error(ERROR_MESSAGES.NON_STREAMING);
        }
      },

      onmessage: (event) => {
        requestState.current.status.hasReceivedData = true;

        try {
          // 解析JSON响应
          let response;
          try {
            response = JSON.parse(event.data) as ChatResponseData;
            console.log('收到数据包:', response.data.type, { isComplete: response.data.is_complete });
          } catch (parseError: Error | any) {
            console.error('JSON解析失败，收到的数据:', event.data);
            throw new Error(`解析响应JSON失败: ${parseError.message}`);
          }
          callbacks.onDataReceived(response);

          if (response.data.is_complete) {
            callbacks.onComplete();
          }
        } catch (err: Error | any) {
          console.error('处理流数据失败:', err, event.data);
          if (options.onerror) options.onerror(err);
        }
      },
      onerror: (err) => {
        console.error('流式请求错误:', err);
        let errorMessage = '请求发送错误';

        if (err.message) {
          errorMessage = `请求发送错误: ${err.message}`;
        }

        setError(errorMessage);
        clearTimers();
        setIsLoading(false);
        if (options.onerror) {
          options.onerror(err);
        }
      },

      // 连接关闭处理
      onclose: () => {
        clearTimers();
        if (options.onclose) {
          options.onclose();
        }
      },
    };

    return fetchEventSource(url, enhancedOptions);
  }, []);

  const handleToolCalls = useCallback(
    async (toolCalls: IToolCall[], currentContent: string) => {
      // 1. 检查是否已在处理中
      if (requestState.current.status.isToolCallActive) {
        console.log('工具调用已在处理中，忽略新请求');
        return;
      }

      requestState.current.status.isToolCallActive = true;

      // 2. 初始化状态变量
      let currentToolMessageId = '';
      let toolCallResults: any[] = [];
      let allToolCallsCompleted = true;
      let currentContentList: any[] = [];

      try {
        // 3. 设置工具调用状态
        const _function = toolCalls[0].function;
        setStreamingContent(currentContent);

        // 4. 获取工具响应
        const toolResponse = await getIdByFunction({ toolName: _function.name, toolArgs: _function.arguments }, selectedMcpIds());

        // 5. 查找已存在的消息
        const messages = useChatStore.getState().messages;
        const existingToolMessage = findExistingToolMessage(messages, toolResponse.mcpId);

        if (existingToolMessage) {
          // 使用已存在的消息
          currentToolMessageId = existingToolMessage.id;
          currentContentList = [...(existingToolMessage.contentList || [])];

          const { toolCallResults: existingResults } = extractToolCallData(currentContentList, toolResponse.mcpId);
          toolCallResults = existingResults;
          allToolCallsCompleted = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');
        } else {
          // 查找进行中的工具调用消息
          const progressToolMessage = findProgressToolMessage(messages);

          if (progressToolMessage) {
            currentToolMessageId = progressToolMessage.id;
            currentContentList = [...(progressToolMessage.contentList || [])];

            const { toolCallResults: progressResults } = extractToolCallData(currentContentList);
            toolCallResults = progressResults;
            allToolCallsCompleted = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');
          }
        }

        // 6. 添加当前工具调用到结果数组
        toolCallResults.push(buildToolCallData(toolResponse, toolCalls[0], 'progress'));
        allToolCallsCompleted = false;

        // 7. 构建并更新消息
        const mcpContent = createMcpContent('progress', toolCallResults);
        currentContentList = currentContentList.length > 0 ? updateContentListWithMcp(currentContentList, mcpContent) : [mcpContent];

        const mcpMessage: MessageType = {
          id: currentToolMessageId || generateUniqueId('mcp_msg'),
          role: 'assistant',
          contentList: currentContentList,
        };

        currentToolMessageId = addMessage(mcpMessage, !!currentToolMessageId);

        // 8. 执行工具调用
        const data = await httpRequest.post('/mcp/client/runTool', toolResponse as IRunToolParams);
        const isToolError = data?.content?.isError === true;
        const toolErrorMessage = isToolError && Array.isArray(data.content) && data.content.length > 0 ? data.content[0]?.text || ERROR_MESSAGES.TOOL_EXECUTION_FAILED : '';

        setStreamingContent(currentContent);
        requestState.current.content.response = currentContent;

        // 9. 更新工具调用结果
        const lastIndex = toolCallResults.length - 1;
        toolCallResults[lastIndex] = {
          ...toolCallResults[lastIndex],
          outputParams: isToolError ? toolErrorMessage : data.content[0]?.text || '',
          status: isToolError ? 'error' : 'success',
        };

        // 10. 检查所有工具是否完成
        const allComplete = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');
        const finalStatus = allComplete ? (toolCallResults.some((t) => t.status === 'error') ? 'error' : 'success') : 'progress';

        // 11. 更新最终消息
        const updatedMcpContent = createMcpContent(finalStatus, toolCallResults);
        const updatedContentList = updateContentListWithMcp(currentContentList, updatedMcpContent);

        const updatedMessage: MessageType = {
          id: currentToolMessageId,
          role: 'assistant',
          contentList: updatedContentList,
        };

        addMessage(updatedMessage, true);

        // 12. 处理后续操作
        if (!isToolError && toolCallHandlersRef.current.continueConversation) {
          await toolCallHandlersRef.current.continueConversation(data.content[0].text);
        } else if (isToolError) {
          console.error('工具调用失败:', toolErrorMessage);
          const errorContent = currentContent + `\n\n[工具调用失败: ${toolErrorMessage}]`;
          // 使用 handleTextContent 处理错误内容
          handleTextContent({ content: errorContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
          requestState.current.content.response = errorContent;
        } else {
          console.error('工具调用处理程序未初始化');
          setError('工具调用处理失败: 内部错误');
        }
      } catch (error: any) {
        // 13. 错误处理
        console.error('工具调用失败:', error);
        const errorMessage = error.message || '未知错误';
        const updatedContent = currentContent + `\n\n[工具调用失败: ${errorMessage}]`;
        // 使用 handleTextContent 处理错误内容
        handleTextContent({ content: updatedContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
        requestState.current.content.response = updatedContent;

        // 更新错误消息
        handleToolCallErrorMessage(toolCallResults, currentContentList, currentToolMessageId, errorMessage, addMessage);
      } finally {
        // 14. 清理状态
        if (allToolCallsCompleted) {
          requestState.current.status.isToolCallActive = false;
        } else {
          console.log('工具调用未全部完成，保持活跃状态');
        }
      }
    },
    [selectedMcpIds, setStreamingContent, setError, addMessage],
  );

  // 使用工具结果继续对话
  const continueConversationWithResult = useCallback(
    async (toolResult: string) => {
      try {
        // 先中止旧的请求控制器，避免状态混乱
        if (abortControllerRef.current) {
          abortControllerRef.current.abort();
        }
        // 创建新的请求控制器
        abortControllerRef.current = new AbortController();
        // 构建继续对话的请求参数
        const continueRequestData: ChatRequestParams = {
          content: toolResult,
          SessionID: currentSessionId,
        };

        if (selectedMcpIds().length > 0) {
          continueRequestData.mcpIds = selectedMcpIds();
        }
        // 继续请求
        const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';
        // 保持当前的响应内容，用于累加
        let localResponseContent = requestState.current.content.response;

        await createStreamRequest(
          `${API_BASE_URL}${API_PREFIX}/playground/message/stream`,
          {
            method: 'POST',
            headers: baseHeaders(),
            body: JSON.stringify(continueRequestData),
            openWhenHidden: true,
            signal: abortControllerRef.current.signal,
          },
          {
            onDataReceived: async (response) => {
              // 重置超时定时器
              resetNoDataTimer();
              if (requestState.current.timers.totalTimer) {
                clearTimeout(requestState.current.timers.totalTimer);
                requestState.current.timers.totalTimer = null;
              }
              const data = response.data;
              // 保存会话ID
              if (data.session_id && (!currentSessionId || currentSessionId !== data.session_id)) {
                setCurrentSessionId(data.session_id);
              }
              // 处理工具调用或文本内容
              if (data?.tool_calls && data.tool_calls.length > 0) {
                await handleToolCalls(data.tool_calls, localResponseContent);
              } else if (data.content) {
                // 使用 handleTextContent 处理文本内容
                localResponseContent = handleTextContent(data, localResponseContent, setStreamingContent, setStreamingThinking, requestState);
                requestState.current.content.response = localResponseContent;
              }
            },
            onComplete: () => {
              // 获取当前所有消息
              const allMessages = useChatStore.getState().messages;

              // 检查是否有进行中的工具调用
              const hasActiveToolCall = allMessages.some((msg: any) =>
                msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem?.content?.status === 'progress'),
              );

              // 只有当没有活跃的工具调用和有内容时，才添加消息
              if (!requestState.current.status.isToolCallActive && !hasActiveToolCall && (requestState.current.content.response || localResponseContent)) {
                const finalContent = requestState.current.content.response || localResponseContent;
                const aiMessage = buildMessageWithThinkContent(finalContent);

                // 添加到消息列表
                addMessage(aiMessage);
              }

              // 如果所有工具调用都已完成，完全清理状态
              if (!hasActiveToolCall) {
                requestState.current.status.isToolCallActive = false;
                clearStreamingState();
              }

              setIsLoading(false);
              clearTimers();
            },
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                try {
                  const fullResponse = JSON.parse(fullResponseText);
                  if (fullResponse.error) {
                    console.error('工具调用返回错误:', fullResponse.error);
                    const errorContent = (requestState.current.content.response || streamingContent) + `\n\n[工具调用失败: ${fullResponse.error}]`;
                    // 使用 handleTextContent 处理错误内容
                    handleTextContent({ content: errorContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
                    requestState.current.content.response = errorContent;
                  } else if (fullResponse?.data?.content) {
                    // 使用 handleTextContent 处理响应内容
                    const processedContent = handleTextContent({ content: fullResponse.data.content, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
                    requestState.current.content.response = processedContent;
                  }
                } catch (parseError) {
                  console.error('解析工具调用响应失败:', parseError);
                  const errorContent = (requestState.current.content.response || streamingContent) + '\n\n[解析工具调用响应失败]';
                  handleTextContent({ content: errorContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
                  requestState.current.content.response = errorContent;
                }
              } catch (readError) {
                console.error('读取工具调用响应失败:', readError);
                const errorContent = (requestState.current.content.response || streamingContent) + '\n\n[读取工具调用响应失败]';
                handleTextContent({ content: errorContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
                requestState.current.content.response = errorContent;
              }
            },
          },
        );
      } catch (error: any) {
        console.error('继续对话失败:', error);
        setError(`继续对话失败: ${error.message}`);
      }
    },
    [createStreamRequest, currentSessionId, selectedMcpIds, resetNoDataTimer, setCurrentSessionId, setError, setStreamingContent, setStreamingThinking, streamingContent, handleToolCalls],
  );

  const sendChatMessageInternal = useCallback(
    async (userMessage: string, isResend = false) => {
      if (!userMessage.trim()) return;
      // 取消之前的请求
      cancelRequest('发送新消息');
      // 清除流状态
      clearStreamingState();
      // 模型检查
      if (!selectedModel) {
        setError(ERROR_MESSAGES.NO_MODEL_SELECTED);
        return;
      }
      // 记录最后一条消息用于重发
      if (!isResend) {
        lastUserMessageRef.current = userMessage;

        // 创建用户消息，仅在非重发时添加
        const userMsg: MessageType = {
          id: generateUniqueId('user_msg'),
          role: 'user',
          contentList: [
            {
              id: generateUniqueId('content'),
              type: 'plain',
              content: userMessage,
            },
          ],
        };
        addMessage(userMsg);
      }
      // 设置加载状态
      setIsLoading(true);
      // 创建请求控制器
      abortControllerRef.current = new AbortController();
      const signal = abortControllerRef.current.signal;
      // 构建请求参数
      const requestData: ChatRequestParams = {
        content: userMessage,
      };
      // 添加会话ID
      if (currentSessionId && currentSessionId.trim()) {
        requestData.SessionID = currentSessionId;
      }
      // 添加MCP IDs
      if (selectedMcpIds().length > 0) {
        requestData.mcpIds = selectedMcpIds();
      }
      try {
        const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';
        // 设置总超时
        requestState.current.timers.totalTimer = setTimeout(() => {
          console.log('总超时触发');
          setError(ERROR_MESSAGES.TOTAL_TIMEOUT);
          cancelRequest('总超时');
        }, TIMEOUT_CONFIG.TOTAL);
        // 在收到第一个数据前启动无数据超时计时器
        resetNoDataTimer();
        // 初始化响应内容
        let responseContent = '';
        await createStreamRequest(
          `${API_BASE_URL}${API_PREFIX}/playground/message/stream`,
          {
            method: 'POST',
            headers: baseHeaders(),
            body: JSON.stringify(requestData),
            openWhenHidden: true,
            signal,
            onerror: (error) => {
              console.error('流式请求错误:', error);
              setError(`请求失败: ${error.message}`);
              clearTimers();
              setIsLoading(false);
            },
            onclose: () => {
              clearTimers();

              // 如果连接关闭但未完成且有内容，保存部分回复
              if (isLoading && (requestState.current.content.response || responseContent)) {
                // 创建部分AI回复消息
                const finalContent = (requestState.current.content.response || responseContent) + ERROR_MESSAGES.CONNECTION_INTERRUPTED;
                const aiMessage = buildMessageWithThinkContent(finalContent);
                addMessage(aiMessage);
              }
              setIsLoading(false);
            },
          },
          {
            onDataReceived: async (response) => {
              resetNoDataTimer();
              if (requestState.current.timers.totalTimer) {
                clearTimeout(requestState.current.timers.totalTimer);
                requestState.current.timers.totalTimer = null;
              }
              const data = response.data;

              // 保存会话ID
              if (data.session_id && (!currentSessionId || currentSessionId !== data.session_id)) {
                setCurrentSessionId(data.session_id);
              }

              if (data?.tool_calls && data.tool_calls.length > 0) {
                // 处理工具调用
                await handleToolCalls(data.tool_calls, responseContent);
              } else if (data.content) {
                // 使用 handleTextContent 处理文本内容
                responseContent = handleTextContent(data, responseContent, setStreamingContent, setStreamingThinking, requestState);
                requestState.current.content.response = responseContent;
              }
            },

            // 完成回调
            onComplete: () => {
              if (!requestState.current.status.isToolCallActive && (requestState.current.content.response || responseContent)) {
                const finalContent = requestState.current.content.response || responseContent;
                const aiMessage = buildMessageWithThinkContent(finalContent);

                // 添加到消息列表
                addMessage(aiMessage);
              }
              clearStreamingState();
              setIsLoading(false);
              clearTimers();
            },

            // 非流式响应处理
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                console.log('接收到非流式响应');

                try {
                  const fullResponse = JSON.parse(fullResponseText);

                  // 检查是否包含错误信息
                  if (fullResponse.error) {
                    console.error('接收到错误响应:', fullResponse.error);
                    setError(`服务器返回错误: ${fullResponse.error}`);
                    cancelRequest('服务器错误');
                    return;
                  }

                  // 处理正常的非流式响应
                  if (fullResponse?.data?.content) {
                    const aiMessage = buildMessageWithThinkContent(fullResponse.data.content);
                    addMessage(aiMessage);
                  } else {
                    setError('服务器返回了非流式数据，但无法解析内容');
                  }
                } catch (parseError) {
                  console.error('解析非流式响应失败:', parseError);
                  setError('服务器返回了非流式数据，但解析失败');
                }
              } catch (readError) {
                console.error('读取响应内容失败:', readError);
                setError('无法读取服务器响应');
              }

              // 停止请求
              cancelRequest('非流式响应处理完成');
            },
          },
        );
      } catch (error: any) {
        console.error('发送聊天请求失败:', error);
        setError(`发送请求失败: ${error.message}`);
        setIsLoading(false);
        clearTimers();
      }
    },
    [selectedModel, addMessage, cancelRequest, clearStreamingState, currentSessionId, setCurrentSessionId, createStreamRequest, clearTimers, resetNoDataTimer, handleToolCalls],
  );

  const sendChatMessage = useCallback(
    (userMessage: string) => {
      if (isLoading) {
        console.log('正在加载中，忽略发送请求');
        return;
      }
      sendChatMessageInternal(userMessage);
    },
    [sendChatMessageInternal, isLoading],
  );

  // 复制消息到剪贴板
  const copyMessageToClipboard = useCallback((content: string) => {
    if (!content) return false;

    try {
      navigator.clipboard.writeText(content);
      return true;
    } catch (err) {
      console.error('复制到剪贴板失败:', err);
      return false;
    }
  }, []);

  // 重发最后一条消息
  const resendLastMessage = useCallback(async () => {
    if (isResending) return;

    try {
      setIsResending(true);
      setError(null);

      // 获取上次发送的消息内容
      const lastMessage = lastUserMessageRef.current;
      if (!lastMessage) {
        console.warn('没有可重发的消息');
        setIsResending(false);
        return;
      }
      const userMessage: MessageType = {
        id: generateUniqueId('user_msg'),
        role: 'user',
        contentList: [
          {
            id: generateUniqueId('content'),
            type: 'plain',
            content: lastMessage,
          },
        ],
      };
      addMessage(userMessage);

      await sendChatMessageInternal(lastMessage, true);
    } catch (error: any) {
      console.error('重发消息失败:', error);
      setError(`重发消息失败: ${error.message || '未知错误'}`);
    } finally {
      setIsResending(false);
    }
  }, [isResending, sendChatMessageInternal, addMessage]);

  // 初始化工具调用处理程序引用
  useEffect(() => {
    toolCallHandlersRef.current.continueConversation = continueConversationWithResult;
    return () => {
      toolCallHandlersRef.current.continueConversation = null;
    };
  }, [continueConversationWithResult]);

  useEffect(() => {
    return () => {
      cleanupResources();
    };
  }, [cleanupResources]);

  return {
    streamingContent,
    streamingThinking,
    isLoading,
    isResending,
    error,
    sendChatMessage,
    cancelRequest,
    resendLastMessage,
    copyMessageToClipboard,
  };
}
