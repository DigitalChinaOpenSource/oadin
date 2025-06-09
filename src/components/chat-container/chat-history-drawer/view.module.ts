import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';

const testData: IChatHistoryItem[] = [
  { id: '1', title: '对话1', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-06-09 10:00:00' },
  {
    id: '2',
    title: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎',
    modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎',
    createdAt: '2025-10-01 10:00:00',
  },
  { id: '3', title: '对话3', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-06-09 10:00:00' },
  { id: '4', title: '对话4', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-06-08 10:00:00' },
  { id: '5', title: '对话5', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-06-07 10:00:00' },
  { id: '6', title: '对话6', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '7', title: '对话7', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '8', title: '对话7', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '9', title: '对话7', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '10', title: '对话7', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '11', title: '对话7', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
  { id: '12', title: '对话12', modelName: '上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎上山打老虎', createdAt: '2025-1-01 10:00:00' },
];

export function useChatHistoryDrawer() {
  // 历史对话记录
  const [chatHistory, setChatHistory] = useState<IChatHistoryItem[]>(testData);

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
