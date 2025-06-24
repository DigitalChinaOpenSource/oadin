import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';
import { getSessionIdFromUrl, setSessionIdToUrl } from '@/utils/sessionParamUtils';
import { IChatDetailItem } from './chat-history-drawer/types';
import { MessageType } from '@res-utiles/ui-components';
import { IModelSquareParams, ModelData } from '@/types';
import { convertMessageFormat } from './utils/historyMessageFormat';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型
  const { selectedModel, setSelectedModel, setIsSelectedModel } = useSelectedModelStore();
  const { createNewChat, messages, setUploadFileList, setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();

  // 从URL中获取当前会话ID
  const currentSessionId = getSessionIdFromUrl();

  // 保存上一次使用的模型ID，用于检测模型是否变化
  const prevSelectedModelIdRef = useRef<string | undefined>(selectedModel?.id);

  useEffect(() => {
    // 场景1: 如果当前会话ID不存在且已选择模型，则创建新的会话
    // 场景2: 如果会话ID存在，但选择的模型发生变化，也创建新的会话
    if (selectedModel?.id && (!currentSessionId || prevSelectedModelIdRef.current !== selectedModel.id)) {
      fetchCreateChat({ modelId: selectedModel.id });
    } else if (currentSessionId && !selectedModel?.id) {
      // 页面刷新时，如果存在会话ID但没有选择模型，则获取历史对话详情
      fetchChatHistoryDetail(currentSessionId);
    }

    // 更新上一次使用的模型ID引用
    prevSelectedModelIdRef.current = selectedModel?.id;
  }, [currentSessionId, selectedModel]);

  useEffect(() => {
    return () => {
      setUploadFileList([]);
      setSelectMcpList([]);
    };
  }, []);

  // 获取历史对话详情
  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: IChatDetailItem[]) => {
        if (!data || data.length === 0) return message.error('获取历史对话记录失败');
        // 将 IChatDetailItem[] 转换为 InputMessage[]
        const inputMessages = data.map((item) => ({
          ...item,
          id: typeof item.id === 'number' ? String(item.id) : item.id,
        }));
        const tempData = convertMessageFormat(inputMessages);
        fetchModelDetail(data[data.length - 1].modelId || '', tempData, data[data.length - 1].sessionId);
      },
      onError: () => {
        message.error('获取历史对话记录失败');
      },
    },
  );

  // 根据模型id获取模型详情
  const fetchModelDetail = async (modelId: string, messages: MessageType[], sessionId: string) => {
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
      setIsSelectedModel(true);
      setSessionIdToUrl(sessionId);
      setMessages(messages);
    } else {
      message.error('获取历史记录详情失败，未找到对应模型');
    }
  };

  // 获取模型列表
  const getModelList = async (params: IModelSquareParams, modelId: string) => {
    try {
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', params);
      if (data && data.data && data.data.length) {
        return data.data.find((item) => item.id === modelId);
      } else {
        return undefined; // 返回undefined表示没有找到对应模型
      }
    } catch (error) {
      return undefined;
    }
  };

  const { run: fetchCreateChat } = useRequest(
    async (params: IPlaygroundSession) => {
      const data = await httpRequest.post('/playground/session', {
        ...params,
      });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (data.id) {
          setSessionIdToUrl(data.id);
          setSelectMcpList([]);
          createNewChat();
        }
      },
    },
  );

  const handleCreateNewChat = () => {
    if (messages.length === 0) {
      return;
    }
    if (!selectedModel?.id) {
      message.error('请先选择一个模型');
      return;
    }
    fetchCreateChat({
      modelId: selectedModel?.id || '',
    });
  };
  return {
    isUploadVisible,
    setIsUploadVisible,
    handleCreateNewChat,
  };
}
