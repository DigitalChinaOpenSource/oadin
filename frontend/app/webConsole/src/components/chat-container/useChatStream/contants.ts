export const ERROR_MESSAGES = {
  // 请求错误
  REQUEST: {
    NO_MODEL_SELECTED: '未选择模型，请先选择一个模型',
    FAILED: '请求发送失败: {0}',
    SERVER_ERROR: '服务器返回错误: {0}',
  },
  // 超时错误
  TIMEOUT: {
    TOTAL: '请求超时：5 分钟内未完成响应',
  },
  // 解析错误
  PARSING: {
    JSON_FAILED: '解析响应数据失败: {0}',
    TOOL_RESPONSE: '解析工具调用响应失败',
  },
  // 连接错误
  CONNECTION: {
    INTERRUPTED: '\n\n「消息传输中断」',
    RESPONSE_INTERRUPTED: '\n\n「回复已中断」',
    READ_FAILED: '读取响应内容失败: {0}',
  },
  // 服务器响应错误
  RESPONSE: {
    EMPTY: '服务器响应异常，未能获取到有效数据',
    NON_STREAMING: '服务器未返回流式数据',
    CANNOT_PARSE: '服务器返回了非流式数据，但无法解析内容',
  },
  // 工具调用错误
  TOOL: {
    EXECUTION_FAILED: '工具执行失败: {0}',
    HANDLER_NOT_INITIALIZED: '工具调用处理程序未初始化',
    CONTINUE_FAILED: '继续对话失败: {0}',
  },
};

// 超时设置（毫秒）
export const TIMEOUT_CONFIG = {
  TOTAL: 500000, // 5 分钟总超时
};

export enum ErrorType {
  REQUEST = 'request', // 请求错误
  TIMEOUT = 'timeout', // 超时错误
  PARSING = 'parsing', // 解析错误
  CONNECTION = 'connection', // 连接错误
  TOOL = 'tool', // 工具调用错误
  INTERNAL = 'internal', // 内部错误
}
