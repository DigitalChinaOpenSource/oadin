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
import { generateUniqueId, formatErrorMessage } from './utils';
import {
  findProgressToolMessage,
  extractToolCallData,
  extractToolCallDataByGroupId,
  createMcpContent,
  createMcpContentWithGroupId,
  updateContentListWithMcp,
  updateContentListWithMcpByGroupId,
  buildToolCallData,
  handleToolCallErrorMessage,
} from './mcpContentUtils';
import { buildMessageWithThinkContent, handleTextContent } from './thinkContentUtils';
import { IStreamData, StreamCallbacks, ChatRequestParams, ChatResponseData, IToolCallData } from './types';
import { ERROR_MESSAGES, TIMEOUT_CONFIG, ErrorType } from './contants';

export function useChatStream() {
  const { addMessage, currentSessionId, isLoading, setIsLoading } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();

  const [streamingContent, setStreamingContent] = useState<string>('');
  const [streamingThinking, setStreamingThinking] = useState<string>('');

  const [isResending, setIsResending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 流式请求控制器
  const abortControllerRef = useRef<AbortController | null>(null);

  // 请求状态统一处理
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
    lastToolGroupIdRef: null as string | null,
  });
  // 上下文保存
  const lastUserMessageRef = useRef<string | null>(null);
  const toolCallHandlersRef = useRef<{
    continueConversation: ((result: string) => Promise<void>) | null;
  }>({ continueConversation: null });
  const functionIdCacheRef = useRef<Record<string, any>>({});

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

  const handleError = (
    message: string,
    originalError?: any,
    errorType: ErrorType = ErrorType.INTERNAL,
    options: {
      shouldCancel?: boolean;
      appendToContent?: boolean;
      updateMcpStatus?: boolean;
      toolGroupId?: string; // 新增：指定要更新的工具组ID
    } = {},
  ) => {
    const { shouldCancel = true, appendToContent = false, updateMcpStatus = true, toolGroupId } = options;

    // 设置错误状态
    setError(message);

    // 更新MCP状态 - 找到当前进行中的工具调用并标记为错误
    if (updateMcpStatus) {
      const allMessages = useChatStore.getState().messages;
      const progressToolMessage = findProgressToolMessage(allMessages) as any;

      if (progressToolMessage) {
        const updatedContentList = [...progressToolMessage.contentList];
        let hasUpdates = false;

        if (toolGroupId) {
          // 如果指定了 toolGroupId，只更新该工具组
          const { toolCallResults, mcpContentIndex } = extractToolCallDataByGroupId(progressToolMessage.contentList, toolGroupId);

          if (toolCallResults.length > 0 && mcpContentIndex >= 0) {
            // 将最后一个工具调用标记为错误
            const lastIndex = toolCallResults.length - 1;
            if (toolCallResults[lastIndex].status === 'progress') {
              toolCallResults[lastIndex] = {
                ...toolCallResults[lastIndex],
                outputParams: message,
                status: 'error',
              };

              // 更新消息状态
              const currentTotalDuration = progressToolMessage.contentList[mcpContentIndex].content.totalDuration || 0;
              const updatedMcpContent = createMcpContentWithGroupId(toolGroupId, 'error', toolCallResults, currentTotalDuration);
              updatedContentList[mcpContentIndex] = updatedMcpContent;
              hasUpdates = true;
            }
          }
        } else {
          // 如果没有指定 toolGroupId，将所有进行中的工具组标记为错误
          updatedContentList.forEach((contentItem, index) => {
            if (contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress') {
              const toolCallResults = [...(contentItem.content.data || [])];

              // 将所有进行中的工具调用标记为错误
              toolCallResults.forEach((tool, toolIndex) => {
                if (tool.status === 'progress') {
                  toolCallResults[toolIndex] = {
                    ...tool,
                    outputParams: message,
                    status: 'error',
                  };
                }
              });

              // 更新工具组状态
              const updatedMcpContent = createMcpContentWithGroupId(contentItem.id, 'error', toolCallResults, contentItem.content.totalDuration || 0);
              updatedContentList[index] = updatedMcpContent;
              hasUpdates = true;
            }
          });
        }

        // 如果有更新，保存消息
        if (hasUpdates) {
          const errorMessage: MessageType = {
            id: progressToolMessage.id,
            role: 'assistant',
            contentList: updatedContentList,
          };

          addMessage(errorMessage, true);
        }
      }
    }

    // 添加错误到内容
    if (appendToContent && (requestState.current.content.response || streamingContent)) {
      const currentContent = requestState.current.content.response || streamingContent;
      const contentWithError = `${currentContent}\n\n[${message}]`;

      requestState.current.content.response = contentWithError;
      setStreamingContent(contentWithError);
    }

    if (shouldCancel) {
      cancelRequest();
    }
  };

  const resetNoDataTimer = useCallback(() => {
    if (requestState.current.timers.noDataTimer) {
      clearTimeout(requestState.current.timers.noDataTimer);
    }

    requestState.current.timers.noDataTimer = setTimeout(() => {
      handleError(ERROR_MESSAGES.TIMEOUT.NO_DATA, null, ErrorType.TIMEOUT, {
        shouldCancel: true,
        appendToContent: true,
        updateMcpStatus: true,
      });
    }, TIMEOUT_CONFIG.NO_DATA);
  }, [handleError]);

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
      lastToolGroupIdRef: null,
    };
    functionIdCacheRef.current = {};
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
  const cancelRequest = () => {
    // 如果有生成的内容，保存为消息
    if (isLoading && (streamingContent || requestState.current.content.response)) {
      const finalContent = (requestState.current.content.response || streamingContent) + ERROR_MESSAGES.CONNECTION.RESPONSE_INTERRUPTED;
      const aiMessage = buildMessageWithThinkContent(finalContent);
      addMessage(aiMessage);
    }

    // 调用纯清理函数
    cleanupResources();
    // 清除流式状态
    clearStreamingState();
    // 更新加载状态
    setIsLoading(false);
  };

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
          throw new Error(ERROR_MESSAGES.RESPONSE.NON_STREAMING);
        }
      },

      onmessage: (event) => {
        requestState.current.status.hasReceivedData = true;

        try {
          // 解析JSON响应
          let response;
          try {
            response = JSON.parse(event.data) as ChatResponseData;
          } catch (parseError: Error | any) {
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
        const errorMessage = err.message ? formatErrorMessage(ERROR_MESSAGES.REQUEST.FAILED, err.message) : '请求发送错误';

        handleError(errorMessage, err, ErrorType.REQUEST, {
          shouldCancel: false,
          appendToContent: true,
          updateMcpStatus: true, // 确保更新MCP状态
        });

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
    async (data: IStreamData, currentContent: any) => {
      const { tool_calls, tool_group_id, id, total_duration } = data;
      console.log('tool_group_id====>', tool_group_id);
      // 如果返回的 content 为空且有 tool_calls，保存 tool_group_id 用于下一次请求
      // (!currentContent || currentContent.trim() === '') &&
      if (tool_calls && tool_calls.length > 0 && tool_group_id) {
        requestState.current.lastToolGroupIdRef = tool_group_id;
      }
      if (!tool_calls || tool_calls.length === 0) {
        requestState.current.lastToolGroupIdRef = null;
        return;
      }

      // 1. 检查是否已在处理中
      if (requestState.current.status.isToolCallActive) {
        console.log('工具调用已在处理中，忽略新请求');
        return;
      }

      requestState.current.status.isToolCallActive = true;

      // 2. 初始化状态变量
      let toolCallResults: any[] = [];
      let currentContentList: any[] = [];
      let currentTotalDuration = 0;
      try {
        // 3. 设置工具调用状态
        setStreamingContent(currentContent);
        const _function = tool_calls?.[0].function;
        // 4. 获取工具响应
        let toolResponse;

        // 检查缓存中是否已有此函数的id
        if (_function.name && functionIdCacheRef.current[_function.name]) {
          toolResponse = functionIdCacheRef.current[_function.name];
        } else {
          toolResponse = await getIdByFunction({ toolName: _function.name, toolArgs: _function.arguments }, selectedMcpIds());

          if (_function.name) {
            functionIdCacheRef.current[_function.name] = toolResponse;
          }
        }

        // 5. 查找已存在的消息
        const messages = useChatStore.getState().messages;
        // 查找进行中的工具调用消息
        const progressToolMessage = findProgressToolMessage(messages);
        if (progressToolMessage) {
          currentContentList = [...(progressToolMessage.contentList || [])];

          if (tool_group_id) {
            // 查找指定 tool_group_id 的现有数据
            const { toolCallResults: existingResults, mcpContentIndex } = extractToolCallDataByGroupId(currentContentList, tool_group_id);

            if (mcpContentIndex >= 0) {
              // 如果找到现有的 tool_group_id，使用其数据
              toolCallResults = existingResults;
              currentTotalDuration = currentContentList[mcpContentIndex]?.content?.totalDuration || 0;
            }
          } else {
            // 兼容没有 tool_group_id 的情况
            const { toolCallResults: _toolCallResults } = extractToolCallData(currentContentList);
            toolCallResults = _toolCallResults;
          }
        }
        // 累加当前的调用时间
        if (total_duration) {
          currentTotalDuration += total_duration;
        }

        // 6. 添加当前工具调用到结果数组
        const mcpId = toolResponse.mcpId || generateUniqueId('mcp');
        toolCallResults.push(buildToolCallData(toolResponse, mcpId, 'progress'));

        // 7. 构建并更新消息
        let mcpContent;
        if (tool_group_id) {
          mcpContent = createMcpContentWithGroupId(tool_group_id, 'progress', toolCallResults, currentTotalDuration);
          currentContentList = updateContentListWithMcpByGroupId(currentContentList, mcpContent, tool_group_id);
        } else {
          mcpContent = createMcpContent('progress', toolCallResults, currentTotalDuration);
          currentContentList = currentContentList.length > 0 ? updateContentListWithMcp(currentContentList, mcpContent) : [mcpContent];
        }

        const mcpMessage: MessageType = {
          id: id,
          role: 'assistant',
          contentList: currentContentList,
        };

        addMessage(mcpMessage, !!id);

        // 8. 执行工具调用
        const data = await httpRequest.post('/mcp/client/runTool', { messageId: id, ...toolResponse });
        const isToolError = data?.content?.isError === true;
        const toolErrorMessage = isToolError && Array.isArray(data.content) && data.content.length > 0 ? data.content[0]?.text || ERROR_MESSAGES.TOOL.EXECUTION_FAILED : '';

        setStreamingContent(currentContent);
        requestState.current.content.response = currentContent;

        // 9. 更新工具调用结果，仅更新单个工具的状态，整体状态保持为progress
        const lastIndex = toolCallResults.length - 1;
        toolCallResults[lastIndex] = {
          ...toolCallResults[lastIndex],
          outputParams: isToolError ? toolErrorMessage : data.content[0]?.text || '',
          logo: data.logo,
          desc: data.toolDesc,
          status: isToolError ? 'error' : 'success',
        } as IToolCallData;

        // 10. 检查所有工具是否完成,但仍保持整体状态为进行中
        const keepInProgress = !isToolError || (isToolError && toolCallHandlersRef.current.continueConversation);
        const finalStatus = keepInProgress ? 'progress' : 'error';

        // 11. 更新最终消息
        let updatedMcpContent;
        let updatedContentList;
        if (tool_group_id) {
          updatedMcpContent = createMcpContentWithGroupId(tool_group_id, finalStatus, toolCallResults);
          updatedContentList = updateContentListWithMcpByGroupId(currentContentList, updatedMcpContent, tool_group_id);
        } else {
          updatedMcpContent = createMcpContent(finalStatus, toolCallResults);
          updatedContentList = updateContentListWithMcp(currentContentList, updatedMcpContent);
        }

        const updatedMessage: MessageType = {
          id: id,
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
        const errorMessage = error.message || '未知错误';
        handleError(formatErrorMessage(ERROR_MESSAGES.TOOL.EXECUTION_FAILED, errorMessage), error, ErrorType.TOOL, {
          shouldCancel: false,
          appendToContent: true,
          updateMcpStatus: true,
          toolGroupId: tool_group_id,
        });

        // 更新错误消息
        if (tool_group_id) {
          handleToolCallErrorMessage(toolCallResults, currentContentList, id, tool_group_id, errorMessage, addMessage);
        } else {
          // 兼容旧版本错误处理
          handleToolCallErrorMessage(toolCallResults, currentContentList, id, '', errorMessage, addMessage);
        }
      } finally {
        // 14. 清理状态 - 只在确认对话继续或出错无法继续时才重置工具调用状态
        if (!toolCallHandlersRef.current.continueConversation) {
          requestState.current.status.isToolCallActive = false;
        }
      }
    },
    [selectedMcpIds, setStreamingContent, setError, addMessage],
  );

  // 使用工具结果继续对话
  const continueConversationWithResult = useCallback(
    async (toolResult: string) => {
      try {
        if (!currentSessionId) return;
        // 中止旧的请求控制器，避免状态混乱
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
        console.log('继续对话使用上次的toolGroupId:', requestState.current.lastToolGroupIdRef);
        if (requestState.current.lastToolGroupIdRef) {
          continueRequestData.toolGroupID = requestState.current.lastToolGroupIdRef;
        }
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
              if (data?.tool_calls && data.tool_calls.length > 0) {
                await handleToolCalls(data, localResponseContent);
              } else if (data.content) {
                localResponseContent = handleTextContent(data, localResponseContent, setStreamingContent, setStreamingThinking, requestState);
                requestState.current.content.response = localResponseContent;
              }
            },
            onComplete: () => {
              // 检查是否有进行中的工具调用
              const allMessages = useChatStore.getState().messages;
              const progressToolMessage = findProgressToolMessage(allMessages) as any;

              // 如果有进行中的工具调用，将其标记为完成
              if (progressToolMessage && !requestState.current.status.isToolCallActive) {
                const updatedContentList = [...progressToolMessage.contentList];
                let hasUpdates = false;

                // 遍历所有 MCP 类型的内容项
                updatedContentList.forEach((contentItem, index) => {
                  if (contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress') {
                    const toolCallResults = contentItem.content.data || [];

                    // 检查当前工具组是否已完成
                    const allComplete = toolCallResults.every((tool: any) => tool.status === 'success' || tool.status === 'error');

                    if (allComplete) {
                      const hasError = toolCallResults.some((tool: any) => tool.status === 'error');
                      const finalStatus = hasError ? 'error' : 'success';

                      // 更新工具组状态 - 使用新的函数
                      const updatedMcpContent = createMcpContentWithGroupId(
                        contentItem.id, // tool_group_id
                        finalStatus,
                        toolCallResults,
                        contentItem.content.totalDuration || 0,
                      );

                      updatedContentList[index] = updatedMcpContent;
                      hasUpdates = true;
                    }
                  }
                });

                // 如果有更新，保存消息
                if (hasUpdates) {
                  const finalMessage: MessageType = {
                    id: progressToolMessage.id,
                    role: 'assistant',
                    contentList: updatedContentList,
                  };

                  addMessage(finalMessage, true);
                }
              }

              if (!requestState.current.status.isToolCallActive && (requestState.current.content.response || localResponseContent)) {
                const finalContent = requestState.current.content.response || localResponseContent;
                const aiMessage = buildMessageWithThinkContent(finalContent);
                addMessage(aiMessage);
              }

              clearStreamingState();
              setIsLoading(false);
              clearTimers();
            },
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                try {
                  const fullResponse = JSON.parse(fullResponseText);
                  if (fullResponse.error) {
                    handleError(formatErrorMessage(ERROR_MESSAGES.TOOL.EXECUTION_FAILED, fullResponse.error), { error: fullResponse.error }, ErrorType.TOOL, {
                      shouldCancel: false,
                      appendToContent: true,
                    });
                  } else if (fullResponse?.data?.content) {
                    // 使用 handleTextContent 处理响应内容
                    const processedContent = handleTextContent({ content: fullResponse.data.content, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState);
                    requestState.current.content.response = processedContent;
                  }
                } catch (parseError) {
                  handleError(ERROR_MESSAGES.PARSING.TOOL_RESPONSE, parseError, ErrorType.PARSING, { shouldCancel: false, appendToContent: true });
                }
              } catch (readError) {
                handleError(formatErrorMessage(ERROR_MESSAGES.CONNECTION.READ_FAILED, '读取工具调用响应失败'), readError, ErrorType.CONNECTION, { shouldCancel: false, appendToContent: true });
              }
            },
          },
        );
      } catch (error: any) {
        handleError(formatErrorMessage(ERROR_MESSAGES.TOOL.CONTINUE_FAILED, error.message || '未知错误'), error, ErrorType.TOOL);
      }
    },
    [createStreamRequest, currentSessionId, selectedMcpIds, resetNoDataTimer, setError, setStreamingContent, setStreamingThinking, handleToolCalls],
  );

  const sendChatMessageInternal = useCallback(
    async (userMessage: string, isResend = false) => {
      if (!userMessage.trim()) return;
      if (!currentSessionId) return;
      // 取消之前的请求
      cancelRequest();
      // 清除流状态
      clearStreamingState();
      // 模型检查
      if (!selectedModel) {
        handleError(ERROR_MESSAGES.REQUEST.NO_MODEL_SELECTED, null, ErrorType.REQUEST, { shouldCancel: false });
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
          handleError(ERROR_MESSAGES.TIMEOUT.TOTAL, null, ErrorType.TIMEOUT, { shouldCancel: true, appendToContent: true, updateMcpStatus: true });
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
              setError(`请求失败: ${error.message}`);
              clearTimers();
              setIsLoading(false);
            },
            onclose: () => {
              clearTimers();
              // 如果连接关闭但未完成且有内容，保存部分回复
              if (isLoading && (requestState.current.content.response || responseContent)) {
                // 创建部分AI回复消息
                const finalContent = (requestState.current.content.response || responseContent) + ERROR_MESSAGES.CONNECTION.INTERRUPTED;
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
              if (data?.tool_calls && data.tool_calls.length > 0) {
                await handleToolCalls(data, responseContent);
              } else if (data.content) {
                responseContent = handleTextContent(data, responseContent, setStreamingContent, setStreamingThinking, requestState);
                requestState.current.content.response = responseContent;
              }
            },
            onComplete: () => {
              // 检查是否有进行中的工具调用
              const allMessages = useChatStore.getState().messages;
              const progressToolMessage = findProgressToolMessage(allMessages) as any;

              // 如果有进行中的工具调用，将其标记为完成
              if (progressToolMessage && !requestState.current.status.isToolCallActive) {
                const updatedContentList = [...progressToolMessage.contentList];
                let hasUpdates = false;

                // 遍历所有 MCP 类型的内容项
                updatedContentList.forEach((contentItem, index) => {
                  if (contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress') {
                    const toolCallResults = contentItem.content.data || [];

                    // 检查当前工具组是否已完成
                    const allComplete = toolCallResults.every((tool: any) => tool.status === 'success' || tool.status === 'error');

                    if (allComplete) {
                      const hasError = toolCallResults.some((tool: any) => tool.status === 'error');
                      const finalStatus = hasError ? 'error' : 'success';

                      // 更新工具组状态 - 使用新的函数
                      const updatedMcpContent = createMcpContentWithGroupId(
                        contentItem.id, // tool_group_id
                        finalStatus,
                        toolCallResults,
                        contentItem.content.totalDuration || 0,
                      );

                      updatedContentList[index] = updatedMcpContent;
                      hasUpdates = true;
                    }
                  }
                });

                // 如果有更新，保存消息
                if (hasUpdates) {
                  const finalMessage: MessageType = {
                    id: progressToolMessage.id,
                    role: 'assistant',
                    contentList: updatedContentList,
                  };

                  addMessage(finalMessage, true);
                }
              }

              if (!requestState.current.status.isToolCallActive && (requestState.current.content.response || responseContent)) {
                const finalContent = requestState.current.content.response || responseContent;
                const aiMessage = buildMessageWithThinkContent(finalContent);
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
                    handleError(formatErrorMessage(ERROR_MESSAGES.REQUEST.SERVER_ERROR, fullResponse.error), { error: fullResponse.error }, ErrorType.REQUEST, { shouldCancel: true });
                    return;
                  }

                  // 处理正常的非流式响应
                  if (fullResponse?.data?.content) {
                    const aiMessage = buildMessageWithThinkContent(fullResponse.data.content);
                    addMessage(aiMessage);
                  } else {
                    handleError(ERROR_MESSAGES.RESPONSE.CANNOT_PARSE, null, ErrorType.PARSING, { shouldCancel: true });
                  }
                } catch (parseError) {
                  handleError(ERROR_MESSAGES.PARSING.JSON_FAILED.replace('{0}', ''), parseError, ErrorType.PARSING, { shouldCancel: true });
                }
              } catch (readError) {
                handleError(formatErrorMessage(ERROR_MESSAGES.CONNECTION.READ_FAILED, ''), readError, ErrorType.CONNECTION, { shouldCancel: true });
              }

              // 停止请求
              cancelRequest();
            },
          },
        );
      } catch (error: any) {
        handleError(formatErrorMessage(ERROR_MESSAGES.REQUEST.FAILED, error.message || '未知错误'), error, ErrorType.REQUEST);
      }
    },
    [selectedModel, addMessage, cancelRequest, clearStreamingState, currentSessionId, createStreamRequest, clearTimers, resetNoDataTimer, handleToolCalls],
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
      handleError(formatErrorMessage('重发消息失败: {0}', error.message || '未知错误'), error, ErrorType.REQUEST);
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
