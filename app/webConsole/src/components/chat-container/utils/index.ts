import useChatStore from '../store/useChatStore';
import useUploadFileListStore from '../store/useUploadFileListStore';
import { setSessionIdToUrl } from '@/utils/sessionParamUtils';
export const createNewChat = () => {
  // 清空当前对话内容
  useChatStore.getState().setMessages([]);
  // 清空上传文件列表
  useUploadFileListStore.getState().setUploadFileList([]);
  // 清除会话ID和来源参数
  setSessionIdToUrl('');
};
