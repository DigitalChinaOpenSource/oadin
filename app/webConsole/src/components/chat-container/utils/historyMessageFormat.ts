import { generateUniqueId } from '../useChatStream/utils';

export interface InputMessage {
  id: string;
  sessionId?: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt?: string;
  modelId?: string;
  modelName?: string;
  type?: string;
  toolCalls?: ToolCall[];
}

export interface ToolCall {
  id: string;
  session_id: string;
  message_id: string;
  mcp_id: string;
  mcp_image: string;
  name: string;
  desc: string;
  input_params: string;
  output_params: string;
  status: boolean;
  execution_time: number;
  created_at: string;
  updated_at: string;
}

export interface McpData {
  mcpId: string;
  name: string;
  desc: string;
  logo: string;
  inputParams: string;
  status: 'success' | 'error';
  startTime: number;
  outputParams: string;
  executionTime: number;
}

export interface PlainContent {
  id: string;
  type: 'plain';
  content: string;
}

export interface ThinkContent {
  id: string;
  type: 'think';
  content: {
    data: string;
    status: 'success' | 'error' | 'pending';
  };
}

export interface McpContent {
  id: string;
  type: 'mcp';
  content: {
    status: 'success' | 'error';
    totalDuration: number;
    data: McpData[];
  };
}

export type ContentItem = PlainContent | ThinkContent | McpContent;

export interface OutputMessage {
  id: string;
  role: 'user' | 'assistant';
  contentList: ContentItem[];
}

export function convertMessageFormat(inputArray: InputMessage[]): OutputMessage[] {
  const results: OutputMessage[] = [];

  inputArray.forEach((item: InputMessage) => {
    const content: string = item.content;

    // 检查是否包含 <think> 标签
    const thinkRegex: RegExp = /<think>([\s\S]*?)<\/think>/;
    const match: RegExpMatchArray | null = content.match(thinkRegex);

    // 如果有 toolCalls，需要拆分成两个消息
    if (item.toolCalls && item.toolCalls.length > 0) {
      // 第一个消息：只包含 mcp 内容
      const mcpData: McpData[] = item.toolCalls.map(
        (toolCall: ToolCall): McpData => ({
          mcpId: toolCall.mcp_id,
          name: toolCall.name,
          desc: toolCall.desc,
          logo: toolCall.mcp_image,
          inputParams: toolCall.input_params,
          status: toolCall.status ? 'success' : 'error',
          startTime: new Date(toolCall.created_at).getTime(),
          outputParams: toolCall.output_params,
          executionTime: toolCall.execution_time,
        }),
      );

      const totalDuration: number = item.toolCalls.reduce((sum, toolCall) => sum + toolCall.execution_time, 0);
      const hasError: boolean = item.toolCalls.some((toolCall) => !toolCall.status);
      const overallStatus: 'success' | 'error' = hasError ? 'error' : 'success';

      // 创建 mcp 消息，放在前面
      results.push({
        id: item.id,
        role: item.role,
        contentList: [
          {
            id: generateUniqueId('content'),
            type: 'mcp',
            content: {
              status: overallStatus,
              totalDuration: totalDuration,
              data: mcpData,
            },
          } as McpContent,
        ],
      });

      // 第二个消息：只包含 plain/think 内容
      const plainMessage: OutputMessage = {
        id: item.id,
        role: item.role,
        contentList: [],
      };

      if (match) {
        // 包含 think 标签的情况
        const beforeThink: string = content.substring(0, match.index!).trim();
        const thinkContent: string = match[1].trim();
        const afterThink: string = content.substring(match.index! + match[0].length).trim();

        // 添加 think 前的内容
        if (beforeThink) {
          plainMessage.contentList.push({
            id: generateUniqueId('content'),
            type: 'plain',
            content: beforeThink,
          } as PlainContent);
        }

        // 添加 think 内容
        plainMessage.contentList.push({
          id: generateUniqueId('content'),
          type: 'think',
          content: {
            data: thinkContent,
            status: 'success',
          },
        } as ThinkContent);

        // 添加 think 后的内容
        if (afterThink) {
          plainMessage.contentList.push({
            id: generateUniqueId('content'),
            type: 'plain',
            content: afterThink,
          } as PlainContent);
        }
      } else {
        // 不包含 think 标签，直接作为 plain 类型
        plainMessage.contentList.push({
          id: generateUniqueId('content'),
          type: 'plain',
          content: content,
        } as PlainContent);
      }

      results.push(plainMessage);
    } else {
      // 没有 toolCalls
      const singleMessage: OutputMessage = {
        id: item.id,
        role: item.role,
        contentList: [],
      };

      if (match) {
        // 包含 think 标签的情况
        const beforeThink: string = content.substring(0, match.index!).trim();
        const thinkContent: string = match[1].trim();
        const afterThink: string = content.substring(match.index! + match[0].length).trim();

        // 添加 think 前的内容（如果存在）
        if (beforeThink) {
          singleMessage.contentList.push({
            id: generateUniqueId('content'),
            type: 'plain',
            content: beforeThink,
          } as PlainContent);
        }

        // 添加 think 内容
        singleMessage.contentList.push({
          id: generateUniqueId('content'),
          type: 'think',
          content: {
            data: thinkContent,
            status: 'success',
          },
        } as ThinkContent);

        // 添加 think 后的内容
        if (afterThink) {
          singleMessage.contentList.push({
            id: generateUniqueId('content'),
            type: 'plain',
            content: afterThink,
          } as PlainContent);
        }
      } else {
        // 不包含 think 标签，直接作为 plain 类型
        singleMessage.contentList.push({
          id: generateUniqueId('content'),
          type: 'plain',
          content: content,
        } as PlainContent);
      }

      results.push(singleMessage);
    }
  });

  return results;
}
