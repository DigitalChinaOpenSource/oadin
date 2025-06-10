import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import { IPlaygroundSession } from './types';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型
  const { selectedModel } = useSelectedModelStore();
  const { setCurrentSessionId, createNewChat } = useChatStore();

  useEffect(() => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      fetchCreateChat({ modelName: selectedModel.name });
    }
  }, [selectedModel]);

  const { run: fetchCreateChat } = useRequest(
    async (params: IPlaygroundSession) => {
      const data = await httpRequest.post('/playground/session/create', {
        ...params,
      });
      return data?.data || {};
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
    fetchCreateChat({
      modelName: selectedModel?.name || '',
    });
  };
  return {
    isUploadVisible,
    setIsUploadVisible,
    handleCreateNewChat,
  };
}
