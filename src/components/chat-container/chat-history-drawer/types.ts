export interface IChatHistoryItem {
  id: string | number;
  title: string;
  modelName: string;
  modelId: string;
  embedModelId?: string;
  thinkingEnabled?: boolean;
  createdAt: string;
  updatedAt?: string;
}

export interface IChatDetailItem {
  id: string | number;
  sessionId: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
  updatedAt?: string;
  modelId?: string;
  modelName?: string;
}
