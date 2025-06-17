import { useState, useCallback, useRef, useEffect } from 'react';
import { fetchEventSource, FetchEventSourceInit } from '@microsoft/fetch-event-source';
import { baseHeaders } from '@/utils';
import { API_PREFIX } from '@/constants';
import { MessageType } from '@res-utiles/ui-components';
import useChatStore from '@/components/chat-container/store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useSelectedModelStore from '@/store/useSelectedModel';
import { getIdByFunction } from '../../select-mcp/lib/useSelectMcpHelper';
import { httpRequest } from '@/utils/httpRequest';
import { ChatRequestParams, ChatResponseData, IRunToolParams, StreamCallbacks, IToolCall } from './types';
import { ERROR_MESSAGES } from './contant';
import { generateUniqueId, copyMessageToClipboard } from './utils';

// 超时设置（毫秒）
const TIMEOUT_CONFIG = {
  NO_DATA: 60000, // 60秒无数据超时
  TOTAL: 120000, // 120秒总超时
};

export function useChatStream(currentSessionId: string) {
  const { addMessage, setCurrentSessionId } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();

  const [streamingContent, setStreamingContent] = useState<string>('');
  const [streamingThinking, setStreamingThinking] = useState<string>('');
  const [isLoading, setIsLoading] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // 在状态定义部分添加工具调用状态
  const [toolCallStatus, setToolCallStatus] = useState<{
    isActive: boolean;
    toolName?: string;
    toolDesc?: string;
    startTime?: number;
    callCount: number;
  }>({
    isActive: false,
    callCount: 0,
  });

  // 请求控制
  const abortControllerRef = useRef<AbortController | null>(null);
  const requestStateRef = useRef<{
    responseContent: string;
    thinkingContent: string;
    hasReceivedData: boolean;
    isToolCallActive: boolean;
    responseId?: string;
  }>({
    responseContent: '',
    thinkingContent: '',
    hasReceivedData: false,
    isToolCallActive: false,
  });

  // 上下文保存
  const lastUserMessageRef = useRef<string | null>(null);
  const toolCallHandlersRef = useRef<{
    continueConversation: ((result: string) => Promise<void>) | null;
  }>({ continueConversation: null });

  // 定时器
  const timeoutRefsRef = useRef<{
    noDataTimer: NodeJS.Timeout | null;
    totalTimer: NodeJS.Timeout | null;
  }>({
    noDataTimer: null,
    totalTimer: null,
  });

  // 重置无数据超时定时器
  const resetNoDataTimer = useCallback(() => {
    if (timeoutRefsRef.current.noDataTimer) {
      clearTimeout(timeoutRefsRef.current.noDataTimer);
    }

    timeoutRefsRef.current.noDataTimer = setTimeout(() => {
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
    requestStateRef.current = {
      responseContent: '',
      thinkingContent: '',
      hasReceivedData: false,
      isToolCallActive: false,
      responseId: undefined,
    };
  }, []);

  const cleanupResources = useCallback(() => {
    // 取消网络请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    // 清理定时器
    if (timeoutRefsRef.current.noDataTimer) {
      clearTimeout(timeoutRefsRef.current.noDataTimer);
      timeoutRefsRef.current.noDataTimer = null;
    }

    if (timeoutRefsRef.current.totalTimer) {
      clearTimeout(timeoutRefsRef.current.totalTimer);
      timeoutRefsRef.current.totalTimer = null;
    }
  }, []);

  // 取消请求并保存已生成内容
  const cancelRequest = useCallback(
    (reason?: string) => {
      // 如果有生成的内容，保存为消息
      if (isLoading && streamingContent) {
        const aiMessage: MessageType = {
          id: generateUniqueId('msg'),
          role: 'assistant',
          contentList: [
            {
              id: generateUniqueId('content'),
              type: 'plain',
              content: streamingContent + ERROR_MESSAGES.RESPONSE_INTERRUPTED,
            },
          ],
        };

        addMessage(aiMessage);
      }

      // 调用纯清理函数
      cleanupResources();
      // 清除流式状态
      clearStreamingState();
      // 更新加载状态
      setIsLoading(false);
    },
    [isLoading, streamingContent, streamingThinking, addMessage, cleanupResources, clearStreamingState],
  );

  // 创建流式请求
  const createStreamRequest = useCallback(async (url: string, options: FetchEventSourceInit, callbacks: StreamCallbacks) => {
    requestStateRef.current.hasReceivedData = false;
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
        requestStateRef.current.hasReceivedData = true;

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
        cleanupResources();
        setIsLoading(false);
        if (options.onerror) {
          options.onerror(err);
        }
      },

      // 连接关闭处理
      onclose: () => {
        cleanupResources();
        if (options.onclose) {
          options.onclose();
        }
      },
    };

    return fetchEventSource(url, enhancedOptions);
  }, []);

  const handleToolCalls = useCallback(
    async (toolCalls: IToolCall[], currentContent: string) => {
      // 重置工具调用状态，确保每次都是新的工具调用
      let currentToolMessageId = '';
      let toolCallResults: any[] = [];
      let currentContentList: any[] = [];

      // 标记工具调用为活跃状态
      requestStateRef.current.isToolCallActive = true;

      try {
        const _function = toolCalls[0].function;
        setToolCallStatus((prev) => ({
          isActive: true,
          toolName: _function.name,
          toolDesc: '工具调用',
          startTime: Date.now(),
          callCount: prev.callCount + 1,
        }));

        setStreamingContent(currentContent);
        const toolResponse = await getIdByFunction({ toolName: _function.name, toolArgs: _function.arguments }, selectedMcpIds());

        // 创建全新的工具调用数据
        toolCallResults = [
          {
            mcpId: toolResponse.mcpId,
            name: toolResponse.toolName,
            desc: toolResponse.toolDesc || '工具调用',
            logo: '',
            inputParams: JSON.stringify(toolResponse.toolArgs),
            status: 'progress',
          },
        ];

        // 构建全新的MCP内容
        const mcpContent = {
          id: generateUniqueId('content'),
          type: 'mcp' as const,
          content: {
            status: 'progress',
            data: toolCallResults,
          },
        };

        // 创建全新的内容列表
        currentContentList = [mcpContent];

        // 构建消息对象
        const mcpMessage: MessageType = {
          id: generateUniqueId('mcp_msg'), // 始终生成新ID
          role: 'assistant',
          contentList: currentContentList,
        };

        // 添加新消息，不更新现有消息
        currentToolMessageId = addMessage(mcpMessage, false);

        // 调用工具API
        const data = await httpRequest.post('/mcp/client/runTool', toolResponse as IRunToolParams);
        const isToolError = data?.content?.isError === true;
        const toolErrorMessage = isToolError && Array.isArray(data.content) && data.content.length > 0 ? data.content[0]?.text || ERROR_MESSAGES.TOOL_EXECUTION_FAILED : '';

        setStreamingContent(currentContent);
        requestStateRef.current.responseContent = currentContent;

        // 更新工具调用状态
        toolCallResults[0] = {
          ...toolCallResults[0],
          outputParams: isToolError ? toolErrorMessage : data.content[0]?.text || '',
          status: isToolError ? 'error' : 'success',
        };

        // 更新MCP内容
        const updatedMcpContent = {
          id: generateUniqueId('content'),
          type: 'mcp' as const,
          content: {
            status: isToolError ? 'error' : 'success',
            data: toolCallResults,
          },
        };

        // 创建更新后的消息对象
        const updatedMessage: MessageType = {
          id: currentToolMessageId,
          role: 'assistant',
          contentList: [updatedMcpContent],
        };

        // 更新消息
        addMessage(updatedMessage, true);

        // 如果工具调用成功，继续对话
        if (!isToolError && toolCallHandlersRef.current.continueConversation) {
          await toolCallHandlersRef.current.continueConversation(data.content[0].text);
        } else if (isToolError) {
          console.error('工具调用失败:', toolErrorMessage);
          const errorContent = currentContent + `\n\n[工具调用失败: ${toolErrorMessage}]`;
          setStreamingContent(errorContent);
          requestStateRef.current.responseContent = errorContent;
        } else {
          console.error('工具调用处理程序未初始化');
          setError('工具调用处理失败: 内部错误');
        }
      } catch (error: any) {
        console.error('工具调用失败:', error);
        const errorMessage = `\n\n[工具调用失败: ${error.message || '未知错误'}]`;
        const updatedContent = currentContent + errorMessage;
        setStreamingContent(updatedContent);
        requestStateRef.current.responseContent = updatedContent;

        // 创建新的错误消息
        const mcpErrorMessage: MessageType = {
          id: currentToolMessageId || generateUniqueId('mcp_error_msg'),
          role: 'assistant',
          contentList: [
            {
              id: generateUniqueId('content'),
              type: 'mcp',
              content: {
                status: 'error',
                data: [
                  {
                    mcpId: selectedMcpIds()[0] || '',
                    name: toolCalls[0].function.name,
                    desc: '工具调用',
                    logo: '',
                    inputParams: JSON.stringify(toolCalls[0].function.arguments),
                    outputParams: error.message || '未知错误',
                    status: 'error',
                  },
                ],
              },
            },
          ],
        };

        // 始终添加新消息，不更新现有消息
        addMessage(mcpErrorMessage, currentToolMessageId ? true : false);
      } finally {
        // 检查是否还有进行中的工具调用
        const allMessages = useChatStore.getState().messages;
        const hasActiveToolCall = allMessages.some((msg) =>
          msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress'),
        );

        if (!hasActiveToolCall) {
          requestStateRef.current.isToolCallActive = false;
          setToolCallStatus((prev) => ({
            isActive: false,
            callCount: prev.callCount,
          }));
        } else {
          console.log('仍有工具调用进行中，保持活跃状态');
        }
      }
    },
    [selectedMcpIds, setStreamingContent, setError, addMessage],
  );

  // 处理onComplete逻辑
  const handleStreamComplete = useCallback(
    (responseContent: string) => {
      // 获取当前所有消息
      const allMessages = useChatStore.getState().messages;

      // 检查是否有进行中的工具调用
      const hasActiveToolCall = allMessages.some((msg) =>
        msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem?.content?.status === 'progress'),
      );

      // 只有当没有活跃的工具调用和有内容时，才添加纯文本消息
      if (!requestStateRef.current.isToolCallActive && !hasActiveToolCall && responseContent) {
        const aiMessage: MessageType = {
          id: requestStateRef.current.responseId || generateUniqueId('ai_msg'),
          role: 'assistant',
          contentList: [
            {
              id: generateUniqueId('content'),
              type: 'plain',
              content: responseContent,
            },
          ],
        };
        addMessage(aiMessage);
      }

      // 如果存在工具调用消息但所有工具调用已完成，则更新工具调用消息状态为最终状态
      const toolMessages = allMessages.filter((msg) => msg.contentList?.some((contentItem) => contentItem.type === 'mcp' && typeof contentItem.content === 'object'));

      toolMessages.forEach((message: any) => {
        const mcpContentIndex = message.contentList.findIndex((item: any) => item.type === 'mcp' && typeof item.content === 'object');

        if (mcpContentIndex >= 0) {
          const mcpContent = message.contentList[mcpContentIndex];
          const toolCallData = mcpContent.content.data || [];

          // 如果所有工具调用都已完成（不是progress状态）
          const allToolCallsCompleted = !toolCallData.some((tool: any) => tool.status === 'progress');

          if (allToolCallsCompleted && mcpContent.content.status === 'progress') {
            // 检查是否有任何错误
            const hasErrors = toolCallData.some((tool: any) => tool.status === 'error');

            // 创建更新后的消息副本
            const updatedContentList = [...message.contentList];
            updatedContentList[mcpContentIndex] = {
              ...mcpContent,
              content: {
                ...mcpContent.content,
                status: hasErrors ? 'error' : 'success',
              },
            };

            // 更新消息
            const updatedMessage = {
              ...message,
              contentList: updatedContentList,
            };

            // 更新消息状态
            addMessage(updatedMessage, true);
          }
        }
      });

      // 如果所有工具调用都已完成，完全清理状态
      if (!hasActiveToolCall) {
        requestStateRef.current.isToolCallActive = false;
        setToolCallStatus((prev) => ({
          isActive: false,
          callCount: prev.callCount,
        }));
        clearStreamingState();
      } else {
        // 如果还有活跃工具调用，记录日志
        console.log('仍有工具调用进行中，保持工具调用活跃状态');
      }

      // 完成流程处理
      setIsLoading(false);
      cleanupResources();
    },
    [addMessage, clearStreamingState, cleanupResources],
  );
  // 处理onDataReceived逻辑
  const handleStreamData = useCallback(
    async (data: any, currentContent: string, updateSessionId = false) => {
      // 重置超时定时器
      resetNoDataTimer();
      if (timeoutRefsRef.current.totalTimer) {
        clearTimeout(timeoutRefsRef.current.totalTimer);
        timeoutRefsRef.current.totalTimer = null;
      }

      // 保存会话ID (如果需要)
      if (updateSessionId && data.session_id && (!currentSessionId || currentSessionId !== data.session_id)) {
        setCurrentSessionId(data.session_id);
      }

      let updatedContent = currentContent;

      // 处理工具调用或文本内容
      if (data?.tool_calls && data.tool_calls.length > 0) {
        // 处理工具调用
        await handleToolCalls(data.tool_calls, currentContent);
      } else if (data.content) {
        // 处理文本内容
        if (data.is_complete) {
          updatedContent = data.content;
          if (data.id) {
            requestStateRef.current.responseId = data.id;
          }
        } else if (data.type === 'answer') {
          if (currentContent.length === 0) {
            updatedContent = data.content;
          } else {
            updatedContent += data.content;
          }
        } else {
          if (data.content.length > currentContent.length || !currentContent.includes(data.content.trim())) {
            updatedContent += data.content;
          } else {
            updatedContent = data.content;
          }
        }

        setStreamingContent(updatedContent);
        requestStateRef.current.responseContent = updatedContent;
      }

      return updatedContent;
    },
    [resetNoDataTimer, handleToolCalls, setStreamingContent, currentSessionId, setCurrentSessionId],
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
        let localResponseContent = requestStateRef.current.responseContent;

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
              localResponseContent = await handleStreamData(response.data, localResponseContent, true);
            },
            // 修改onComplete回调处理逻辑
            onComplete: () => {
              handleStreamComplete(requestStateRef.current.responseContent || localResponseContent);
            },
            onFallbackResponse: async (response) => {
              cancelRequest('非流式响应');
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

      // 确保工具调用状态重置
      requestStateRef.current.isToolCallActive = false;
      setToolCallStatus((prev) => ({
        isActive: false,
        callCount: prev.callCount,
      }));

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
        timeoutRefsRef.current.totalTimer = setTimeout(() => {
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
              cleanupResources();
              setIsLoading(false);
            },
            onclose: () => {
              cleanupResources();

              // 如果连接关闭但未完成且有内容，保存部分回复
              if (isLoading && responseContent) {
                // 创建部分AI回复消息
                const aiMessage: MessageType = {
                  id: generateUniqueId('ai_msg'),
                  role: 'assistant',
                  contentList: [
                    {
                      id: generateUniqueId('content'),
                      type: 'plain',
                      content: responseContent + ERROR_MESSAGES.CONNECTION_INTERRUPTED,
                    },
                  ],
                };
                addMessage(aiMessage);
              }
              setIsLoading(false);
            },
          },
          {
            onDataReceived: async (response) => {
              responseContent = await handleStreamData(response.data, responseContent);
            },

            onComplete: () => {
              handleStreamComplete(requestStateRef.current.responseContent || responseContent);
            },

            // 非流式响应处理
            onFallbackResponse: async (response) => {
              // 停止请求
              cancelRequest('非流式响应处理完成');
            },
          },
        );
      } catch (error: any) {
        console.error('发送聊天请求失败:', error);
        setError(`发送请求失败: ${error.message}`);
        setIsLoading(false);
        cleanupResources();
      }
    },
    [selectedModel, addMessage, cancelRequest, clearStreamingState, currentSessionId, setCurrentSessionId, createStreamRequest, cleanupResources, resetNoDataTimer, handleToolCalls],
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

  // 重发最后一条消息
  const resendLastMessage = useCallback(async () => {
    if (isResending) return;

    try {
      setIsResending(true);
      setError(null);

      // 重置工具调用状态
      requestStateRef.current.isToolCallActive = false;
      setToolCallStatus((prev) => ({
        isActive: false,
        callCount: prev.callCount,
      }));

      // 获取上次发送的消息内容
      const lastMessage = lastUserMessageRef.current;
      if (!lastMessage) {
        console.warn('没有可重发的消息');
        setIsResending(false);
        return;
      }

      // 添加用户消息
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
  };
}
