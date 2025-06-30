import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';
import { getSessionIdFromUrl, setSessionIdToUrl, saveSessionIdToStorage, getSessionSource } from '@/utils/sessionParamUtils';
import { IChatDetailItem } from './chat-history-drawer/types';
import { MessageType } from '@res-utiles/ui-components';
import { IModelSquareParams, ModelData } from '@/types';
import { convertMessageFormat } from './utils/historyMessageFormat';
import embedDownloadEventBus from '@/utils/embedDownload';

export default function useViewModel() {
  const { selectedModel, setSelectedModel } = useSelectedModelStore();
  const { createNewChat, messages, setUploadFileList, setMessages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();

  const [isUploadVisible, setIsUploadVisible] = useState(false);
  const [prevModelId, setPrevModelId] = useState<string | undefined>(selectedModel?.id);
  const [prevSessionId, setPrevSessionId] = useState<string | null>(null);
  const [initialized, setInitialized] = useState(false);
  const [isDownloadEmbed, setIsDownloadEmbed] = useState<boolean>(false);

  const isLoadingHistory = useRef(false);

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search);
    const sessionIdFromUrl = urlParams.get('sessionId');
    const sessionIdFromStorage = sessionStorage.getItem('currentSessionId');

    if (!sessionIdFromUrl && sessionIdFromStorage) {
      setSessionIdToUrl(sessionIdFromStorage);
    }
    // 在组件卸载时保存会话ID到sessionStorage
    return () => {
      saveSessionIdToStorage();
    };
  }, []);

  useEffect(() => {
    embedDownloadEventBus.on('embedDownloadComplete', () => {
      setIsDownloadEmbed(true);
    });
    return () => {
      embedDownloadEventBus.off('embedDownloadComplete');
    };
  }, []);

  // 从URL中获取当前会话ID
  const currentSessionId = getSessionIdFromUrl();
  const source = getSessionSource();

  useEffect(() => {
    if (initialized) return;
    // 如果来源是history，历史记录已经在chat-history-drawer组件中处理，这里不做任何操作
    if (source === 'history') {
      setPrevSessionId(currentSessionId);
      setInitialized(true);
      return;
    }

    if (currentSessionId && selectedModel?.id) {
      // 有会话ID且有模型，加载历史记录
      fetchChatHistoryDetail(currentSessionId);
      setPrevSessionId(currentSessionId);
    } else if (selectedModel?.id) {
      // 无会话ID但有模型，创建新会话
      fetchCreateChat({ modelId: selectedModel.id });
      setPrevModelId(selectedModel.id);
    }
    setInitialized(true);
  }, [initialized, selectedModel]);

  useEffect(() => {
    if (!initialized || isLoadingHistory.current) return;

    // 如果会话ID变化了
    if (currentSessionId !== prevSessionId) {
      const source = getSessionSource();
      // 来自历史记录的会话变更，已在chat-history-drawer中处理，这里仅更新状态
      if (source === 'history' || source === 'new') {
        setPrevSessionId(currentSessionId);
      }
      // 其他来源（如URL直接修改或分享链接）且有会话ID，需要加载历史记录
      else if (currentSessionId && !isLoadingHistory.current) {
        isLoadingHistory.current = true;
        fetchChatHistoryDetail(currentSessionId);
        setPrevSessionId(currentSessionId);
      }
    }
  }, [currentSessionId, initialized, prevSessionId]);

  // 处理模型变化
  useEffect(() => {
    if (!initialized) return;

    // 如果没有选定模型，清除URL中的sessionId和source参数
    if (!selectedModel) {
      setSessionIdToUrl(''); // 传入空字符串会清除sessionId和source参数
      setPrevSessionId(null);
      return;
    }

    // 当模型变化时，自动创建新会话
    if (selectedModel.id !== prevModelId && prevModelId !== undefined) {
      // 清空当前对话内容
      createNewChat();
      setSelectMcpList([]);

      // 创建新会话
      fetchCreateChat({
        modelId: selectedModel.id,
        embedModelId: isDownloadEmbed ? '87c0b009-2d93-4f00-9662-3330376662613373163373263' : undefined,
      });
    }

    // 更新 prevModelId
    setPrevModelId(selectedModel.id);
  }, [selectedModel, initialized, prevModelId, isDownloadEmbed]);

  useEffect(() => {
    fetchAllModels();
  }, []);

  useEffect(() => {
    if (isDownloadEmbed && currentSessionId && selectedModel?.id) {
      fetchChooseModelNotify({
        service_name: 'embed',
        local_provider: 'local_ollama_embed',
      });
      fetchChangeModel({
        sessionId: currentSessionId,
        modelId: selectedModel?.id || '',
        embedModelId: 'bc8ca0995fcd651',
      });
    }
  }, [isDownloadEmbed, currentSessionId, selectedModel?.id]);

  useEffect(() => {
    return () => {
      setUploadFileList([]);
      setSelectMcpList([]);
    };
  }, []);

  // 选择模型后需要将所选择的通知奥丁
  const { run: fetchChooseModelNotify } = useRequest(async (params: { service_name: string; local_provider?: string; remote_provider?: string }) => {
    if (!params?.service_name) return;
    const data = await httpRequest.put('/service', {
      ...params,
    });
    return data || {};
  });

  // 每次切换会话模型，都需要通知后端
  const { run: fetchChangeModel } = useRequest(async (params: { sessionId: string; modelId: string; embedModelId: string }) => {
    if (!params?.sessionId || !params.modelId || !params.embedModelId) {
      return {};
    }
    const data = await httpRequest.post('/playground/session/model', {
      ...params,
    });
    return data?.data || {};
  });

  // 查询当前 sessionId 的历史对话
  const { run: fetchChatHistoryDetail } = useRequest(
    async (sessionId: string) => {
      const data = await httpRequest.get<IChatDetailItem[]>(`/playground/messages?sessionId=${sessionId}`);
      return data || [];
    },
    {
      manual: true,
      onSuccess: (data: IChatDetailItem[]) => {
        isLoadingHistory.current = false; // 重置加载标志

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
        isLoadingHistory.current = false; // 重置加载标志
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
        console.log('获取所有下载模型列表', data);
        if (!data || data?.length === 0) return;
        const isEmbed = data.some((item) => item.model_name === 'quentinz/bge-large-zh-v1.5:f16' && item.status === 'downloaded');
        setIsDownloadEmbed(isEmbed);
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
      // 保存来源信息，防止被覆盖
      const source = getSessionSource();

      setSelectedModel(res);
      setMessages(messages);

      // 更新prevModelId，避免模型变化触发新建会话
      setPrevModelId(res.id);
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
        return undefined;
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
          // 1. 首先清空内容和选择的MCP
          setSelectMcpList([]);
          createNewChat();

          // 2. 更新会话相关状态
          setPrevSessionId(data.id);

          // 3. 最后更新 URL
          setSessionIdToUrl(data.id, 'new');
        }
      },
      onError: () => {},
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
      embedModelId: isDownloadEmbed ? '87c0b009-2d93-4f00-9662-3330376662613373163373263' : undefined,
    });
  };

  return {
    isDownloadEmbed,
    isUploadVisible,
    setIsUploadVisible,
    handleCreateNewChat,
  };
}
