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
      if (requestStateRef.current.isToolCallActive) {
        console.log('工具调用已在处理中，忽略新请求');
        return;
      }

      requestStateRef.current.isToolCallActive = true;

      // 用于跟踪当前工具调用的消息ID
      let currentToolMessageId = '';
      // 用于保存所有工具调用的结果数组
      let toolCallResults: any[] = [];
      // 跟踪是否所有工具调用完成
      let allToolCallsCompleted = true;
      // 跟踪当前消息的contentList
      let currentContentList: any[] = [];

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

        // 检查是否已存在工具调用消息
        const existingToolMessage = useChatStore
          .getState()
          .messages.find(
            (msg) =>
              msg.role === 'assistant' &&
              msg.contentList?.some(
                (contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.data?.some((tool: any) => tool.mcpId === toolResponse.mcpId),
              ),
          );

        // 如果找到包含同一mcpId的消息，使用它
        if (existingToolMessage) {
          currentToolMessageId = existingToolMessage.id;
          currentContentList = [...(existingToolMessage.contentList || [])];

          // 查找该消息中包含mcpId的内容项
          const mcpContentIndex = currentContentList.findIndex(
            (content) => content.type === 'mcp' && typeof content.content === 'object' && content.content.data?.some((tool: any) => tool.mcpId === toolResponse.mcpId),
          );

          if (mcpContentIndex >= 0) {
            // 获取现有的工具调用数据
            toolCallResults = [...currentContentList[mcpContentIndex].content.data];

            // 检查现有工具调用是否都已完成
            allToolCallsCompleted = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');
          }
        } else {
          // 如果没找到同mcpId的消息，检查是否有进行中的工具调用消息
          const progressToolMessage = useChatStore
            .getState()
            .messages.find((msg) => msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress'));

          if (progressToolMessage) {
            currentToolMessageId = progressToolMessage.id;
            currentContentList = [...(progressToolMessage.contentList || [])];

            // 找到progress状态的内容项
            const progressContentIndex = currentContentList.findIndex((content) => content.type === 'mcp' && typeof content.content === 'object' && content.content.status === 'progress');

            if (progressContentIndex >= 0) {
              toolCallResults = [...currentContentList[progressContentIndex].content.data];
              allToolCallsCompleted = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');
            }
          }
        }

        // 添加当前进行中的工具到结果数组
        toolCallResults.push({
          mcpId: toolResponse.mcpId, // 添加mcpId用于标识
          name: toolResponse.toolName,
          desc: toolResponse.toolDesc || '工具调用',
          logo: '',
          inputParams: JSON.stringify(toolResponse.toolArgs),
          status: 'progress',
        });

        // 因为有新的进行中的工具调用，整体状态应该是进行中
        allToolCallsCompleted = false;

        // 构建或更新工具调用内容
        const mcpContent = {
          id: generateUniqueId('content'),
          type: 'mcp' as const,
          content: {
            status: 'progress',
            data: toolCallResults,
          },
        };

        // 如果已有消息，找到或替换MCP内容项
        if (currentContentList.length > 0) {
          const mcpIndex = currentContentList.findIndex((item) => item.type === 'mcp');
          if (mcpIndex >= 0) {
            // 替换现有的MCP内容项
            currentContentList[mcpIndex] = mcpContent;
          } else {
            // 添加新的MCP内容项
            currentContentList.push(mcpContent);
          }
        } else {
          // 创建新的内容列表
          currentContentList = [mcpContent];
        }

        // 构建消息对象
        const mcpMessage: MessageType = {
          id: currentToolMessageId || generateUniqueId('mcp_msg'),
          role: 'assistant',
          contentList: currentContentList,
        };

        // 更新或添加消息
        currentToolMessageId = addMessage(mcpMessage, !!currentToolMessageId);

        // 调用工具API
        const data = await httpRequest.post('/mcp/client/runTool', toolResponse as IRunToolParams);
        const isToolError = data?.content?.isError === true;
        const toolErrorMessage = isToolError && Array.isArray(data.content) && data.content.length > 0 ? data.content[0]?.text || ERROR_MESSAGES.TOOL_EXECUTION_FAILED : '';

        setStreamingContent(currentContent);
        requestStateRef.current.responseContent = currentContent;

        // 更新数组中最后一个工具的状态
        const lastIndex = toolCallResults.length - 1;
        toolCallResults[lastIndex] = {
          ...toolCallResults[lastIndex],
          outputParams: isToolError ? toolErrorMessage : data.content[0]?.text || '',
          status: isToolError ? 'error' : 'success',
          executionTime: Date.now() - (toolCallStatus.startTime || Date.now()),
        };

        // 检查所有工具是否完成
        const allComplete = toolCallResults.every((tool) => tool.status === 'success' || tool.status === 'error');

        // 更新MCP内容
        const updatedMcpContent = {
          id: generateUniqueId('content'),
          type: 'mcp' as const,
          content: {
            status: allComplete ? (toolCallResults.some((t) => t.status === 'error') ? 'error' : 'success') : 'progress',
            data: toolCallResults,
          },
        };

        // 再次找到或替换MCP内容项
        const updatedContentList = [...currentContentList];
        const mcpIndex = updatedContentList.findIndex((item) => item.type === 'mcp');
        if (mcpIndex >= 0) {
          updatedContentList[mcpIndex] = updatedMcpContent;
        } else {
          updatedContentList.push(updatedMcpContent);
        }

        // 构建更新后的消息对象
        const updatedMessage: MessageType = {
          id: currentToolMessageId,
          role: 'assistant',
          contentList: updatedContentList,
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

        // 如果有正在处理的工具调用，将其状态更新为错误
        if (toolCallResults.length > 0) {
          const lastIndex = toolCallResults.length - 1;
          toolCallResults[lastIndex] = {
            ...toolCallResults[lastIndex],
            outputParams: error.message || '未知错误',
            status: 'error',
            executionTime: Date.now() - (toolCallStatus.startTime || Date.now()),
          };

          // 更新MCP内容
          const errorMcpContent = {
            id: generateUniqueId('content'),
            type: 'mcp' as const,
            content: {
              status: toolCallResults.every((t) => t.status === 'error' || t.status === 'success') ? 'error' : 'progress',
              data: toolCallResults,
            },
          };

          // 更新内容列表
          const errorContentList = [...currentContentList];
          const mcpIndex = errorContentList.findIndex((item) => item.type === 'mcp');
          if (mcpIndex >= 0) {
            errorContentList[mcpIndex] = errorMcpContent;
          } else {
            errorContentList.push(errorMcpContent);
          }

          // 构建错误消息
          const errorMessage: MessageType = {
            id: currentToolMessageId || generateUniqueId('mcp_error_msg'),
            role: 'assistant',
            contentList: errorContentList,
          };

          addMessage(errorMessage, !!currentToolMessageId);
        } else {
          // 如果没有正在处理的工具调用，创建新的错误消息
          const mcpErrorMessage: MessageType = {
            id: generateUniqueId('mcp_error_msg'),
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
                      executionTime: Date.now() - (toolCallStatus.startTime || Date.now()),
                    },
                  ],
                },
              },
            ],
          };

          addMessage(mcpErrorMessage);
        }
      } finally {
        // 这里不重置工具调用状态，只在最终完成时重置
        if (allToolCallsCompleted) {
          requestStateRef.current.isToolCallActive = false;
          setToolCallStatus((prev) => ({
            isActive: false,
            callCount: prev.callCount,
          }));
        } else {
          // 如果工具调用还未全部完成，保持活跃状态
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
              // 重置超时定时器
              resetNoDataTimer();
              if (timeoutRefsRef.current.totalTimer) {
                clearTimeout(timeoutRefsRef.current.totalTimer);
                timeoutRefsRef.current.totalTimer = null;
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
                // 处理文本内容
                if (data.is_complete) {
                  localResponseContent = data.content;
                } else if (data.type === 'answer') {
                  if (localResponseContent.length === 0) {
                    localResponseContent = data.content;
                  } else {
                    localResponseContent += data.content;
                  }
                } else {
                  if (data.content.length > localResponseContent.length || !localResponseContent.includes(data.content.trim())) {
                    localResponseContent += data.content;
                  } else {
                    localResponseContent = data.content;
                  }
                }
                setStreamingContent(localResponseContent);
                requestStateRef.current.responseContent = localResponseContent;
              }
            },
            onComplete: () => {
              // 获取当前所有消息
              const allMessages = useChatStore.getState().messages;

              // 检查是否有进行中的工具调用
              const hasActiveToolCall = allMessages.some((msg) =>
                msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem?.content?.status === 'progress'),
              );

              // 只有当没有活跃的工具调用和有内容时，才添加纯文本消息
              if (!requestStateRef.current.isToolCallActive && !hasActiveToolCall && (requestStateRef.current.responseContent || localResponseContent)) {
                const aiMessage: MessageType = {
                  id: generateUniqueId('ai_msg'),
                  role: 'assistant',
                  contentList: [
                    {
                      id: generateUniqueId('content'),
                      type: 'plain',
                      content: requestStateRef.current.responseContent || localResponseContent,
                    },
                  ],
                };

                // 添加到消息列表
                addMessage(aiMessage);
              }

              // 如果所有工具调用都已完成，完全清理状态
              if (!hasActiveToolCall) {
                requestStateRef.current.isToolCallActive = false;
                setToolCallStatus((prev) => ({
                  isActive: false,
                  callCount: prev.callCount,
                }));
                clearStreamingState();
              }

              setIsLoading(false);
              cleanupResources();
            },
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                try {
                  const fullResponse = JSON.parse(fullResponseText);
                  if (fullResponse.error) {
                    console.error('工具调用返回错误:', fullResponse.error);
                    const errorContent = streamingContent + `\n\n[工具调用失败: ${fullResponse.error}]`;
                    setStreamingContent(errorContent);
                    requestStateRef.current.responseContent = errorContent;
                  } else if (fullResponse?.data?.content) {
                    setStreamingContent(fullResponse.data.content);
                    requestStateRef.current.responseContent = fullResponse.data.content;
                  }
                } catch (parseError) {
                  console.error('解析工具调用响应失败:', parseError);
                  const errorContent = streamingContent + '\n\n[解析工具调用响应失败]';
                  setStreamingContent(errorContent);
                  requestStateRef.current.responseContent = errorContent;
                }
              } catch (readError) {
                console.error('读取工具调用响应失败:', readError);
                const errorContent = streamingContent + '\n\n[读取工具调用响应失败]';
                setStreamingContent(errorContent);
                requestStateRef.current.responseContent = errorContent;
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
              resetNoDataTimer();
              if (timeoutRefsRef.current.totalTimer) {
                clearTimeout(timeoutRefsRef.current.totalTimer);
                timeoutRefsRef.current.totalTimer = null;
              }
              const data = response.data;
              if (data?.tool_calls && data.tool_calls.length > 0) {
                // 处理工具调用
                await handleToolCalls(data.tool_calls, responseContent);
              } else if (data.content) {
                // 处理文本内容
                if (data.is_complete) {
                  responseContent = data.content;
                } else if (data.type === 'answer') {
                  if (responseContent.length === 0) {
                    responseContent = data.content;
                  } else {
                    responseContent += data.content;
                  }
                } else {
                  if (data.content.length > responseContent.length || !responseContent.includes(data.content.trim())) {
                    responseContent += data.content;
                  } else {
                    responseContent = data.content;
                  }
                }
                setStreamingContent(responseContent);
                requestStateRef.current.responseContent = responseContent;
              }
            },

            // 完成回调
            onComplete: () => {
              if (!requestStateRef.current.isToolCallActive && (requestStateRef.current.responseContent || responseContent)) {
                const aiMessage: MessageType = {
                  id: generateUniqueId('ai_msg'),
                  role: 'assistant',
                  contentList: [
                    {
                      id: generateUniqueId('content'),
                      type: 'plain',
                      content: requestStateRef.current.responseContent || responseContent,
                    },
                  ],
                };

                // 添加到消息列表
                addMessage(aiMessage);
              }
              clearStreamingState();
              setIsLoading(false);
              cleanupResources();
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
                    const aiMessage: MessageType = {
                      id: generateUniqueId('ai_msg'),
                      role: 'assistant',
                      contentList: [
                        {
                          id: generateUniqueId('content'),
                          type: 'plain',
                          content: fullResponse.data.content,
                        },
                      ],
                    };
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
