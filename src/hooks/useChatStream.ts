import { useState, useCallback, useRef, useEffect } from 'react';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { baseHeaders } from '@/utils';
import { API_PREFIX } from '@/constants';
import { MessageType } from '@res-utiles/ui-components';
import useChatStore from '@/components/chat-container/store/useChatStore';
import useSelectedModelStore from '@/store/useSelectedModel';

// 聊天请求参数
interface ChatRequestParams {
  model: string;
  messages: Array<{
    role: string;
    content: string;
    thinking?: string;
    images?: string[];
    tool_calls?: any[];
  }>;
  tools?: any[];
  think?: boolean;
}

// 聊天响应数据
interface ChatResponseData {
  created_at: string;
  finished: boolean;
  id: string;
  message: {
    content: string;
    role: string;
    thinking?: string;
  };
  model: string;
  finish_reason?: string;
}

export function useChatStream() {
  const { messages, addMessage } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
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
  const responseIdRef = useRef<string | null>(null);

  // 清除状态
  const clearStreamingState = useCallback(() => {
    setStreamingContent('');
    setStreamingThinking('');
    setError(null);
  }, []);

  // 取消当前请求
  const cancelRequest = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      setIsLoading(false);
    }
  }, []);

  // 将消息转换为API请求格式
  const convertMessagesToApiFormat = useCallback((messages: MessageType[]) => {
    return messages.map((msg) => {
      // 提取消息内容
      const content = msg.contentList?.find((item) => item.type === 'plain')?.content || '';
      // 提取思考内容
      const thinking = msg.contentList?.find((item) => item.type === 'think')?.content || '';
      // TODO 增加 MCP 工具
      return {
        role: msg.role,
        content,
        ...(thinking ? { thinking } : {}),
      };
    });
  }, []);

  // 发送聊天消息并获取流式响应
  const sendChatMessage = useCallback(
    async (userMessage: string) => {
      if (!userMessage.trim()) return;

      // 取消之前的请求
      cancelRequest();
      // 清除流状态
      clearStreamingState();
      // 模型检查
      if (!selectedModel) {
        setError('未选择模型，请先选择一个模型');
        return;
      }
      // 创建用户消息
      const userMsg: MessageType = {
        id: Date.now().toString(),
        role: 'user',
        contentList: [
          {
            id: '1',
            type: 'plain',
            content: userMessage,
          },
        ],
      };
      // 添加到消息列表
      addMessage(userMsg);
      // 设置加载状态
      setIsLoading(true);
      // 创建请求控制器
      abortControllerRef.current = new AbortController();
      const signal = abortControllerRef.current.signal;
      // 准备历史消息
      const historyMessages = convertMessagesToApiFormat([...messages, userMsg]);

      // 请求参数
      const requestData: ChatRequestParams = {
        model: selectedModel.name,
        // TODO
        messages: historyMessages as any,
        think: true, // 启用思考模式
      };

      // 设置超时处理
      let noDataTimer: NodeJS.Timeout | null = null;
      let totalTimeoutId: NodeJS.Timeout | null = null;
      const NO_DATA_TIMEOUT = 10000;
      const TOTAL_TIMEOUT = 30000;

      const resetNoDataTimer = () => {
        if (noDataTimer) clearTimeout(noDataTimer);
        noDataTimer = setTimeout(() => {
          setError('请求超时：10秒未收到数据');
          cancelRequest();
        }, NO_DATA_TIMEOUT);
      };

      const clearTimers = () => {
        if (noDataTimer) clearTimeout(noDataTimer);
        if (totalTimeoutId) clearTimeout(totalTimeoutId);
      };

      try {
        const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';

        totalTimeoutId = setTimeout(() => {
          setError('请求超时：30秒内未收到任何数据');
          cancelRequest();
        }, TOTAL_TIMEOUT);

        let isCollectingThinking = false;
        let responseContent = '';
        let thinkingContent = '';

        await fetchEventSource(`${API_BASE_URL}${API_PREFIX}/services/chat`, {
          method: 'POST',
          headers: baseHeaders(),
          body: JSON.stringify(requestData),
          openWhenHidden: true,
          signal,
          onmessage: (event) => {
            if (event.data) {
              try {
                resetNoDataTimer();

                if (totalTimeoutId) {
                  clearTimeout(totalTimeoutId);
                  totalTimeoutId = null;
                }

                const data = JSON.parse(event.data) as ChatResponseData;
                // 保存响应ID
                if (!responseIdRef.current && data.id) {
                  responseIdRef.current = data.id;
                }
                // 处理思考内容
                if (data.message?.thinking) {
                  thinkingContent += data.message.thinking;
                  setStreamingThinking(thinkingContent);
                  isCollectingThinking = true;
                }
                // 处理正常内容
                if (data.message?.content) {
                  responseContent += data.message.content;
                  setStreamingContent(responseContent);
                }
                // 处理完成状态
                if (data.finished) {
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
                  responseIdRef.current = null;
                  clearTimers();
                }
              } catch (err) {
                console.error('解析流数据失败:', err);
                setError('解析响应数据失败');
              }
            }
          },
          onerror: (error) => {
            console.error('流请求错误:', error);
            setError(`请求失败: ${error.message}`);
            clearTimers();
            setIsLoading(false);
          },
          onclose: () => {
            console.log('流连接关闭');
            clearTimers();
            // 如果连接关闭但未完成，可能是网络问题
            if (isLoading && responseContent) {
              // 创建部分AI回复消息
              const aiMessage: MessageType = {
                id: (Date.now() + 1).toString(),
                role: 'assistant',
                contentList: [
                  {
                    id: '1',
                    type: 'plain',
                    content: responseContent + '\n\n[消息传输中断]',
                  },
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

              addMessage(aiMessage);
            }
            setIsLoading(false);
          },
        });
      } catch (error: any) {
        console.error('发送聊天请求失败:', error);
        setError(`发送请求失败: ${error.message}`);
        setIsLoading(false);
        clearTimers();
      }

      // 返回取消方法
      return {
        cancelRequest,
      };
    },
    [messages, selectedModel, addMessage, cancelRequest, clearStreamingState, convertMessagesToApiFormat],
  );

  useEffect(() => {
    return () => {
      cancelRequest();
    };
  }, [cancelRequest]);

  return {
    streamingContent,
    streamingThinking,
    isLoading,
    error,
    sendChatMessage,
    cancelRequest,
  };
}
