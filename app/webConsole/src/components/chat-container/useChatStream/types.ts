// 请求参数接口
export interface ChatRequestParams {
  content: string;
  SessionID?: string;
  mcpIds?: string[];
  embedModelId?: string;
}

// 响应数据接口
export interface ChatResponseData {
  bcode: {
    business_code: number;
    message: string;
  };
  data: IStreamData;
}

export interface IStreamData {
  id: string;
  session_id: string;
  content: string;
  is_complete: boolean;
  type: string;
  thoughts?: string;
  tool_calls?: IToolCall[];
  total_duration?: number;
  tool_group_id?: string;
}

// 工具调用接口
export interface IToolCall {
  function: {
    name: string;
    arguments: Record<string, any>;
  };
}

// 工具调用参数接口
export interface IRunToolParams {
  mcpId: string;
  toolName: string;
  toolArgs: Record<string, any>;
}

// 流式请求回调接口
export interface StreamCallbacks {
  onDataReceived: (data: ChatResponseData) => void;
  onComplete: () => void;
  onFallbackResponse: (response: Response) => Promise<void>;
}

export interface IContentItem {
  type: 'plain' | 'map' | 'think';
  // 可能是文字，也可能是展示 mcp 的结构
  content: string | any;
  id: string;
}

export interface IToolCallData {
  id: string;
  mcpId?: string;
  desc?: string;
  logo?: string;
  name: string;
  inputParams: string;
  outputParams?: string;
  status: 'success' | 'error' | 'loading' | string;
  duration?: number;
}

// 请求参数接口
export interface ChatRequestParams {
  content: string;
  SessionID?: string;
  mcpIds?: string[];
  toolGroupID?: string;
}

// 响应数据接口
export interface ChatResponseData {
  bcode: {
    business_code: number;
    message: string;
  };
  data: IStreamData;
}

// 工具调用接口
export interface IToolCall {
  function: {
    name: string;
    arguments: Record<string, any>;
  };
}

// 工具调用参数接口
export interface IRunToolParams {
  messageId: string;
  mcpId: string;
  toolName: string;
  toolArgs: Record<string, any>;
}

// 流式请求回调接口
export interface StreamCallbacks {
  onDataReceived: (data: ChatResponseData) => void;
  onComplete: () => void;
  onFallbackResponse: (response: Response) => Promise<void>;
}

// MCP 工具聊天数据接口
export interface IMcpToolChatData {
  status: 'success' | 'error' | 'progress';
  data: IToolCallData[];
  totalDuration: number;
}
