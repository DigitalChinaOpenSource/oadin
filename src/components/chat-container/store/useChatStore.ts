import { create } from 'zustand';
import type { MessageType } from '@res-utiles/ui-components';
import type { UploadFile } from 'antd';
import { persist } from 'zustand/middleware';

// 用于 localStorage 中的 key
const STORAGE_KEY = 'vanta-chat-store';

interface ChatState {
  // 聊天状态
  messages: MessageType[];
  uploadFileList: UploadFile[];
  historyVisible: boolean;
  currentSessionId: string;

  // 操作方法
  setMessages: (messages: MessageType[]) => void;
  addMessage: (message: MessageType) => void;
  setUploadFileList: (files: UploadFile[]) => void;
  setHistoryVisible: (visible: boolean) => void;
  setCurrentSessionId: (sessionId: string) => void;
  createNewChat: () => void;
}

const useChatStore = create<ChatState>()(
  persist(
    (set) => ({
      // 初始状态
      messages: [],
      uploadFileList: [],
      historyVisible: false,
      currentSessionId: '',

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
      setCurrentSessionId: (sessionId) => set({ currentSessionId: sessionId }),
    }),
    {
      name: STORAGE_KEY, // localStorage 中使用的键名
      partialize: (state) => ({
        // 只持久化这些字段到 localStorage
        currentSessionId: state.currentSessionId,
        messages: state.messages,
      }),
    },
  ),
);

export default useChatStore;
