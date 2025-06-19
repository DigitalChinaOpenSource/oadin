import { MessageType } from '@res-utiles/ui-components';
import { IToolCall, IToolCallData } from './types';
import { generateUniqueId } from './utils';

/**
 * 查找包含特定 mcp id 的工具调用消息
 */
export const findExistingToolMessage = (messages: MessageType[], id: string) => {
  return messages.find(
    (msg) =>
      msg.role === 'assistant' &&
      msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.data?.some((tool: any) => tool.id === id)),
  );
};

/**
 * 查找进行中的工具调用消息
 */
export const findProgressToolMessage = (messages: MessageType[]) => {
  return messages.find((msg) => msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress'));
};

/**
 * 从消息中提取工具调用数据
 */
export const extractToolCallData = (contentList: any[], id?: string) => {
  let toolCallResults: any[] = [];
  let mcpContentIndex = -1;

  if (id) {
    mcpContentIndex = contentList.findIndex((content) => content.type === 'mcp' && typeof content.content === 'object' && content.content.data?.some((tool: any) => tool.id === id));
  } else {
    mcpContentIndex = contentList.findIndex((content) => content.type === 'mcp' && typeof content.content === 'object' && content.content.status === 'progress');
  }

  if (mcpContentIndex >= 0) {
    toolCallResults = [...contentList[mcpContentIndex].content.data];
  }

  return { toolCallResults, mcpContentIndex };
};

/**
 * 创建 MCP 内容对象
 */
export const createMcpContent = (status: string, data: any[], totalDuration?: number) => {
  return {
    id: generateUniqueId('content'),
    type: 'mcp' as const,
    content: {
      status,
      data,
      totalDuration,
    },
  };
};

/**
 * 更新内容列表中的 MCP 项
 */
export const updateContentListWithMcp = (contentList: any[], mcpContent: any) => {
  const updatedList = [...contentList];
  const mcpIndex = updatedList.findIndex((item) => item.type === 'mcp');

  if (mcpIndex >= 0) {
    updatedList[mcpIndex] = mcpContent;
  } else {
    updatedList.push(mcpContent);
  }

  return updatedList;
};

/**
 * 构建工具调用数据对象
 */
export const buildToolCallData = (toolResponse: any, toolCall: IToolCall, toolGroupId: string, status: 'progress' | 'success' | 'error', outputParams?: string): IToolCallData => {
  if (!toolGroupId) return {} as IToolCallData;
  return {
    id: toolGroupId,
    name: toolResponse.toolName,
    desc: toolResponse.toolDesc || '工具调用',
    logo: toolResponse.toolLogo || '',
    inputParams: JSON.stringify(toolResponse.toolArgs),
    outputParams: outputParams || '',
    status,
  };
};

/**
 * 处理工具调用错误时的消息更新
 */
export const handleToolCallErrorMessage = (
  toolCallResults: any[],
  currentContentList: any[],
  currentToolMessageId: string,
  errorMessage: string,
  addMessage: (msg: MessageType, isUpdate?: boolean) => string,
) => {
  if (toolCallResults.length > 0) {
    // 更新最后一个工具的状态为错误
    const lastIndex = toolCallResults.length - 1;
    toolCallResults[lastIndex] = {
      ...toolCallResults[lastIndex],
      outputParams: errorMessage,
      status: 'error',
    };

    // 创建错误状态的 MCP 内容
    const errorMcpContent = createMcpContent(toolCallResults.every((t) => t.status === 'error' || t.status === 'success') ? 'error' : 'progress', toolCallResults);

    // 更新内容列表
    const errorContentList = updateContentListWithMcp(currentContentList, errorMcpContent);

    // 构建错误消息
    const errorMsg: MessageType = {
      id: currentToolMessageId || generateUniqueId('mcp_error_msg'),
      role: 'assistant',
      contentList: errorContentList,
    };

    addMessage(errorMsg, !!currentToolMessageId);
  } else {
    // 创建新的错误消息
    const mcpErrorMessage: MessageType = {
      id: generateUniqueId('mcp_error_msg'),
      role: 'assistant',
      contentList: [
        createMcpContent('error', [
          {
            id: '',
            name: '',
            desc: '工具调用',
            logo: '',
            inputParams: '',
            outputParams: errorMessage,
            status: 'error',
          },
        ]),
      ],
    };

    addMessage(mcpErrorMessage);
  }
};
