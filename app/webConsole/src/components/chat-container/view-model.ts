import { useState, useEffect, useRef, useCallback } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import type { ChatMessageItem } from '@res-utiles/ui-components';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useUploadFileListStore from './store/useUploadFileListStore';
import { createNewChat } from './utils';
import { IPlaygroundSession } from './types';
import { message } from 'antd';
import { getSessionIdFromUrl, setSessionIdToUrl, saveSessionIdToStorage, getSessionSource } from '@/utils/sessionParamUtils';
import { IChatDetailItem } from './chat-history-drawer/types';
import { IModelSquareParams, ModelData } from '@/types';
import { convertMessageFormat } from './utils/historyMessageFormat';
import embedDownloadEventBus from '@/utils/embedDownload';
import { EMBEDMODELID } from '@/constants';

/** 封装一些自定义hooks */
function useInitialization() {
  const [initialized, setInitialized] = useState(false);
  const [isDownloadEmbed, setIsDownloadEmbed] = useState<boolean>(false);

  return { initialized, setInitialized, isDownloadEmbed, setIsDownloadEmbed };
}

function useSessionManagement() {
  const [prevSessionId, setPrevSessionId] = useState<string | null>(null);
  const currentSessionId = getSessionIdFromUrl();
  const source = getSessionSource();

  return { prevSessionId, setPrevSessionId, currentSessionId, source };
}

function useModelManagement() {
  const { selectedModel, setSelectedModel } = useSelectedModelStore();
  const [prevModelId, setPrevModelId] = useState<string | undefined>(selectedModel?.id);

  return { selectedModel, setSelectedModel, prevModelId, setPrevModelId };
}

