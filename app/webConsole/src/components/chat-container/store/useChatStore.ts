import { create } from 'zustand';
import type { MessageType } from '@res-utiles/ui-components';
import type { UploadFile } from 'antd';
import { persist } from 'zustand/middleware';
import { IUploadFile } from '../types';

// 用于 localStorage 中的 key
const STORAGE_KEY = 'vanta-chat-store';

interface ChatState {
  messages: MessageType[];
  uploadFileList: UploadFile[];
  historyVisible: boolean;
  currentSessionId: string;

  // 操作方法
  setMessages: (messages: MessageType[]) => void;
  addMessage: (message: MessageType, isReplace?: boolean) => string;
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
      createNewChat: () =>
        set({
          messages: [],
          uploadFileList: [],
        }),
      setCurrentSessionId: (sessionId) => set({ currentSessionId: sessionId }),
    }),
    {
      name: STORAGE_KEY,
      partialize: (state) => ({
        currentSessionId: state.currentSessionId,
        messages: state.messages,
        uploadFileList: state.uploadFileList,
      }),
    },
  ),
);

export default useChatStore;
