import { create } from 'zustand';
import type { MessageType } from '@res-utiles/ui-components';
import type { UploadFile } from 'antd';

interface ChatState {
  // 聊天状态
  messages: MessageType[];
  uploadFileList: UploadFile[];
  historyVisible: boolean;

  // 操作方法
  setMessages: (messages: MessageType[]) => void;
  addMessage: (message: MessageType) => void;
  setUploadFileList: (files: UploadFile[]) => void;
  setHistoryVisible: (visible: boolean) => void;
  createNewChat: () => void;
}

const useChatStore = create<ChatState>((set) => ({
  // 初始状态
  messages: [],
  uploadFileList: [],
  historyVisible: false,

  // 操作方法
  setMessages: (messages) => set({ messages }),
  addMessage: (message) =>
    set((state) => ({
      messages: [...state.messages, message],
    })),
  setUploadFileList: (files) => set({ uploadFileList: files }),
  setHistoryVisible: (visible) => set({ historyVisible: visible }),
  createNewChat: () =>
    set({
      messages: [],
      uploadFileList: [],
    }),
}));

export default useChatStore;
