import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型
  const { selectedModel } = useSelectedModelStore();
  const { setCurrentSessionId, createNewChat } = useChatStore();

  useEffect(() => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      fetchCreateChat({ modelId: selectedModel.id });
      // TODO 通过判断 sessionId 查对应的历史记录，如果没有则新建一个会话
    }
  }, [selectedModel]);

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
          setCurrentSessionId(data.id);
        }
      },
    },
  );

  const handleCreateNewChat = () => {
    createNewChat();
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
