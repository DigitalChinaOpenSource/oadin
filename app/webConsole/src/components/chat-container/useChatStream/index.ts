import { useState, useCallback, useRef, useEffect } from 'react';
import { fetchEventSource, FetchEventSourceInit } from '@microsoft/fetch-event-source';
import { baseHeaders } from '@/utils';
import { API_PREFIX } from '@/constants';
import { ChatMessageItem } from '@res-utiles/ui-components';
import useChatStore from '@/components/chat-container/store/useChatStore';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useSelectedModelStore from '@/store/useSelectedModel';
import { getIdByFunction } from '../..//select-mcp/lib/useSelectMcpHelper';
import { httpRequest } from '@/utils/httpRequest';
import { generateUniqueId, formatErrorMessage, fetchGenChatTitle } from './utils';
import {
  findProgressToolMessage,
  extractToolCallDataByGroupId,
  createMcpContentWithGroupId,
  updateContentListWithMcpByGroupId,
  buildToolCallData,
  handleToolCallErrorMessage,
} from './mcpContentUtils';
import { buildMessageWithThinkContent, handleTextContent } from './thinkContentUtils';
import { IStreamData, StreamCallbacks, ChatRequestParams, ChatResponseData, IToolCallData } from './types';
import { ERROR_MESSAGES, TIMEOUT_CONFIG, ErrorType } from './contants';
import { getSessionIdFromUrl } from '@/utils/sessionParamUtils';
import { message } from 'antd';
import useUploadFileListStore from '@/components/chat-container/store/useUploadFileListStore.ts';

// 统一的流式请求参数接口
interface StreamRequestOptions {
  content: string;
  attachmentFiles?: []; // 附件文件列表名称
  toolGroupID?: string;
  isUserMessage?: boolean; // 是否需要创建用户消息
  isResend?: boolean;
  sessionId?: string;
}

