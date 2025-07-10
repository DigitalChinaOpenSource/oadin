export interface IPlaygroundSession {
  sessionId?: string; // 指定则复用，否则自动生成
  title?: string; // 前端可传，后端可自动生成
  modelId: string; // 模型唯一标识
  embedModelId?: string; // RAG 场景下使用
  ragEnabled?: boolean; // 是否启用 RAG
  fileIds?: string[]; // RAG 场景下关联文件
  modelName: string;
}

export interface IUploadFile {
  created_at: string;
  id: string;
  name: string;
  session_id: string;
  size: number;
  type: any;
  status: 'error' | 'progress' | 'success'; // 文件状态
}

export interface IChangeModelParams {
  sessionId: string;
  modelId: string;
  modelName: string;
  embedModelId?: string;
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

export interface IMessageResponse {
  bcode: {
    business_code: number;
    message: string;
  };
  data: { messages: IChatDetailItem[]; thinkingActive: boolean };
}
