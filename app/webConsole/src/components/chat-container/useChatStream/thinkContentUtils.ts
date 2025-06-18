import { generateUniqueId } from './utils';
import { MessageType } from '@res-utiles/ui-components';

interface ThinkContentItem {
  id: string;
  type: 'think';
  content: {
    status: 'complete';
    data: string;
  };
}

interface PlainContentItem {
  id: string;
  type: 'plain';
  content: string;
}

/**
 * 处理流式响应数据，返回累积的内容
 * @param data 响应数据
 * @param currentResponseContent 当前累积的响应内容
 * @returns 更新后的响应内容
 */
export const accumulateResponseContent = (data: any, currentResponseContent: string): string => {
  let responseContent = currentResponseContent;

  // 根据不同的数据类型处理内容累积
  if (data.is_complete) {
    responseContent = data.content;
  } else if (data.type === 'answer') {
    if (responseContent.length === 0) {
      responseContent = data.content;
    } else {
      responseContent += data.content;
    }
  } else {
    if (data.content.length > responseContent.length || !responseContent.includes(data.content.trim())) {
      responseContent += data.content;
    } else {
      responseContent = data.content;
    }
  }

  return responseContent;
};

/**
 * 提取并处理完整的 think 标签
 * @param responseContent 包含 think 标签的完整响应内容
 * @param processedThinkIds 已处理的 think 内容标识集合
 * @param addMessage 添加消息的函数
 * @returns 已处理的 think 内容标识集合
 */
export const processCompletedThinkTags = (responseContent: string, processedThinkIds: Set<string>, addMessage: (message: MessageType, isUpdate?: boolean) => string): Set<string> => {
  const updatedProcessedIds = new Set(processedThinkIds);

  // 正则表达式匹配所有完整的 think 标签
  const completedThinkRegex = /<think>([\s\S]*?)<\/think>/g;
  let match;

  while ((match = completedThinkRegex.exec(responseContent)) !== null) {
    const thinkContent = match[1].trim();

    // 为每个 think 内容生成唯一标识（基于内容和位置）
    const thinkId = `${match.index}-${thinkContent.length}`;

    // 只处理尚未处理过的 think 内容
    if (thinkContent && !updatedProcessedIds.has(thinkId)) {
      const thinkMessage: MessageType = {
        id: generateUniqueId('think_msg'),
        role: 'assistant',
        contentList: [
          {
            id: generateUniqueId('content'),
            type: 'think' as const,
            content: {
              status: 'complete',
              data: thinkContent,
            },
          },
        ],
      };

      addMessage(thinkMessage);
      updatedProcessedIds.add(thinkId);
    }
  }

  return updatedProcessedIds;
};

/**
 * 从完整内容中提取 plain 内容（不包含 think 标签）
 * @param fullContent 包含 think 标签的完整内容
 * @returns 仅包含 plain 内容的字符串
 */
export const extractPlainContent = (fullContent: string): string => {
  // 移除所有 think 标签及其内容
  return fullContent.replace(/<think>[\s\S]*?<\/think>/g, '').trim();
};

/**
 * 构建最终的纯文本消息（不包含 think 内容）
 * @param responseContent 包含 think 标签的响应内容
 * @returns 仅包含 plain 内容的消息
 */
export const buildPlainMessage = (responseContent: string): MessageType => {
  const plainContent = extractPlainContent(responseContent);

  if (!plainContent) {
    // 如果没有 plain 内容，返回空消息
    return {
      id: generateUniqueId('ai_msg'),
      role: 'assistant',
      contentList: [],
    };
  }

  return {
    id: generateUniqueId('ai_msg'),
    role: 'assistant',
    contentList: [
      {
        id: generateUniqueId('content'),
        type: 'plain' as const,
        content: plainContent,
      },
    ],
  };
};

/**
 * 解析包含 <think> 标签的内容（用于非流式响应）
 * @param content 原始内容
 * @param addMessage 添加消息的函数
 */
export const parseAndAddThinkMessages = (content: string, addMessage: (message: MessageType, isUpdate?: boolean) => string): void => {
  const processedIds = new Set<string>();
  processCompletedThinkTags(content, processedIds, addMessage);
};
