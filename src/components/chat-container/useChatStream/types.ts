// 请求参数接口
export interface ChatRequestParams {
  content: string;
  SessionID?: string;
  mcpIds?: string[];
}

// 响应数据接口
export interface ChatResponseData {
  bcode: {
    business_code: number;
    message: string;
  };
  data: {
    id: string;
    session_id: string;
    content: string;
    is_complete: boolean;
    type: string;
    thinking?: string;
    tool_calls?: IToolCall[];
  };
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
