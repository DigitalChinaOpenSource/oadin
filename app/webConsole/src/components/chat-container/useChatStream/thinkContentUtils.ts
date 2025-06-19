import { generateUniqueId } from './utils';
import { MessageType } from '@res-utiles/ui-components';

interface ParsedContent {
  type: 'think' | 'plain';
  content: string | { data: string; status: 'progress' | 'success' | 'error' };
}
/**
 * 解析包含 <think> 标签的内容
 * @param content 原始内容
 * @returns 解析后的内容数组
 */
export const parseThinkContent = (content: string): Array<ParsedContent> => {
  const result: Array<ParsedContent> = [];

  // 正则表达式匹配 <think> 和 </think> 标签
  const thinkRegex = /<think>([\s\S]*?)<\/think>/g;
  let lastIndex = 0;
  let match;

  while ((match = thinkRegex.exec(content)) !== null) {
    // 添加 <think> 标签之前的内容
    if (match.index > lastIndex) {
      const beforeContent = content.substring(lastIndex, match.index).trim();
      if (beforeContent) {
        result.push({
          type: 'plain',
          content: beforeContent,
        });
      }
    }

    // 添加 <think> 标签内的内容，使用新格式
    const thinkContent = match[1].trim();
    if (thinkContent) {
      result.push({
        type: 'think',
        content: {
          data: thinkContent,
          status: 'success',
        },
      });
    }

    lastIndex = match.index + match[0].length;
  }

  // 添加最后一个 </think> 之后的内容
  if (lastIndex < content.length) {
    const afterContent = content.substring(lastIndex).trim();
    if (afterContent) {
      result.push({
        type: 'plain',
        content: afterContent,
      });
    }
  }

  // 如果没有找到 think 标签，返回原始内容
  if (result.length === 0 && content.trim()) {
    result.push({
      type: 'plain',
      content: content,
    });
  }

  return result;
};

/**
 * 构建包含 think 内容的消息
 * @param responseContent 响应内容
 * @param isComplete 是否完成
 * @returns MessageType 对象
 */
export const buildMessageWithThinkContent = (responseContent: string, isComplete: boolean = false): MessageType => {
  const parsedContents = parseThinkContent(responseContent);

  const contentList = parsedContents.map((item) => ({
    id: generateUniqueId('content'),
    type: item.type,
    content: item.content,
  }));

  return {
    id: generateUniqueId('ai_msg'),
    role: 'assistant',
    contentList,
  };
};

/**
 * 处理包含 think 标签的文本内容
 * @param data 响应数据
 * @param currentResponseContent 当前累积的响应内容
 * @param setStreamingContent 设置流式内容的函数
 * @param setStreamingThinking 设置流式思考内容的函数
 * @param requestStateRef 请求状态引用
 * @returns 更新后的响应内容
 */
export const handleTextContent = (
  data: any,
  currentResponseContent: string,
  setStreamingContent: (content: string) => void,
  setStreamingThinking: (content: string) => void,
  requestStateRef: any,
): string => {
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

  // 检查是否有未闭合的 <think> 标签
  const openTagCount = (responseContent.match(/<think>/g) || []).length;
  const closeTagCount = (responseContent.match(/<\/think>/g) || []).length;
  const hasUnfinishedThink = openTagCount > closeTagCount;

  // 解析内容，处理完整的 <think>...</think> 块
  const parsedContents = parseThinkContent(responseContent);

  if (hasUnfinishedThink) {
    // 找到最后一个未闭合的 <think> 标签内容
    const lastThinkMatch = responseContent.lastIndexOf('<think>');
    if (lastThinkMatch !== -1) {
      const thinkContent = responseContent.substring(lastThinkMatch + 7); // +7 是 <think> 的长度
      if (thinkContent.trim()) {
        // 将未完成的思考内容加入解析结果，并标记为 progress 状态
        parsedContents.push({
          type: 'think',
          content: {
            data: thinkContent,
            status: 'progress', // 未闭合标签使用 progress 状态
          },
        });
      }
    }
  }

  if (responseContent.includes('<think>') || hasUnfinishedThink) {
    // 从解析内容中提取 thinking 文本
    const thinkContents = parsedContents
      .filter((item) => item.type === 'think')
      .map((item) => {
        if (typeof item.content === 'object') {
          return item.content.data;
        }
        return '';
      })
      .join('\n\n');

    // 更新思考内容
    setStreamingThinking(thinkContents);
    if (requestStateRef && requestStateRef.current) {
      requestStateRef.current.content.thinking = thinkContents;
    }

    // 提取纯文本内容
    const plainContents = parsedContents
      .filter((item) => item.type === 'plain')
      .map((item) => (typeof item.content === 'string' ? item.content : ''))
      .join('\n\n');

    setStreamingContent(plainContents);
  } else {
    // 没有 think 标签，正常设置
    setStreamingContent(responseContent);
  }

  return responseContent;
};
