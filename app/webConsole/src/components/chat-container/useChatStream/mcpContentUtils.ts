import { MessageType } from '@res-utiles/ui-components';
import { IToolCallData } from './types';
import { generateUniqueId } from './utils';

/**
 * 查找进行中的工具调用消息
 */
export const findProgressToolMessage = (messages: MessageType[]) => {
  return messages.find((msg) => msg.contentList?.some((contentItem: any) => contentItem.type === 'mcp' && typeof contentItem.content === 'object' && contentItem.content.status === 'progress'));
};

/**
 * 从消息中提取指定 tool_group_id 的工具调用数据
 */
export const extractToolCallDataByGroupId = (contentList: any[], toolGroupId: string) => {
  let toolCallResults: any[] = [];
  let mcpContentIndex = -1;

  // 查找指定 tool_group_id 对应的 contentList 项，使用 content.groupId 字段
  mcpContentIndex = contentList.findIndex((content) => content.type === 'mcp' && typeof content.content === 'object' && content.content.groupId === toolGroupId);

  if (mcpContentIndex >= 0) {
    toolCallResults = [...(contentList[mcpContentIndex].content.data || [])];
  }

  return { toolCallResults, mcpContentIndex };
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
export const createMcpContentWithGroupId = (toolGroupId: string, status: string, data: any[], totalDuration?: number) => {
  return {
    id: generateUniqueId('mcp_content'),
    type: 'mcp' as const,
    content: {
      groupId: toolGroupId, // 添加 groupId 字段用于标识和分组
      status,
      data,
      totalDuration,
    },
  };
};

/**
 * 更新内容列表中指定 tool_group_id 的 MCP 项
 */
export const updateContentListWithMcpByGroupId = (contentList: any[], mcpContent: any, toolGroupId: string) => {
  const updatedList = [...contentList];
  // 使用 tool_group_id 作为唯一标识查找 MCP 内容项
  const mcpIndex = updatedList.findIndex((item) => item.type === 'mcp' && item.content && item.content.groupId === toolGroupId);

  if (mcpIndex >= 0) {
    // 更新现有项
    updatedList[mcpIndex] = {
      ...updatedList[mcpIndex],
      content: {
        ...mcpContent.content,
        groupId: toolGroupId, // 保留 groupId 以便后续查找
      },
    };
  } else {
    // 添加新项，并附加 groupId 属性
    updatedList.push({
      ...mcpContent,
      content: {
        ...mcpContent.content,
        groupId: toolGroupId, // 添加 groupId 属性
      },
    });
  }

  return updatedList;
};

/**
 * 构建工具调用数据对象
 */
export const buildToolCallData = (toolResponse: any, mcpId: string, status: 'progress' | 'success' | 'error', outputParams?: string): IToolCallData => {
  return {
    id: mcpId,
    name: toolResponse.toolName,
    desc: toolResponse.desc || '',
    logo: toolResponse.logo || '',
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
  toolGroupId: string,
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

    // 查找原有内容中的 totalDuration，使用 content.groupId 字段
    const existingContent = currentContentList.find((item) => item.type === 'mcp' && item.content?.groupId === toolGroupId)?.content || {};
    const existingTotalDuration = existingContent.totalDuration || 0;

    // 创建错误状态的 MCP 内容
    const errorMcpContent = createMcpContentWithGroupId(
      toolGroupId,
      toolCallResults.every((t) => t.status === 'error' || t.status === 'success') ? 'error' : 'progress',
      toolCallResults,
      existingTotalDuration,
    );

    // 更新内容列表
    const errorContentList = updateContentListWithMcpByGroupId(currentContentList, errorMcpContent, toolGroupId);

    // 构建错误消息，使用 toolGroupId 作为消息 ID
    const errorMsg: MessageType = {
      id: toolGroupId || generateUniqueId('mcp_error_msg'),
      role: 'assistant',
      contentList: errorContentList,
    };

    addMessage(errorMsg, !!toolGroupId);
  } else {
    // 创建新的错误消息，使用 toolGroupId 作为消息 ID
    const mcpErrorMessage: MessageType = {
      id: toolGroupId || generateUniqueId('mcp_error_msg'),
      role: 'assistant',
      contentList: [
        createMcpContentWithGroupId(
          toolGroupId,
          'error',
          [
            {
              id: generateUniqueId('error_tool'),
              name: '',
              desc: '工具调用',
              logo: '',
              inputParams: '',
              outputParams: errorMessage,
              status: 'error',
            },
          ],
          0,
        ),
      ],
    };

    addMessage(mcpErrorMessage);
  }
};
