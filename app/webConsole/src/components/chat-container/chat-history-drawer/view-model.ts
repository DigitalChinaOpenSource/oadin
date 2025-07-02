import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { type ChatMessageItem } from '@res-utiles/ui-components';
import { message } from 'antd';
import { ModelData } from '@/types';
import useChatStore from '@/components/chat-container/store/useChatStore.ts';
import useSelectedModelStore from '@/store/useSelectedModel.ts';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import useUploadFileStore from '../store/useUploadFileListStore';
import { IChatHistoryItem, GroupedChatHistory, IChatDetailItem, IChatHistoryDrawerProps } from './types';
import { IModelSquareParams } from '@/types';
import { convertMessageFormat } from '../utils/historyMessageFormat';
import { createNewChat } from '../utils';

import dayjs from 'dayjs';
import { getSessionIdFromUrl, setSessionIdToUrl } from '@/utils/sessionParamUtils';

export function useChatHistoryDrawer(props: IChatHistoryDrawerProps) {
  const { onHistorySelect, onHistoryDrawerClose } = props;
  // 获取对话store
  const { setHistoryVisible, setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();
  const { setUploadFileList } = useUploadFileStore();
  // 从URL获取当前会话ID
  const currentSessionId = getSessionIdFromUrl();

  const { setSelectedModel } = useSelectedModelStore();
  const [chatHistory, setChatHistory] = useState<IChatHistoryItem[]>([]);
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
        if (!data || !data.length) return;
        // 过滤空对话
        const tempData = data.filter((item: IChatHistoryItem) => item.title);
        setChatHistory(tempData);
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
          props.handleCreateNewChat?.();
          setHistoryVisible(false);
        }
        if (!data) return;
      },
      onFinally: () => {
        setShowDeleteId(null);
      },
    },
  );

  // 获取历史对话详情
  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        if (!data || !data.length) return message.error('获取历史对话记录失败');
        const tempData = convertMessageFormat(data) as any;
        fetchModelDetail(data[data.length - 1].modelId, tempData, data[data.length - 1].sessionId);
      },
      onError: () => {
        message.error('获取历史对话记录失败');
      },
    },
  );

  const getModelList = async (params: IModelSquareParams, modelId: string) => {
    try {
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', params);
      if (data && data.data && data.data.length) {
        return data.data.find((item) => item.id === modelId);
      } else {
        return undefined; // 返回空数组表示没有数据
      }
    } catch (error) {
      return undefined;
    }
  };

  // 根据模型id获取模型详情
  const fetchModelDetail = async (modelId: string, messages: ChatMessageItem[], sessionId: string) => {
    let res;
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
      setSessionIdToUrl(sessionId, 'history');
      setMessages(messages);
      setHistoryVisible(false);
    } else {
      message.error('获取历史记录详情失败，未找到对应模型');
    }
  };

  const handleHistoryClick = async (id: string) => {
    if (delHistoryLoading) {
      return;
    }
    setSelectMcpList([]);
    setMessages([]);
    setUploadFileList([]);
    try {
      // 获取详细对话内容
      const historyDetail = await fetchChatHistoryDetail(id);

      // 如果提供了选择回调，则调用
      if (onHistorySelect && historyDetail) {
        onHistorySelect(id, historyDetail);
      }
    } catch (error) {
      console.error('加载历史对话失败:', error);
      // 可以添加错误提示
    }
  };

  /**
   * 按日期对聊天历史记录进行分组
   * @param history 聊天历史记录数组
   * @returns 分组后的聊天历史记录对象
   */
  function groupChatHistoryByDate(history: IChatHistoryItem[]): GroupedChatHistory {
    const now = dayjs();
    const todayStart = now.startOf('day');
    const yesterdayStart = now.subtract(1, 'day').startOf('day');
    const last7DaysStart = now.subtract(7, 'day').startOf('day');

    return history.reduce<GroupedChatHistory>(
      (groups, item) => {
        const itemDate = dayjs(item.createdAt);

        if (itemDate.isSame(todayStart, 'day')) {
          groups.today.push(item);
        } else if (itemDate.isSame(yesterdayStart, 'day')) {
          groups.yesterday.push(item);
        } else if (itemDate.isAfter(last7DaysStart) && itemDate.isBefore(yesterdayStart)) {
          groups.last7Days.push(item);
        } else {
          groups.earlier.push(item);
        }

        return groups;
      },
      { today: [], yesterday: [], last7Days: [], earlier: [] },
    );
  }

  useEffect(() => {
    fetchChatHistory();
  }, []);

  return {
    historyLoading,
    fetchChatHistory,
    chatHistory,
    delHistoryLoading,
    deleteChatHistory,
    showDeleteId,
    setShowDeleteId,
    groupChatHistoryByDate,
    handleHistoryClick,
    onHistoryDrawerClose,
  };
}
