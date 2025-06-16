import { useState, useCallback, useRef, useEffect } from 'react';
import { fetchEventSource, FetchEventSourceInit } from '@microsoft/fetch-event-source';
import { baseHeaders } from '@/utils';
import { API_PREFIX } from '@/constants';
import { MessageType } from '@res-utiles/ui-components';
import useChatStore from '@/components/chat-container/store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useSelectedModelStore from '@/store/useSelectedModel';

interface ChatRequestParams {
  content: string; // 消息内容，现在是字符串格式
  SessionID?: string; // 会话ID参数
  mcpIds?: string[]; // 选中的MCP IDs
}

// 响应流式数据格式
interface ChatResponseData {
  bcode: {
    business_code: number;
    message: string;
  };
  data: {
    id: string;
    session_id: string;
    content: string;
    is_complete: boolean;
    type: string;
    thinking?: string; // 可能存在的思考内容
  };
}

// 错误消息常量
const ERROR_MESSAGES = {
  NO_MODEL_SELECTED: '未选择模型，请先选择一个模型',
  NO_DATA_TIMEOUT: '请求超时：10秒未收到数据',
  TOTAL_TIMEOUT: '请求超时：30秒内未收到任何数据',
  PARSE_ERROR: '解析响应数据失败',
  CONNECTION_INTERRUPTED: '\n\n[消息传输中断]',
  EMPTY_RESPONSE: '服务器响应异常，未能获取到有效数据',
  NON_STREAMING: '服务器未返回流式数据',
};

/**
 * 对流式聊天API的封装，提供流式对话能力
 * 包含多重机制来保证在非流式或异常情况下能够停止请求
 */
