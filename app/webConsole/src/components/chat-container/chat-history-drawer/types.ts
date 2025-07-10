import type { ChatMessageItem } from '@res-utiles/ui-components';

export interface IChatHistoryItem {
  id: string | number;
  title: string;
  modelName: string;
  modelId?: string;
  embedModelId?: string;
  thinkingEnabled?: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface GroupedChatHistory {
  today: IChatHistoryItem[];
  yesterday: IChatHistoryItem[];
  last7Days: IChatHistoryItem[];
  earlier: IChatHistoryItem[];
}

export interface IChatHistoryDrawerProps {
  onHistoryDrawerClose?: () => void;
  onHistorySelect?: (historyId: string, historyMessages: ChatMessageItem[]) => void;
  handleCreateNewChat?: () => void;
}