export function useChatStream() {
  const { addMessage, isLoading, setIsLoading } = useChatStore();
  const migratingStatus = useModelPathChangeStore.getState().migratingStatus;
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();

  const [streamingContent, setStreamingContent] = useState<string>('');
  const [streamingThinking, setStreamingThinking] = useState<string | { data: string; status: string }>('');
  const [isResending, setIsResending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 流式请求控制器
  const abortControllerRef = useRef<AbortController | null>(null);

  // 请求状态统一处理
  const requestState = useRef({
    content: {
      response: '',
      thinking: '',
      thinkingFromField: '',
      isThinkingActive: false,
      thinkingTotalDuration: undefined as number | undefined,
    },
    status: {
      isToolCallActive: false,
    },
    timers: {
      totalTimer: null as NodeJS.Timeout | null,
    },
    lastToolGroupIdRef: null as string | null,
    lastDataPacket: null as any, // 保存最后一个数据包
  });
  const toolCallHandlersRef = useRef<{
    continueConversation: ((result: string) => Promise<void>) | null;
  }>({ continueConversation: null });
  const functionIdCacheRef = useRef<Record<string, any>>({});

  /**
   * 清除所有定时器
   */
  const clearTimers = useCallback(() => {
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
      toolGroupId?: string;
    } = {},
  ) => {
    console.log('handleError', message, originalError, errorType, options);
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
          const errorMessage: ChatMessageItem = {
            id: progressToolMessage.id,
            role: 'assistant',
            contentList: updatedContentList,
          };
          console.log('errorMessage', errorMessage);
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
        thinkingFromField: '',
        isThinkingActive: false,
        thinkingTotalDuration: undefined,
      },
      status: {
        isToolCallActive: false,
      },
      timers: {
        totalTimer: null,
      },
      lastToolGroupIdRef: null,
      lastDataPacket: null,
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
    if (requestState.current.timers.totalTimer) {
      clearTimeout(requestState.current.timers.totalTimer);
      requestState.current.timers.totalTimer = null;
    }
  }, []);

  // 取消请求并保存已生成内容
  const cancelRequest = () => {
    const { response, thinking, thinkingFromField, isThinkingActive } = requestState.current.content;
    // 如果有生成的内容，保存为消息
    const tempStreamingThinking = typeof streamingThinking === 'string' ? streamingThinking : streamingThinking.data;
    if (isLoading && (response || streamingContent || tempStreamingThinking || thinkingFromField)) {
      const finalContent = (response || streamingContent || '') + ERROR_MESSAGES.CONNECTION.RESPONSE_INTERRUPTED;
      // 如果正在进行深度思考，标记为用户取消（错误状态）
      if (isThinkingActive || thinkingFromField) {
        setStreamingThinking({
          data: thinkingFromField || '',
          status: 'error',
        });
      }
      // 将 thinking 内容也传入构建消息的函数

      const aiMessage = buildMessageWithThinkContent(finalContent, false, thinking, thinkingFromField, false);
      addMessage(aiMessage);

      // 保存当前的toolGroupId，但清除内部状态
      const savedToolGroupId = requestState.current.lastToolGroupIdRef;
      requestState.current = {
        content: {
          response: '',
          thinking: '',
          thinkingFromField: '',
          isThinkingActive: false,
          thinkingTotalDuration: undefined,
        },
        status: {
          isToolCallActive: false,
        },
        timers: {
          totalTimer: null,
        },
        lastToolGroupIdRef: savedToolGroupId,
        lastDataPacket: null,
      };
      functionIdCacheRef.current = {};

      setIsLoading(false);
      cleanupResources();
    } else {
      cleanupResources();
      clearStreamingState();
      setIsLoading(false);
    }
  };

  // 创建流式请求
  const createStreamRequest = useCallback(async (url: string, options: FetchEventSourceInit, callbacks: StreamCallbacks) => {
    const enhancedOptions: FetchEventSourceInit = {
      ...options,
      async onopen(response) {
        // 检查响应状态
        if (!response.ok) {
          const errorText = await response.text().catch(() => 'Unknown error');
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
        console.log('onerror314', err);
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

      // 检查是否有累积的内容需要先保存（包括思考内容和普通内容）
      const { thinking, thinkingFromField, response } = requestState.current.content;
      const hasAccumulatedThinking = (thinking && thinking.trim() !== '') || (thinkingFromField && thinkingFromField.trim() !== '');
      const hasAccumulatedContent = (currentContent && currentContent.trim() !== '') || (response && response.trim() !== '');

      // 如果有累积的思考内容或普通内容，需要在工具调用前保存为一条消息
      if (hasAccumulatedThinking || hasAccumulatedContent) {
        const contentToSave = currentContent || response || '';

        // 构建包含思考和普通内容的消息
        const accumulatedMessage = buildMessageWithThinkContent(
          contentToSave,
          true, // 标记为完成
          thinking || '', // 来自 <think> 标签的思考内容
          thinkingFromField || '', // 来自 thoughts 字段的思考内容
          false, // 已经完成，不再活跃
          data.total_duration,
        );

        // 添加累积内容消息
        addMessage(accumulatedMessage);

        // 清除累积的内容，避免重复显示
        requestState.current.content.thinking = '';
        requestState.current.content.thinkingFromField = '';
        requestState.current.content.response = '';
        requestState.current.content.isThinkingActive = false;

        // 清除流式显示状态
        setStreamingContent('');
        setStreamingThinking('');
      }

      // 修改：只要有 tool_calls 和 tool_group_id，就保存用于下一次请求
      if (tool_calls && tool_calls.length > 0 && tool_group_id) {
        requestState.current.lastToolGroupIdRef = tool_group_id;
      }

      if (!tool_calls || tool_calls.length === 0) {
        requestState.current.lastToolGroupIdRef = null;
        return;
      }

      // 初始化状态变量
      let toolCallResults: any[] = [];
      let currentContentList: any[] = [];
      let currentTotalDuration = 0;
      try {
        // 设置工具调用状态
        setStreamingContent(currentContent);
        const _function = tool_calls?.[0].function;
        // 获取工具响应
        const toolResponse = await getIdByFunction({ toolName: _function.name, toolArgs: _function.arguments }, selectedMcpIds());

        // 查找已存在的消息
        const messages = useChatStore.getState().messages;
        // 查找进行中的工具调用消息
        const progressToolMessage = findProgressToolMessage(messages);
        if (progressToolMessage) {
          currentContentList = [...(progressToolMessage.contentList || [])];

          if (tool_group_id) {
            const { toolCallResults: existingResults, mcpContentIndex } = extractToolCallDataByGroupId(currentContentList, tool_group_id);

            if (mcpContentIndex >= 0) {
              // 如果找到现有的 tool_group_id，使用其数据
              toolCallResults = existingResults;
              currentTotalDuration = currentContentList[mcpContentIndex]?.content?.totalDuration || 0;
            }
          }
        }
        // 累加当前的调用时间
        if (total_duration) {
          currentTotalDuration += total_duration;
        }

        // 添加当前工具调用到结果数组
        const mcpId = toolResponse.mcpId || generateUniqueId('mcp');
        toolCallResults.push(buildToolCallData(toolResponse, mcpId, 'progress'));

        // 构建并更新消息
        let mcpContent;
        if (tool_group_id) {
          mcpContent = createMcpContentWithGroupId(tool_group_id, 'progress', toolCallResults, currentTotalDuration);
          currentContentList = updateContentListWithMcpByGroupId(currentContentList, mcpContent, tool_group_id);
        }

        const mcpMessage: ChatMessageItem = {
          id: tool_group_id || id,
          role: 'assistant',
          contentList: currentContentList,
        };

        addMessage(mcpMessage, !!(tool_group_id || id));

        // 执行工具调用
        const data = await httpRequest.post('/mcp/client/runTool', { messageId: id, ...toolResponse });
        const isToolError = data?.isError === true;
        const isEmptyResponse = data?.isError === undefined || data?.isError === null;
        const toolErrorMessage = isToolError && Array.isArray(data.content) && data.content.length > 0 ? data.content[0]?.text || ERROR_MESSAGES.TOOL.EXECUTION_FAILED : '';

        setStreamingContent(currentContent);
        requestState.current.content.response = currentContent;

        // 更新工具调用结果，仅更新单个工具的状态，整体状态保持为progress
        const lastIndex = toolCallResults.length - 1;
        toolCallResults[lastIndex] = {
          ...toolCallResults[lastIndex],
          outputParams: isToolError ? toolErrorMessage : data.content[0]?.text || '',
          logo: data.logo,
          desc: data.toolDesc,
          status: isEmptyResponse ? 'empty' : isToolError ? 'error' : 'success',
        } as IToolCallData;

        // 两个层次的状态：
        // 1. 单个工具状态（上面已更新）：由 /runTool 结果决定 ('success'/'error')
        // 2. MCP 组状态：根据工具调用结果决定
        //    - 如果工具调用出错，整个组状态立即变为 'error'
        //    - 如果工具调用成功但还有后续工具调用，保持 'progress'
        //    - 如果工具调用成功且没有后续工具调用，在 onComplete 中设置最终状态
        const mcpGroupStatus = isToolError || isEmptyResponse ? 'error' : 'progress';

        // 更新最终消息
        let updatedMcpContent;
        let updatedContentList;
        if (tool_group_id) {
          updatedMcpContent = createMcpContentWithGroupId(tool_group_id, mcpGroupStatus, toolCallResults, currentTotalDuration);
          updatedContentList = updateContentListWithMcpByGroupId(currentContentList, updatedMcpContent, tool_group_id);
        }

        const updatedMessage: ChatMessageItem = {
          id: tool_group_id || id,
          role: 'assistant',
          contentList: updatedContentList,
        };
        addMessage(updatedMessage, true);

        // 处理后续操作
        if (!isToolError && toolCallHandlersRef.current.continueConversation) {
          // 调用继续对话函数，处理后续流式响应
          await toolCallHandlersRef.current.continueConversation(data.content[0].text);
          // } else if (isToolError || isEmptyResponse) {
          //   const errorContent = currentContent + `\n\n[工具调用失败: ${toolErrorMessage}]`;
          //   handleTextContent({ content: errorContent, is_complete: true }, '', setStreamingContent, setStreamingThinking, requestState, false);
          //   requestState.current.content.response = errorContent;
        } else {
          console.error('工具调用处理程序未初始化');
          setError('工具调用处理失败: 内部错误');
        }
      } catch (error: any) {
        console.log('error496', error);
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
      }
    },
    [selectedMcpIds, setStreamingContent, setStreamingThinking, setError, addMessage],
  );

  // 重构后的统一流式请求函数
  const sendStreamRequest = async (options: StreamRequestOptions) => {
    const { content, toolGroupID, isUserMessage = false, isResend = false } = options;

    if (!toolGroupID && !content.trim()) return;
    if (migratingStatus === 'pending') {
      message.warning('模型迁移中，请稍后再试');
      return;
    }
    // 取消之前的请求
    cancelRequest();
    // 清除流状态
    clearStreamingState();

    // 模型检查
    if (!selectedModel) {
      handleError(ERROR_MESSAGES.REQUEST.NO_MODEL_SELECTED, null, ErrorType.REQUEST, { shouldCancel: false });
      return;
    }

    // 创建用户消息（仅在需要时）
    if (isUserMessage) {
      const uploadFileList = useUploadFileListStore.getState().uploadFileList || [];
      const userMsg: ChatMessageItem = {
        id: generateUniqueId('user_msg'),
        role: 'user',
        contentList: [
          {
            id: generateUniqueId('content'),
            type: 'plain',
            content: content,
            attachmentFiles: uploadFileList.length > 0 ? uploadFileList : undefined,
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
      content: toolGroupID ? '' : content, // 如果有 toolGroupID，content 设为空
      SessionID: getSessionIdFromUrl(),
    };

    // 添加工具组ID
    if (toolGroupID) {
      requestData.toolGroupID = toolGroupID;
    }

    // 添加MCP IDs
    if (selectedMcpIds().length > 0) {
      requestData.mcpIds = selectedMcpIds();
    }

    try {
      const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';

      // 设置总超时
      requestState.current.timers.totalTimer = setTimeout(() => {
        handleError(ERROR_MESSAGES.TIMEOUT.TOTAL, null, ErrorType.TIMEOUT, {
          shouldCancel: true,
          appendToContent: true,
          updateMcpStatus: true,
        });
      }, TIMEOUT_CONFIG.TOTAL);

      // 初始化响应内容
      let responseContent = requestState.current.content.response;

      await createStreamRequest(
        `${API_BASE_URL}${API_PREFIX}/playground/message/stream`,
        {
          method: 'POST',
          headers: baseHeaders(),
          body: JSON.stringify(requestData),
          openWhenHidden: true,
          signal,
          onerror: (error) => {
            console.log('error599', error);
            setError(`请求失败: ${error.message}`);
            clearTimers();
            setIsLoading(false);
          },
          onclose: () => {
            clearTimers();
            const { response, thinking, thinkingFromField, isThinkingActive } = requestState.current.content;
            const tempStreamingThinking = typeof streamingThinking === 'string' ? streamingThinking : streamingThinking.data;
            // 如果连接关闭但未完成且有内容，保存部分回复
            if (isLoading && (response || responseContent || thinkingFromField)) {
              const finalContent = (response || responseContent || '') + ERROR_MESSAGES.CONNECTION.INTERRUPTED;
              const aiMessage = buildMessageWithThinkContent(finalContent, false, thinking || tempStreamingThinking, thinkingFromField, isThinkingActive);
              addMessage(aiMessage);
            }

            if (!requestState.current.status.isToolCallActive) {
              setIsLoading(false);
            }
          },
        },
        {
          onDataReceived: async (response) => {
            // 清除总超时定时器
            if (requestState.current.timers.totalTimer) {
              clearTimeout(requestState.current.timers.totalTimer);
              requestState.current.timers.totalTimer = null;
            }

            const data = response.data;
            // 保存最后一个数据包，用于在 onComplete 中检查
            requestState.current.lastDataPacket = data;

            // 根据是否有 tool_calls 来设置 isToolCallActive
            if (data?.tool_calls && data.tool_calls.length > 0) {
              requestState.current.status.isToolCallActive = true;
            }
            // 注意：这里不再在没有 tool_calls 时立即设置 isToolCallActive = false
            // 因为工具调用是异步的，需要等待完整的工具调用流程结束

            // 保存 total_duration 到 requestState
            if (data.total_duration !== undefined) {
              requestState.current.content.thinkingTotalDuration = data.total_duration;
            }

            if (data?.tool_calls && data.tool_calls.length > 0) {
              // 处理工具调用（在 handleToolCalls 中已经处理了内容保存）
              await handleToolCalls(data, responseContent);
            } else if (data.content || data.thoughts) {
              // 处理普通文本内容
              responseContent = handleTextContent(data, responseContent, setStreamingContent, setStreamingThinking, requestState, false);
              requestState.current.content.response = responseContent;
            }
          },

          onComplete: () => {
            // 检查最后一个数据包是否包含 tool_calls
            const lastDataPacket = requestState.current.lastDataPacket;
            const hasToolCallsInLastPacket = lastDataPacket?.tool_calls && lastDataPacket.tool_calls.length > 0;

            // 如果最后一个数据包包含 tool_calls，表示 MCP 工具调用还在进行中
            // 此时应该保持 isLoading 和 isToolCallActive 为 true
            if (hasToolCallsInLastPacket) {
              // 确保工具调用状态保持为 true
              requestState.current.status.isToolCallActive = true;
              // 不设置 isLoading = false，让工具调用流程继续
              clearTimers();
              return;
            }

            // 只有当最后一个数据包不包含 tool_calls 时，才处理完成逻辑
            requestState.current.status.isToolCallActive = false;

            // 处理进行中的工具调用完成状态
            const allMessages = useChatStore.getState().messages;
            const progressToolMessage = findProgressToolMessage(allMessages) as any;

            if (progressToolMessage && !requestState.current.status.isToolCallActive) {
              const updatedContentList = [...progressToolMessage.contentList];
              let hasUpdates = false;

              updatedContentList.forEach((contentItem, index) => {
                if (contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress') {
                  const toolCallResults = contentItem.content.data || [];
                  const allComplete = toolCallResults.every((tool: any) => tool.status === 'success' || tool.status === 'error');

                  if (allComplete) {
                    const hasError = toolCallResults.some((tool: any) => tool.status === 'error');
                    const finalStatus = hasError ? 'error' : 'success';
                    const currentGroupId = contentItem.content.groupId || requestState.current.lastToolGroupIdRef;

                    if (currentGroupId) {
                      const updatedMcpContent = createMcpContentWithGroupId(currentGroupId, finalStatus, toolCallResults, contentItem.content.totalDuration || 0);
                      updatedContentList[index] = updatedMcpContent;
                      hasUpdates = true;
                    }
                  }
                }
              });

              if (hasUpdates) {
                const finalMessage: ChatMessageItem = {
                  id: progressToolMessage.id,
                  role: 'assistant',
                  contentList: updatedContentList,
                };
                addMessage(finalMessage, true);
              }
            }

            const { response, thinking, thinkingFromField, isThinkingActive } = requestState.current.content;
            const tempStreamingThinking = typeof streamingThinking === 'string' ? streamingThinking : streamingThinking.data;
            // 保存最终消息（仅在没有工具调用活动时）
            if (!requestState.current.status.isToolCallActive && (response || responseContent || thinkingFromField)) {
              const finalContent = response || responseContent || '';
              const aiMessage = buildMessageWithThinkContent(
                finalContent,
                true,
                thinking || tempStreamingThinking,
                thinkingFromField,
                isThinkingActive,
                requestState.current.content.thinkingTotalDuration,
              );
              addMessage(aiMessage);

              requestState.current = {
                content: {
                  response: '',
                  thinking: '',
                  thinkingFromField: '',
                  isThinkingActive: false,
                  thinkingTotalDuration: undefined,
                },
                status: {
                  isToolCallActive: false,
                },
                timers: {
                  totalTimer: null,
                },
                lastToolGroupIdRef: requestState.current.lastToolGroupIdRef,
                lastDataPacket: null,
              };
              functionIdCacheRef.current = {};
            } else {
              const savedToolGroupId = requestState.current.lastToolGroupIdRef;
              clearStreamingState();
              requestState.current.lastToolGroupIdRef = savedToolGroupId;
            }

            if (!requestState.current.status.isToolCallActive) {
              setIsLoading(false);
              // 表示会话已完全结束，检查是否需要生成标题
              const isGenChatTitle = sessionStorage.getItem('isGenChatTitle');
              if (isGenChatTitle !== 'false') {
                fetchGenChatTitle().finally(() => {
                  sessionStorage.setItem('isGenChatTitle', 'false');
                });
              }
            }

            clearTimers();
          },

          onFallbackResponse: async (response) => {
            handleError(ERROR_MESSAGES.RESPONSE.NON_STREAMING, null, ErrorType.PARSING, { shouldCancel: true, appendToContent: false });
            cancelRequest();
          },
        },
      );
    } catch (error: any) {
      console.log('error770', error);
      handleError(formatErrorMessage(ERROR_MESSAGES.REQUEST.FAILED, error.message || '未知错误'), error, ErrorType.REQUEST);
    }
  };

  // 简化后的用户消息发送函数
  const sendChatMessage = (userMessage: string) => {
    if (isLoading) return;

    sendStreamRequest({
      content: userMessage,
      isUserMessage: true,
    });
  };

  // 简化后的工具结果继续对话函数
  const continueConversationWithResult = async (toolResult: string) => {
    try {
      await sendStreamRequest({
        content: toolResult,
        toolGroupID: requestState.current.lastToolGroupIdRef || undefined,
        isUserMessage: false,
      });
    } catch (error: any) {
      handleError(formatErrorMessage(ERROR_MESSAGES.TOOL.CONTINUE_FAILED, error.message || '未知错误'), error, ErrorType.TOOL);
    }
  };

  // 重发消息
  const resendLastMessage = async () => {
    if (isResending) return;

    try {
      setIsResending(true);
      setError(null);

      const messages = useChatStore.getState().messages;
      const lastUserMessage = [...messages].reverse().find((msg) => msg.role === 'user');

      if (!lastUserMessage || !lastUserMessage.contentList || lastUserMessage.contentList.length === 0) {
        console.warn('没有找到可重发的用户消息');
        setIsResending(false);
        return;
      }

      const lastContent = lastUserMessage.contentList[lastUserMessage.contentList.length - 1];
      if (!lastContent || !lastContent.content || typeof lastContent.content !== 'string') {
        console.warn('用户消息内容格式不正确');
        setIsResending(false);
        return;
      }

      await sendStreamRequest({
        content: lastContent.content,
        isUserMessage: true,
        isResend: true,
      });
    } catch (error: any) {
      handleError(formatErrorMessage('重发消息失败: {0}', error.message || '未知错误'), error, ErrorType.REQUEST);
    } finally {
      setIsResending(false);
    }
  };

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
    setStreamingContent,
    streamingThinking,
    isLoading,
    isResending,
    error,
    sendChatMessage,
    cancelRequest,
    resendLastMessage,
  };
}
