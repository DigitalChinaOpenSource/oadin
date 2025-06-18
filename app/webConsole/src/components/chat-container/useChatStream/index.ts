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
import { accumulateResponseContent, processCompletedThinkTags, extractPlainContent, buildPlainMessage, parseAndAddThinkMessages } from './thinkContentUtils';
import { IContentItem, IRunToolParams, IToolCall, IToolCallData, StreamCallbacks, ChatRequestParams, ChatResponseData, IStreamData } from './types';

// 错误消息常量
const ERROR_MESSAGES = {
  NO_MODEL_SELECTED: '未选择模型，请先选择一个模型',
  NO_DATA_TIMEOUT: '请求超时：15秒未收到数据',
  TOTAL_TIMEOUT: '请求超时：30秒内未收到任何数据',
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
  // 状态与引用
  const { addMessage, setCurrentSessionId } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();

  // UI状态
  const [streamingContent, setStreamingContent] = useState<string>('');
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
    hasReceivedData: boolean;
    isToolCallActive: boolean;
    processedThinkIds: Set<string>;
  }>({
    responseContent: '',
    hasReceivedData: false,
    isToolCallActive: false,
    processedThinkIds: new Set<string>(),
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

  /**
   * 清除所有定时器
   */
  const clearTimers = useCallback(() => {
    if (timeoutRefsRef.current.noDataTimer) {
      clearTimeout(timeoutRefsRef.current.noDataTimer);
      timeoutRefsRef.current.noDataTimer = null;
    }

    if (timeoutRefsRef.current.totalTimer) {
      clearTimeout(timeoutRefsRef.current.totalTimer);
      timeoutRefsRef.current.totalTimer = null;
    }
  }, []);

  // 清除流式状态
  const clearStreamingState = useCallback(() => {
    setStreamingContent('');
    setError(null);

    // 重置请求状态
    requestStateRef.current = {
      responseContent: '',
      hasReceivedData: false,
      isToolCallActive: false,
      processedThinkIds: new Set<string>(),
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

  /**
   * 重置无数据超时定时器
   */
  const resetNoDataTimer = useCallback(() => {
    if (timeoutRefsRef.current.noDataTimer) {
      clearTimeout(timeoutRefsRef.current.noDataTimer);
    }

    timeoutRefsRef.current.noDataTimer = setTimeout(() => {
      console.log('无数据超时触发');
      setError(ERROR_MESSAGES.NO_DATA_TIMEOUT);

      // 直接执行清理操作
      cleanupResources();
      clearStreamingState();
      setIsLoading(false);
    }, TIMEOUT_CONFIG.NO_DATA);
  }, [cleanupResources, clearStreamingState]);

  // 取消请求并保存已生成内容
  const cancelRequest = useCallback(
    (reason?: string) => {
      // 如果有生成的内容，保存为消息
      if (isLoading && (streamingContent || requestStateRef.current.responseContent)) {
        const finalContent = (requestStateRef.current.responseContent || streamingContent) + ERROR_MESSAGES.RESPONSE_INTERRUPTED;

        // 处理任何未完成的 think 标签
        if (finalContent.includes('<think>')) {
          parseAndAddThinkMessages(finalContent, addMessage);
        }

        // 添加 plain 内容
        const plainMessage = buildPlainMessage(finalContent);
        if ((plainMessage?.contentList || []).length > 0) {
          addMessage(plainMessage);
        }
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
      if (requestStateRef.current.isToolCallActive) {
        console.log('工具调用已在处理中，忽略新请求');
        return;
      }

      requestStateRef.current.isToolCallActive = true;

      // 2. 初始化状态变量
      let currentToolMessageId = '';
      let toolCallResults: any[] = [];
      let allToolCallsCompleted = true;
      let currentContentList: any[] = [];

      try {
        // 3. 设置工具调用状态
        const _function = toolCalls[0].function;
        setToolCallStatus((prev) => ({
          isActive: true,
          toolName: _function.name,
          toolDesc: '工具调用',
          startTime: Date.now(),
          callCount: prev.callCount + 1,
        }));

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
        requestStateRef.current.responseContent = currentContent;

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
          setStreamingContent(errorContent);
          requestStateRef.current.responseContent = errorContent;
        } else {
          console.error('工具调用处理程序未初始化');
          setError('工具调用处理失败: 内部错误');
        }
      } catch (error: any) {
        // 13. 错误处理
        console.error('工具调用失败:', error);
        const errorMessage = error.message || '未知错误';
        const updatedContent = currentContent + `\n\n[工具调用失败: ${errorMessage}]`;

        setStreamingContent(updatedContent);
        requestStateRef.current.responseContent = updatedContent;

        // 更新错误消息
        handleToolCallErrorMessage(toolCallResults, currentContentList, currentToolMessageId, errorMessage, addMessage);
      } finally {
        // 14. 清理状态
        if (allToolCallsCompleted) {
          requestStateRef.current.isToolCallActive = false;
          setToolCallStatus((prev) => ({
            isActive: false,
            callCount: prev.callCount,
          }));
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
                // 累积响应内容
                localResponseContent = accumulateResponseContent(data, localResponseContent);
                requestStateRef.current.responseContent = localResponseContent;

                // 处理已完成的 think 标签
                requestStateRef.current.processedThinkIds = processCompletedThinkTags(localResponseContent, requestStateRef.current.processedThinkIds, addMessage);

                // 更新显示的内容（只显示 plain 部分）
                const plainContent = extractPlainContent(localResponseContent);
                setStreamingContent(plainContent);
              }
            },
            onComplete: () => {
              // 获取当前所有消息
              const allMessages = useChatStore.getState().messages;

              // 检查是否有进行中的工具调用
              const hasActiveToolCall = allMessages.some((msg) =>
                msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem?.content?.status === 'progress'),
              );

              // 只有当没有活跃的工具调用和有内容时，才添加消息
              if (!requestStateRef.current.isToolCallActive && !hasActiveToolCall && requestStateRef.current.responseContent) {
                // 使用 buildPlainMessage 只添加 plain 内容
                const plainMessage = buildPlainMessage(requestStateRef.current.responseContent);

                // 只有当有实际的 plain 内容时才添加消息
                if ((plainMessage?.contentList || []).length > 0) {
                  addMessage(plainMessage);
                }
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
              clearTimers();
            },
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                try {
                  const fullResponse = JSON.parse(fullResponseText);
                  if (fullResponse.error) {
                    console.error('工具调用返回错误:', fullResponse.error);
                    const errorContent = (requestStateRef.current.responseContent || streamingContent) + `\n\n[工具调用失败: ${fullResponse.error}]`;

                    // 直接设置错误内容
                    requestStateRef.current.responseContent = errorContent;
                    setStreamingContent(extractPlainContent(errorContent));
                  } else if (fullResponse?.data?.content) {
                    // 处理响应内容
                    const responseContent = fullResponse.data.content;

                    // 处理所有的 think 标签
                    parseAndAddThinkMessages(responseContent, addMessage);

                    // 设置最终内容
                    requestStateRef.current.responseContent = responseContent;
                    const plainContent = extractPlainContent(responseContent);
                    setStreamingContent(plainContent);
                  }
                } catch (parseError) {
                  console.error('解析工具调用响应失败:', parseError);
                  const errorContent = (requestStateRef.current.responseContent || streamingContent) + '\n\n[解析工具调用响应失败]';
                  requestStateRef.current.responseContent = errorContent;
                  setStreamingContent(errorContent);
                }
              } catch (readError) {
                console.error('读取工具调用响应失败:', readError);
                const errorContent = (requestStateRef.current.responseContent || streamingContent) + '\n\n[读取工具调用响应失败]';
                requestStateRef.current.responseContent = errorContent;
                setStreamingContent(errorContent);
              }
            },
          },
        );
      } catch (error: any) {
        console.error('继续对话失败:', error);
        setError(`继续对话失败: ${error.message}`);
      }
    },
    [createStreamRequest, currentSessionId, selectedMcpIds, resetNoDataTimer, setCurrentSessionId, setError, setStreamingContent, streamingContent, handleToolCalls, addMessage],
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
              clearTimers();
              setIsLoading(false);
            },
            onclose: () => {
              clearTimers();

              // 如果连接关闭但未完成且有内容，保存部分回复
              if (isLoading && (requestStateRef.current.responseContent || responseContent)) {
                // 创建部分AI回复消息
                const finalContent = (requestStateRef.current.responseContent || responseContent) + ERROR_MESSAGES.CONNECTION_INTERRUPTED;

                // 处理任何未完成的 think 标签
                if (finalContent.includes('<think>')) {
                  parseAndAddThinkMessages(finalContent, addMessage);
                }

                // 添加 plain 内容
                const plainMessage = buildPlainMessage(finalContent);
                if ((plainMessage?.contentList || []).length > 0) {
                  addMessage(plainMessage);
                }
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

              // 保存会话ID
              if (data.session_id && (!currentSessionId || currentSessionId !== data.session_id)) {
                setCurrentSessionId(data.session_id);
              }

              if (data?.tool_calls && data.tool_calls.length > 0) {
                // 处理工具调用
                await handleToolCalls(data.tool_calls, responseContent);
              } else if (data.content) {
                // 累积响应内容
                responseContent = accumulateResponseContent(data, responseContent);
                requestStateRef.current.responseContent = responseContent;

                // 处理已完成的 think 标签
                requestStateRef.current.processedThinkIds = processCompletedThinkTags(responseContent, requestStateRef.current.processedThinkIds, addMessage);

                // 更新显示的内容（只显示 plain 部分）
                const plainContent = extractPlainContent(responseContent);
                setStreamingContent(plainContent);
              }
            },

            // 完成回调
            onComplete: () => {
              if (!requestStateRef.current.isToolCallActive && requestStateRef.current.responseContent) {
                // 只添加 plain 内容的消息
                const plainMessage = buildPlainMessage(requestStateRef.current.responseContent);

                // 只有当有实际的 plain 内容时才添加消息
                if ((plainMessage?.contentList || []).length > 0) {
                  addMessage(plainMessage);
                }
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
                    const responseContent = fullResponse.data.content;

                    // 处理所有的 think 标签
                    parseAndAddThinkMessages(responseContent, addMessage);

                    // 添加 plain 内容消息
                    const plainMessage = buildPlainMessage(responseContent);
                    if ((plainMessage?.contentList || []).length > 0) {
                      addMessage(plainMessage);
                    }
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
    isLoading,
    isResending,
    error,
    sendChatMessage,
    cancelRequest,
    resendLastMessage,
    copyMessageToClipboard,
  };
}
