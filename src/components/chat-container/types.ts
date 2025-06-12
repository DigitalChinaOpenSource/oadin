export interface IPlaygroundSession {
  sessionId?: string; // 指定则复用，否则自动生成
  title?: string; // 前端可传，后端可自动生成
  modelId: string; // 模型唯一标识
  embedModelId?: string; // RAG 场景下使用
  ragEnabled?: boolean; // 是否启用 RAG
  fileIds?: string[]; // RAG 场景下关联文件
}
