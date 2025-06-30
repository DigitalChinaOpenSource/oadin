import { create } from 'zustand';
import type { MessageType } from '@res-utiles/ui-components';
import type { UploadFile } from 'antd';
import { IUploadFile } from '../types';

interface ChatState {
  messages: MessageType[];
  uploadFileList: UploadFile[];
  historyVisible: boolean;
  isLoading: boolean;
  isUploading: boolean;

  // 操作方法
  setMessages: (messages: MessageType[]) => void;
  addMessage: (message: MessageType, isReplace?: boolean) => string;
  setUploadFileList: (files: UploadFile[]) => void;
  setHistoryVisible: (visible: boolean) => void;
  createNewChat: () => void;
  setIsLoading: (isLoading: boolean) => void;
  setIsUploading: (isUploading: boolean) => void;
}

const useChatStore = create<ChatState>((set, get) => ({
  // 初始状态
  messages: [],
  uploadFileList: [],
  historyVisible: false,
  isLoading: false,
  isUploading: false,

  // 操作方法
  setMessages: (messages) => set({ messages }),
  addMessage: (message: MessageType, isReplace?: boolean): string => {
    // 生成消息ID
    const messageId = message.id;
    message.id = messageId;

    set((state) => {
      const messages = [...state.messages];

      // 如果是替换模式，查找并替换相同ID的消息
      if (isReplace) {
        const msgIndex = messages.findIndex((msg) => msg.id === messageId);
        if (msgIndex >= 0) {
          messages[msgIndex] = message;
        } else {
          messages.push(message);
        }
      } else {
        // 正常添加
        messages.push(message);
      }

      return { messages };
    });

    return messageId;
  },
  setUploadFileList: (files) => set({ uploadFileList: files }),
  setHistoryVisible: (visible) => set({ historyVisible: visible }),
  createNewChat: () => {
    set({
      messages: [],
      uploadFileList: [],
    });
  },
  setIsLoading: (isLoading) => set({ isLoading: isLoading ?? false }),
  setIsUploading: (isUploading) => set({ isUploading: isUploading ?? false }),
}));

export default useChatStore;
