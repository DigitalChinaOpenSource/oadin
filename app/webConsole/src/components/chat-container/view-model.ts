import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';
import { getSessionIdFromUrl, setSessionIdToUrl, saveSessionIdToStorage } from '@/utils/sessionParamUtils';
import { IChatDetailItem } from './chat-history-drawer/types';
import { MessageType } from '@res-utiles/ui-components';
import { IModelSquareParams, ModelData } from '@/types';
import { convertMessageFormat } from './utils/historyMessageFormat';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  const { selectedModel, setSelectedModel, setIsSelectedModel } = useSelectedModelStore();
  const isDownloadEmbed = useModelDownloadStore((state) => state.isDownloadEmbed);
  const { createNewChat, messages, setUploadFileList, setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const sessionIdFromUrl = urlParams.get('sessionId');
    const sessionIdFromStorage = sessionStorage.getItem('currentSessionId');

    // 如果URL中没有sessionId但sessionStorage中有，则从sessionStorage恢复到URL
    if (!sessionIdFromUrl && sessionIdFromStorage) {
      setSessionIdToUrl(sessionIdFromStorage);
    }
  }, []);

  // 从URL中获取当前会话ID
  const currentSessionId = getSessionIdFromUrl();
  // 保存上一次使用的模型ID，用于检测模型是否变化
  const prevSelectedModelIdRef = useRef<string | undefined>(selectedModel?.id);
  // 用于标记是否是页面初始化加载
  const isInitialLoad = useRef(true);

  // 在组件卸载时保存会话ID到sessionStorage
  useEffect(() => {
    return () => {
      saveSessionIdToStorage();
    };
  }, []);

  useEffect(() => {
    // 场景1: 如果有会话ID(URL或sessionStorage)，则尝试获取历史对话详情
    if (currentSessionId) {
      if ((!selectedModel?.id || messages.length === 0) && isInitialLoad.current) {
        // 页面初始化加载时，如果存在会话ID，则获取历史对话详情
        fetchChatHistoryDetail(currentSessionId);
      } else if (prevSelectedModelIdRef.current !== selectedModel?.id && selectedModel?.id) {
        // 如果选择的模型发生变化，则创建新的会话
        fetchCreateChat({ modelId: selectedModel.id });
      }
    } else if (selectedModel?.id) {
      // 场景2: 如果没有会话ID但已选择模型，则创建新的会话
      fetchCreateChat({ modelId: selectedModel.id });
    }

    // 更新上一次使用的模型ID引用
    prevSelectedModelIdRef.current = selectedModel?.id;

    // 首次运行后将初始化标志设为 false
    isInitialLoad.current = false;
  }, [currentSessionId, selectedModel, messages.length]);

  useEffect(() => {
    if (!isDownloadEmbed) return;
    fetchAllModels();
  }, [isDownloadEmbed]);

  useEffect(() => {
    return () => {
      setUploadFileList([]);
      setSelectMcpList([]);
    };
  }, []);

  // 获取所有白泽下载的模型
  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: IChatDetailItem[]) => {
        if (!data || data.length === 0) return;
        // 将 IChatDetailItem[] 转换为 InputMessage[]
        const inputMessages = (data || []).map((item) => ({
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

  // 定义模型数据接口
  interface ModelItem {
    model_name: string;
    [key: string]: unknown;
  }

  // 获取历史对话详情
  const { run: fetchAllModels } = useRequest(
    async () => {
      const data = await httpRequest.get<ModelItem[]>('/model');
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: ModelItem[]) => {
        if (!data || data?.length === 0) return;
        const isEmbed = data.some((item) => item.model_name === 'quentinz/bge-large-zh-v1.5:f16');
        setIsUploadVisible(isEmbed);
      },
      onError: () => {
        message.error('获取所有下载模型列表失败');
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
