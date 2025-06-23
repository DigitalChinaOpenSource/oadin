import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';
import { getSessionIdFromUrl, setSessionIdToUrl } from '@/utils/sessionParamUtils';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型
  const { selectedModel } = useSelectedModelStore();
  const { createNewChat, messages, setUploadFileList } = useChatStore();
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
