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
import { getSessionIdFromUrl, setSessionIdToUrl, saveSessionIdToStorage } from '@/utils/sessionParamUtils';
import { IChatDetailItem, IMessageResponse } from './types';
import { IModelSquareParams, ModelData } from '@/types';
import { convertMessageFormat } from './utils/historyMessageFormat';
import embedDownloadEventBus from '@/utils/embedDownloadEventBus';
import { useChatStream } from './useChatStream';
import { EMBEDMODELID } from '@/constants';

function useInitialization() {
  const [initialized, setInitialized] = useState(false);
  const [isDownloadEmbed, setIsDownloadEmbed] = useState<boolean>(false);
  return { initialized, setInitialized, isDownloadEmbed, setIsDownloadEmbed };
}

function useSessionManagement() {
  const [prevSessionId, setPrevSessionId] = useState<string | null>(null);
  const currentSessionId = getSessionIdFromUrl();

  return { prevSessionId, setPrevSessionId, currentSessionId };
}

function useModelManagement() {
  const { selectedModel, setSelectedModel } = useSelectedModelStore();
  const [prevModelId, setPrevModelId] = useState<string | undefined>(selectedModel?.id);
  const [thinkingActive, setThinkingActive] = useState<boolean>(false);
  return { selectedModel, setSelectedModel, prevModelId, setPrevModelId, thinkingActive, setThinkingActive };
}

export default function useViewModel() {
  const { setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();
  const { setUploadFileList } = useUploadFileListStore();
  const [isUploadVisible, setIsUploadVisible] = useState(false);

  const { initialized, setInitialized, isDownloadEmbed, setIsDownloadEmbed } = useInitialization();
  const { prevSessionId, setPrevSessionId, currentSessionId } = useSessionManagement();
  const { selectedModel, setSelectedModel, prevModelId, setPrevModelId, thinkingActive, setThinkingActive } = useModelManagement();
  const { cancelRequest } = useChatStream();

  const isLoadingHistory = useRef(false);
  const isLoadingHistoryModel = useRef(false);

  const { run: fetchChooseModelNotify } = useRequest(
    async (params: { service_name: string; local_provider?: string; remote_provider?: string; hybrid_policy?: string }) => {
      if (!params?.service_name) return;
      const data = await httpRequest.put('/service', { ...params });
      return data || {};
    },
    {
      manual: true,
    },
  );

  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IMessageResponse['data']>(`/playground/messages?sessionId=${sessionId}`);
      return data;
    },
    {
      manual: true,
      onSuccess: (data: IMessageResponse['data']) => {
        isLoadingHistory.current = false;
        const { messages, thinkingActive } = data;

        setThinkingActive(thinkingActive);
        if (!messages || !messages.length) return;
        const inputMessages = messages.map((item) => ({
          ...item,
          id: typeof item.id === 'number' ? String(item.id) : item.id,
        }));
        const tempData = convertMessageFormat(inputMessages) as any;
        fetchModelDetail(messages[messages.length - 1].modelId || '', tempData, messages[messages.length - 1].sessionId);
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
        const isEmbed = (data || []).some((item) => item.model_name === 'quentinz/bge-large-zh-v1.5:f16' && item.status === 'downloaded');
        setIsDownloadEmbed(isEmbed);

        const isSelectedModel = (data || []).some((item) => item.model_name === selectedModel?.name && item.status === 'downloaded');
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
          setSessionIdToUrl(data.id);
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
      modelId: selectedModel?.id,
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
        isLoadingHistoryModel.current = true;
        setSelectedModel(res);
        setMessages(messages as ChatMessageItem[]);
        setPrevModelId(res.id);
        // 在下一个事件循环中重置标志，确保useEffect能正确判断
        setTimeout(() => {
          isLoadingHistoryModel.current = false;
        }, 0);
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

    if (!sessionIdFromUrl && sessionIdFromStorage) {
      setSessionIdToUrl(sessionIdFromStorage);
    }

    // 3. 监听 embed 下载事件
    const handleEmbedComplete = () => {
      setIsDownloadEmbed(true);
      if (selectedModel && currentSessionId) {
        fetchChooseModelNotify({
          service_name: 'embed',
          hybrid_policy: 'always_local',
          local_provider: 'local_ollama_embed',
        });
      }
    };
    embedDownloadEventBus.on('embedDownloadComplete', handleEmbedComplete);

    // 4. 清理函数
    return () => {
      saveSessionIdToStorage();
      cancelRequest();
      embedDownloadEventBus.off('embedDownloadComplete', handleEmbedComplete);
    };
  }, []);

  useEffect(() => {
    if (initialized) return;

    // 初始化逻辑：如果有 sessionId，加载历史记录；如果没有且有模型，创建新会话
    if (currentSessionId) {
      setPrevSessionId(currentSessionId);
      fetchChatHistoryDetail(currentSessionId);
    } else if (selectedModel?.id) {
      handleCreateNewChat();
      setPrevModelId(selectedModel.id);
    }

    setInitialized(true);
  }, [initialized, selectedModel, currentSessionId, fetchChatHistoryDetail, handleCreateNewChat, setPrevModelId, setPrevSessionId]);

  useEffect(() => {
    if (!initialized) return;

    // 会话ID变化处理
    if (currentSessionId !== prevSessionId && !isLoadingHistory.current) {
      if (currentSessionId) {
        isLoadingHistory.current = true;
        fetchChatHistoryDetail(currentSessionId);
      }
      setPrevSessionId(currentSessionId);
    }

    // 模型变化处理
    // 只有在不是加载历史记录模型时，才创建新会话
    if (selectedModel?.id !== prevModelId && prevModelId !== undefined && !isLoadingHistoryModel.current) {
      if (selectedModel) {
        handleCreateNewChat();
      } else {
        setSessionIdToUrl('');
        setPrevSessionId(null);
      }
    }

    if (selectedModel?.id !== prevModelId) {
      setPrevModelId(selectedModel?.id);
    }
  }, [initialized, currentSessionId, prevSessionId, selectedModel, prevModelId, fetchChatHistoryDetail, handleCreateNewChat, setPrevSessionId, setPrevModelId]);

  return {
    isDownloadEmbed,
    isUploadVisible,
    setIsUploadVisible,
    handleCreateNewChat,
    thinkingActive,
  };
}
