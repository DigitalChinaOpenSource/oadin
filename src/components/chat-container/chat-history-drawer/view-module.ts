import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IChatDetailItem, IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';
import { chatHistoryData } from './mock-data.ts';
import useChatStore from '@/components/chat-container/store/useChatStore.ts';
import { message } from 'antd';
import { IModelDataItem } from '@/types';
import useSelectedModelStore from '@/store/useSelectedModel.ts';
import { MessageType } from '@res-utiles/ui-components';

export function useChatHistoryDrawer() {
  // 获取对话store
  const { setHistoryVisible, createNewChat, setCurrentSessionId, setMessages } = useChatStore();
  const currentSessionId = useChatStore((state) => state.currentSessionId);

  // 获取模型store
  const { setSelectedModel, setIsSelectedModel } = useSelectedModelStore();
  // 历史对话记录
  const [chatHistory, setChatHistory] = useState<IChatHistoryItem[]>([]);

  // 用于记录当前显示 Popconfirm 的卡片 id 并设置是否显示确认弹窗
  const [showDeleteId, setShowDeleteId] = useState<string | null>(null);

  // 获取历史对话记录
  const { loading: historyLoading, run: fetchChatHistory } = useRequest(
    async () => {
      const data = await httpRequest.get<IChatHistoryItem[]>('/playground/sessions');
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        if (!data) return;
        setChatHistory(data);
      },
      onError: (error) => {
        console.error('获取历史对话记录失败:', error);
      },
    },
  );

  // 删除对话历史记录
  const {
    params,
    loading: delHistoryLoading,
    run: deleteChatHistory,
  } = useRequest(
    async (id: string | number) => {
      const data = await httpRequest.del('/playground/session', { sessionId: id });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        // 删除成功 刷新历史记录(刷新会导致重新加载，现使用前台过滤)
        setChatHistory((prev) => prev.filter((item) => item.id !== params[0]));
        message.success('删除对话历史记录成功');
        // 如果删除的是当前会话的历史记录，则清空当前会话(能否新建一个对话)
        if (params[0] === currentSessionId) {
          createNewChat();
        }
        if (!data) return;
      },
      onError: (error) => {
        console.error('获取更新日志失败:', error);
      },
      onFinally: (data: any) => {
        setShowDeleteId(null);
      },
    },
  );

  // 获取历史对话详情
  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        if (!data || !data.length) return;

        fetchModelDetail(data[data.length - 1].modelId, data, data[data.length - 1].sessionId);
      },
      onError: (error) => {
        console.error('获取历史对话记录失败:', error);
        message.error('获取历史对话记录失败');
      },
    },
  );

  // 根据模型id获取模型详情
  const { run: fetchModelDetail } = useRequest(
    async (modelId: string, messages: MessageType[], sessionId: string) => {
      const data = await httpRequest.get<IModelDataItem>(`control_panel/model/square?modelId=${modelId}`);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any, params) => {
        if (!data) return;
        setSelectedModel(data);
        setIsSelectedModel(true);
        setCurrentSessionId(params[2]);
        setMessages(params[1]);
        setHistoryVisible(false);
      },
      onError: (error) => {
        console.error('获取历史对话记录失败:', error);
      },
    },
  );

  useEffect(() => {
    fetchChatHistory();
  }, []);

  return { historyLoading, fetchChatHistory, chatHistory, delHistoryLoading, deleteChatHistory, showDeleteId, setShowDeleteId, fetchChatHistoryDetail };
}