export default function useViewModel() {
  const { setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();
  const { setUploadFileList } = useUploadFileListStore();
  const [isUploadVisible, setIsUploadVisible] = useState(false);

  // 使用自定义 hooks
  const { initialized, setInitialized, isDownloadEmbed, setIsDownloadEmbed } = useInitialization();
  const { prevSessionId, setPrevSessionId, currentSessionId, source } = useSessionManagement();
  const { selectedModel, setSelectedModel, prevModelId, setPrevModelId } = useModelManagement();

  const isLoadingHistory = useRef(false);

  // 合并相关的请求函数
  const { run: fetchChooseModelNotify } = useRequest(async (params: { service_name: string; local_provider?: string; remote_provider?: string }) => {
    if (!params?.service_name) return;
    const data = await httpRequest.put('/service', { ...params });
    return data || {};
  });

  const { run: fetchChangeModel } = useRequest(async (params: { sessionId: string; modelId: string; embedModelId: string }) => {
    if (!params?.sessionId || !params.modelId || !params.embedModelId) return {};
    const data = await httpRequest.post('/playground/session/model', { ...params });
    return data?.data || {};
  });

  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: IChatDetailItem[]) => {
        isLoadingHistory.current = false;
        if (!data || data.length === 0) return;

        const inputMessages = data.map((item) => ({
          ...item,
          id: typeof item.id === 'number' ? String(item.id) : item.id,
        }));
        const tempData = convertMessageFormat(inputMessages) as any;
        fetchModelDetail(data[data.length - 1].modelId || '', tempData, data[data.length - 1].sessionId);
      },
      onError: () => {
        isLoadingHistory.current = false;
        message.error('获取历史对话记录失败');
      },
    },
  );

  const { run: fetchAllModels } = useRequest(
    async () => {
      const data = await httpRequest.get('/model');
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: any[]) => {
        if (!data || data?.length === 0) return;

        const isEmbed = data.some((item) => item.model_name === 'quentinz/bge-large-zh-v1.5:f16' && item.status === 'downloaded');
        setIsDownloadEmbed(isEmbed);

        const isSelectedModel = data.some((item) => item.model_name === selectedModel?.name && item.status === 'downloaded');
        if (!isSelectedModel) {
          // 重置所有状态
          setSelectedModel(null);
          setSessionIdToUrl('');
          setPrevModelId(undefined);
          setPrevSessionId(null);
          setMessages([]);
          setUploadFileList([]);
          setSelectMcpList([]);
        }
      },
      onError: () => {
        message.error('获取所有下载模型列表失败');
      },
    },
  );

  const { run: fetchCreateChat } = useRequest(
    async (params: IPlaygroundSession) => {
      const data = await httpRequest.post('/playground/session', { ...params });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (data.id) {
          setSelectMcpList([]);
          createNewChat();
          setPrevSessionId(data.id);
          setSessionIdToUrl(data.id, 'new');
        }
      },
      onError: () => {},
    },
  );

  const handleCreateNewChat = useCallback(() => {
    if (!selectedModel?.id) {
      message.error('请先选择一个模型');
      return;
    }
    const params = {
      modelId: selectedModel?.source === 'local' ? selectedModel?.id : selectedModel?.name || '',
      embedModelId: isDownloadEmbed ? EMBEDMODELID : undefined,
      modelName: selectedModel?.name || '',
    };
    fetchCreateChat(params);
  }, [selectedModel, isDownloadEmbed, fetchCreateChat]);

  const fetchModelDetail = useCallback(
    async (modelId: string, messages: ChatMessageItem[], sessionId: string) => {
      const getModelList = async (params: IModelSquareParams, modelId: string) => {
        try {
          const data = await httpRequest.get<ModelData>('/control_panel/model/square', params);
          if (data && data.data && data.data.length) {
            return data.data.find((item) => item.id === modelId);
          }
          return undefined;
        } catch (error) {
          return undefined;
        }
      };

      let res;
      const localParams: IModelSquareParams = { service_source: 'local', page_size: 999, mine: false };
      res = await getModelList(localParams, modelId);

      if (!res) {
        const remoteParams: IModelSquareParams = { service_source: 'remote', page_size: 999, mine: false, env_type: 'product' };
        res = await getModelList(remoteParams, modelId);
      }

      if (res) {
        setSelectedModel(res);
        setMessages(messages as ChatMessageItem[]);
        setPrevModelId(res.id);
      } else {
        message.error('获取历史记录详情失败，未找到对应模型');
      }
    },
    [setSelectedModel, setMessages, setPrevModelId],
  );

  useEffect(() => {
    // 1. 初始化获取模型列表
    fetchAllModels();

    // 2. 处理 URL 参数和 sessionStorage
    const urlParams = new URLSearchParams(window.location.search);
    const sessionIdFromUrl = urlParams.get('sessionId');
    const sessionIdFromStorage = sessionStorage.getItem('currentSessionId');

    if (source === 'new') {
      const url = new URL(window.location.href);
      url.searchParams.delete('source');
      window.history.replaceState({}, '', url);
    }

    if (!sessionIdFromUrl && sessionIdFromStorage) {
      setSessionIdToUrl(sessionIdFromStorage);
    }

    // 3. 监听 embed 下载事件
    const handleEmbedComplete = () => {
      setIsDownloadEmbed(true);
      if (selectedModel) {
        const tempParams = {
          service_name: 'embed',
          hybrid_policy: `always_${selectedModel.source}`,
        } as any;
        if (selectedModel.source === 'local') {
          tempParams.local_provider = selectedModel.service_provider_name;
        } else if (selectedModel.source === 'remote') {
          tempParams.remote_provider = selectedModel.service_provider_name;
        }
        fetchChooseModelNotify(tempParams);
      }
    };
    embedDownloadEventBus.on('embedDownloadComplete', handleEmbedComplete);

    // 4. 清理函数
    return () => {
      saveSessionIdToStorage();
      embedDownloadEventBus.off('embedDownloadComplete', handleEmbedComplete);
    };
  }, []); // 只在组件挂载时执行一次

  // 合并初始化和会话管理逻辑
  useEffect(() => {
    if (initialized) return;

    // 初始化逻辑
    if (source === 'history') {
      setPrevSessionId(currentSessionId);
      if (currentSessionId) {
        fetchChatHistoryDetail(currentSessionId);
      }
      setInitialized(true);
      return;
    }

    if (currentSessionId && selectedModel?.id) {
      fetchChatHistoryDetail(currentSessionId);
      setPrevSessionId(currentSessionId);
    }

    if (!currentSessionId && selectedModel?.id) {
      handleCreateNewChat();
      setPrevModelId(selectedModel.id);
    }

    setInitialized(true);
  }, [initialized, selectedModel, currentSessionId, source, fetchChatHistoryDetail, handleCreateNewChat, setPrevModelId, setPrevSessionId]);

  // 处理会话 ID 和模型变化
  useEffect(() => {
    if (!initialized) return;

    // 会话ID变化处理
    if (currentSessionId !== prevSessionId && !isLoadingHistory.current) {
      const currentSource = getSessionSource();
      if (currentSource === 'history' || currentSource === 'new') {
        setPrevSessionId(currentSessionId);
      } else if (currentSessionId) {
        isLoadingHistory.current = true;
        fetchChatHistoryDetail(currentSessionId);
        setPrevSessionId(currentSessionId);
      }
    }

    // 模型变化处理
    if (selectedModel?.id !== prevModelId && prevModelId !== undefined && source !== 'history') {
      if (selectedModel) {
        handleCreateNewChat();
      } else {
        setSessionIdToUrl('');
        setPrevSessionId(null);
      }
    }

    // 更新 prevModelId
    if (selectedModel?.id !== prevModelId) {
      setPrevModelId(selectedModel?.id);
    }
  }, [initialized, currentSessionId, prevSessionId, selectedModel, prevModelId, source, fetchChatHistoryDetail, handleCreateNewChat, setPrevSessionId, setPrevModelId]);

  useEffect(() => {
    if (!currentSessionId || !selectedModel?.id) return;

    fetchChangeModel({
      sessionId: currentSessionId,
      modelId: selectedModel?.id || '',
      embedModelId: EMBEDMODELID,
    });
  }, [currentSessionId, selectedModel, fetchChooseModelNotify, fetchChangeModel]);

  return {
    isDownloadEmbed,
    isUploadVisible,
    setIsUploadVisible,
    handleCreateNewChat,
  };
}
