import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';
import { chatHistoryData } from './mock-data.ts';

export function useChatHistoryDrawer() {
  // 历史对话记录
  const [chatHistory, setChatHistory] = useState<IChatHistoryItem[]>(chatHistoryData);

  // 用于记录当前显示 Popconfirm 的卡片 id 并设置是否显示确认弹窗
  const [showDeleteId, setShowDeleteId] = useState<string | null>(null);

  // 获取c
  const { loading: historyLoading, run: fetchChatHistory } = useRequest(
    async () => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get('/service/update/history');
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        if (!data) return;
      },
      onError: (error) => {
        console.error('获取历史对话记录失败:', error);
      },
    },
  );

  // 删除对话历史记录
  const { loading: delHistoryLoading, run: deleteChatHistory } = useRequest(
    async (id: string | number) => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get('/service/update/history');
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        // 删除成功 刷新历史记录
        fetchChatHistory();
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

  useEffect(() => {
    fetchChatHistory();
  }, []);

  return { historyLoading, fetchChatHistory, chatHistory, delHistoryLoading, deleteChatHistory, showDeleteId, setShowDeleteId };
}
