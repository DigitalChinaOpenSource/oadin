import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IChatDetailItem, IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';
import { chatHistoryData } from './mock-data.ts';
import useChatStore from '@/components/chat-container/store/useChatStore.ts';
import { message } from 'antd';
import { IModelDataItem, ModelData } from '@/types';
import useSelectedModelStore from '@/store/useSelectedModel.ts';
import { MessageType } from '@res-utiles/ui-components';
import { IModelSquareParams } from '@/components/model-manage-tab/model-list-content/view-model.ts';

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
        if (!data || !data.length) return message.error('获取历史对话记录失败');

        fetchModelDetail(data[data.length - 1].modelId, data, data[data.length - 1].sessionId);
      },
      onError: (error) => {
        message.error('获取历史对话记录失败');
      },
    },
  );

  const getModelList = async (params: IModelSquareParams, modelId: string) => {
    const data = await httpRequest.get<ModelData>('/control_panel/model/square', params);
    if (data && data.data && data.data.length) {
      return data.data.find((item) => item.id === modelId);
    } else {
      return undefined; // 返回空数组表示没有数据
    }
  };

  // 根据模型id获取模型详情
  const fetchModelDetail = async (modelId: string, messages: MessageType[], sessionId: string) => {
    let res = undefined;
    const localParams: IModelSquareParams = {
      service_source: 'local',
      page_size: 999,
      mine: false,
    };
    res = await getModelList(localParams, modelId);
    if (!res) {
      const remoteParams: IModelSquareParams = {
        service_source: 'remote',
        page_size: 999,
        mine: false,
        env_type: 'product',
      };
      res = await getModelList(remoteParams, modelId);
    }
    if (res) {
      setSelectedModel(res);
      setIsSelectedModel(true);
      setCurrentSessionId(sessionId);
      setMessages(messages);
      setHistoryVisible(false);
    } else {
      message.error('获取历史记录详情失败，未找到对应模型');
    }
  };

  useEffect(() => {
    fetchChatHistory();
  }, []);

  return { historyLoading, fetchChatHistory, chatHistory, delHistoryLoading, deleteChatHistory, showDeleteId, setShowDeleteId, fetchChatHistoryDetail };
}
