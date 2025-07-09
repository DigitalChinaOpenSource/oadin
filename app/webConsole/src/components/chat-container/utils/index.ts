import useChatStore from '../store/useChatStore';
import useUploadFileListStore from '../store/useUploadFileListStore';
import { setSessionIdToUrl } from '@/utils/sessionParamUtils';
import { healthRequest, httpRequest } from '@/utils/httpRequest';

export const createNewChat = () => {
  // 清空当前对话内容
  useChatStore.getState().setMessages([]);
  // 清空上传文件列表
  useUploadFileListStore.getState().setUploadFileList([]);
  // 清除会话ID和来源参数
  setSessionIdToUrl('');
  sessionStorage.setItem('isGenChatTitle', 'false');
};

export const fetchCheckEngineStatus = async (): Promise<boolean> => {
  const res = await healthRequest.get('/engine/health');
  // 1 可用；0 不可用；
  return res.status === 1;
};

interface ModelInfo {
  model_name?: string;
  provider_name?: string;
  status?: string;
  created_at?: string | Date;
  updated_at?: string | Date;
}

export const fetchAllModels = async (): Promise<ModelInfo[]> => {
  const data = await httpRequest.get('/model');
  return data || [];
};

export const checkIsModelDownloaded = async (modelName: string): Promise<boolean> => {
  const data = await fetchAllModels();
  const model = data.find((item) => item.model_name === modelName && item.provider_name === 'local_ollama_chat');
  return model?.status === 'downloaded';
};
