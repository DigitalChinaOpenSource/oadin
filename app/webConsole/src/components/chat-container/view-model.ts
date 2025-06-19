import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from './store/useChatStore';
import useSelectMcpStore from '@/store/useSelectMcpStore';
import { IPlaygroundSession } from './types';
import { message } from 'antd';

export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型
  const { selectedModel } = useSelectedModelStore();
  const { setCurrentSessionId, createNewChat, messages } = useChatStore();
  const { setSelectMcpList } = useSelectMcpStore();

  const currentSessionId = useChatStore((state) => state.currentSessionId);

  useEffect(() => {
    // 如果当前会话ID不存在且已选择模型，则创建新的会话
    if (!currentSessionId && selectedModel?.id) {
      fetchCreateChat({ modelId: selectedModel.id });
    }
  }, [currentSessionId, selectedModel]);

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