export function useChatStream() {
  const { messages, addMessage, setMessages, currentSessionId, setCurrentSessionId } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const { selectedMcpIds } = useSelectMcpStore();
  // 当前正在生成的消息内容
  const [streamingContent, setStreamingContent] = useState<string>('');
  // 当前正在生成的思考内容
  const [streamingThinking, setStreamingThinking] = useState<string>('');
  // 是否正在加载
  const [isLoading, setIsLoading] = useState(false);
  // 当前消息的错误状态
  const [error, setError] = useState<string | null>(null);
  // 流处理控制器
  const abortControllerRef = useRef<AbortController | null>(null);
  // 最后一条用户消息
  const lastUserMessageRef = useRef<string | null>(null);
  // 是否正在重发消息
  const [isResending, setIsResending] = useState(false);
  // 记录是否已经尝试过重试
  const hasRetriedRef = useRef<boolean>(false);

  // 清除状态
  const clearStreamingState = useCallback(() => {
    setStreamingContent('');
    setStreamingThinking('');
    setError(null);
    hasRetriedRef.current = false;
  }, []);

  const noDataTimerRef = useRef<NodeJS.Timeout | null>(null);
  const totalTimeoutIdRef = useRef<NodeJS.Timeout | null>(null);

  // 清除定时器的函数
  const clearTimers = () => {
    if (noDataTimerRef.current) {
      clearTimeout(noDataTimerRef.current);
      noDataTimerRef.current = null;
    }
    if (totalTimeoutIdRef.current) {
      clearTimeout(totalTimeoutIdRef.current);
      totalTimeoutIdRef.current = null;
    }
  };

  // 取消当前请求
  const cancelRequest = useCallback(() => {
    // 取消网络请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    // 清除所有定时器
    clearTimers();

    // 更新加载状态
    setIsLoading(false);
  }, []);

  // 重发最后一条消息
  const resendLastMessage = useCallback(async () => {
    if (!lastUserMessageRef.current || isLoading) return;

    setIsResending(true);
    try {
      // 判断是否有选中的 MCP
      await sendChatMessageInternal(lastUserMessageRef.current, true);
    } finally {
      setIsResending(false);
    }
  }, [isLoading]);

  function generateUniqueId(prefix: string = 'id'): string {
    return `${prefix}_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
  }

  /**
   * 创建一个包含所有停止条件的流式请求
   */
  const createStreamRequest = useCallback(
    async (
      url: string,
      options: FetchEventSourceInit,
      {
        onDataReceived,
        onComplete,
        onFallbackResponse,
      }: {
        onDataReceived: (data: ChatResponseData) => void;
        onComplete: () => void;
        onFallbackResponse: (response: Response) => Promise<void>;
      },
    ) => {
      let hasReceivedData = false;
      let nonStreamingResponseCount = 0;
      const MAX_NON_STREAMING_RESPONSES = 3;

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
            await onFallbackResponse(response);
            throw new Error(ERROR_MESSAGES.NON_STREAMING);
          }
        },
        onmessage: (event) => {
          if (!event.data) {
            nonStreamingResponseCount++;
            // 如果连续收到多次空数据，可能接口不支持流式响应
            if (nonStreamingResponseCount >= MAX_NON_STREAMING_RESPONSES) {
              console.error('连续多次接收到空数据，停止请求');
              const error = new Error(ERROR_MESSAGES.EMPTY_RESPONSE);
              if (options.onerror) options.onerror(error);
              return;
            }
            return;
          }

          // 收到有效数据，重置计数器
          nonStreamingResponseCount = 0;
          hasReceivedData = true;

          try {
            let response;
            try {
              response = JSON.parse(event.data) as ChatResponseData;
            } catch (parseError: Error | any) {
              console.error('JSON解析失败，收到的数据:', event.data);
              throw new Error(`解析响应JSON失败: ${parseError.message}`);
            }
            // 检查业务状态码
            if (response.bcode.business_code !== 200) {
              throw new Error(`API错误: ${response.bcode.message}`);
            }

            onDataReceived(response);

            // 如果数据标记为完成，调用完成回调
            if (response.data.is_complete) {
              onComplete();
            }
          } catch (err: Error | any) {
            console.error('处理流数据失败:', err, event.data);
            if (options.onerror) options.onerror(err);
          }
        },
        // 添加错误处理回调
        onerror: (err) => {
          console.error('流式请求错误:', err);
          if (err.message) {
            hasRetriedRef.current = true;
            setError(`请求发送错误: ${JSON.stringify(err.message)}`);
          }
          clearTimers();
          setIsLoading(false);

          // 调用原始错误处理器（如果有）
          if (options.onerror) {
            options.onerror(err);
          }
        },

        // 添加关闭处理回调
        onclose: () => {
          // 如果未收到任何数据且未主动取消，可能是网络问题
          if (!hasReceivedData && !options.signal?.aborted) {
            console.warn('流式连接关闭但未收到任何数据');
            const closeError = new Error('连接已关闭但未收到任何数据');
            if (options.onerror) {
              options.onerror(closeError);
            }
          }

          // 调用原始关闭处理器（如果有）
          if (options.onclose) {
            options.onclose();
          }
        },
      };

      return fetchEventSource(url, enhancedOptions);
    },
    [],
  );

  // 发送聊天消息
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
        setError(ERROR_MESSAGES.NO_MODEL_SELECTED);
        return;
      }

      // 记录最后一条消息用于重发
      if (!isResend) {
        lastUserMessageRef.current = userMessage;
      }

      // 创建用户消息，仅在非重发时添加
      const userMsg: MessageType = {
        id: Date.now().toString(),
        role: 'user',
        contentList: [
          {
            id: generateUniqueId('content'),
            type: 'plain',
            content: userMessage,
          },
        ],
      };

      // 如果不是重发，添加用户消息到消息列表
      if (!isResend) {
        addMessage(userMsg);
      }

      // 设置加载状态
      setIsLoading(true);
      // 创建请求控制器
      abortControllerRef.current = new AbortController();
      const signal = abortControllerRef.current.signal;

      const requestData: ChatRequestParams = {
        content: userMessage,
        SessionID: currentSessionId,
      };

      if (selectedMcpIds.length > 0) {
        requestData['mcpIds'] = selectedMcpIds();
      }
      const NO_DATA_TIMEOUT = 10000;
      const TOTAL_TIMEOUT = 30000;
      console.log('====>,', requestData);
      // return;
      const resetNoDataTimer = () => {
        if (noDataTimerRef.current) clearTimeout(noDataTimerRef.current);
        noDataTimerRef.current = setTimeout(() => {
          // 首次超时尝试重试一次
          if (!hasRetriedRef.current && !isResend) {
            console.log('数据超时，尝试重新发送请求');
            hasRetriedRef.current = true;
            // 取消当前请求
            cancelRequest();
            // 延迟200ms后重新发送
            setTimeout(() => {
              sendChatMessageInternal(userMessage, true);
            }, 200);
          } else {
            setError(ERROR_MESSAGES.NO_DATA_TIMEOUT);
            cancelRequest();
          }
        }, NO_DATA_TIMEOUT);
      };

      try {
        const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';

        // 设置总超时
        totalTimeoutIdRef.current = setTimeout(() => {
          setError(ERROR_MESSAGES.TOTAL_TIMEOUT);
          cancelRequest();
        }, TOTAL_TIMEOUT);

        // 在收到第一个数据前启动无数据超时计时器
        resetNoDataTimer();

        let responseContent = '';
        let thinkingContent = '';

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
              // 如果连接关闭但未完成，可能是网络问题
              if (isLoading && responseContent) {
                // 创建部分AI回复消息
                const aiMessage: MessageType = {
                  id: (Date.now() + 1).toString(),
                  role: 'assistant',
                  contentList: [
                    {
                      id: generateUniqueId('content'),
                      type: 'plain',
                      content: responseContent + ERROR_MESSAGES.CONNECTION_INTERRUPTED,
                    },
                    ...(thinkingContent
                      ? [
                          {
                            id: generateUniqueId('content'),
                            type: 'think',
                            content: thinkingContent,
                          },
                        ]
                      : []),
                  ],
                };
                addMessage(aiMessage);
              }
              setIsLoading(false);
            },
          },
          {
            // 数据处理回调
            onDataReceived: (response) => {
              // 重置无数据超时计时器
              resetNoDataTimer();
              // 清除总超时
              if (totalTimeoutIdRef.current) {
                clearTimeout(totalTimeoutIdRef.current);
                totalTimeoutIdRef.current = null;
              }

              const data = response.data;

              // 保存会话ID
              if (data.session_id && (!currentSessionId || currentSessionId !== data.session_id)) {
                setCurrentSessionId(data.session_id);
              }

              // 处理思考内容 (如果有提供)
              if (data.thinking) {
                thinkingContent += data.thinking;
                setStreamingThinking(thinkingContent);
              }

              // 处理正常内容
              if (data.content) {
                if (data.is_complete) {
                  responseContent = data.content;
                  setStreamingContent(responseContent);
                  return;
                }

                // 根据类型处理增量内容
                if (data.type === 'delta') {
                  // delta 类型表示增量
                  responseContent += data.content;
                } else if (data.type === 'answer') {
                  // answer 类型往往是服务器的标准响应格式
                  if (responseContent.length === 0) {
                    // 如果是第一个响应，直接赋值
                    responseContent = data.content;
                  } else {
                    // 否则考虑为增量
                    responseContent += data.content;
                  }
                } else if (data.type === 'replace' || responseContent.length === 0) {
                  responseContent = data.content;
                } else {
                  // 对于未知类型，使用启发式判断
                  if (data.content.length > responseContent.length || !responseContent.includes(data.content.trim())) {
                    responseContent += data.content;
                  } else {
                    responseContent = data.content;
                  }
                }
                setStreamingContent(responseContent);
              }
            },
            // 完成回调
            onComplete: () => {
              // 创建最终的AI回复消息
              const aiMessage: MessageType = {
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                contentList: [
                  {
                    id: '1',
                    type: 'plain',
                    content: responseContent,
                  },
                  // 添加思考内容（如果有）
                  ...(thinkingContent
                    ? [
                        {
                          id: '2',
                          type: 'think',
                          content: thinkingContent,
                        },
                      ]
                    : []),
                ],
              };

              // 添加到消息列表
              addMessage(aiMessage);
              clearStreamingState();
              setIsLoading(false);
              clearTimers();
            },
            // 非流式响应处理
            onFallbackResponse: async (response) => {
              try {
                const fullResponseText = await response.text();
                console.log('接收到非流式响应:', fullResponseText);

                // 尝试解析为 JSON
                try {
                  const fullResponse = JSON.parse(fullResponseText);

                  // 检查是否包含错误信息
                  if (fullResponse.error) {
                    console.error('接收到错误响应:', fullResponse.error);
                    // 阻止重试
                    hasRetriedRef.current = true;
                    setError(`服务器返回错误: ${fullResponse.error}`);
                    cancelRequest();
                    return;
                  }

                  // 处理正常的非流式响应
                  if (fullResponse?.data?.content) {
                    const aiMessage: MessageType = {
                      id: generateUniqueId('msg'),
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
              cancelRequest();
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
    [selectedModel, addMessage, cancelRequest, clearStreamingState, currentSessionId, setCurrentSessionId, createStreamRequest],
  );

  // 对外暴露的发送消息函数
  const sendChatMessage = useCallback(
    (userMessage: string) => {
      if (isLoading) return; // 正在加载时不允许发送
      sendChatMessageInternal(userMessage);
    },
    [sendChatMessageInternal, isLoading],
  );

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

  // 组件卸载时清理
  useEffect(() => {
    return () => {
      cancelRequest();
    };
  }, [cancelRequest]);

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
