import { generateUniqueId } from './utils';
import { MessageType } from '@res-utiles/ui-components';

interface ParsedContent {
  type: 'think' | 'plain';
  content: string | { data: string; status: 'progress' | 'success' | 'error' };
}

interface StreamData {
  content?: string;
  thinking?: string;
  is_complete?: boolean;
  type?: string;
}

interface RequestState {
  current: {
    content: {
      response: string;
      thinking: string;
      // 新增：thinking 字段的累加状态
      thinkingFromField: string;
      isThinkingActive: boolean;
    };
    [key: string]: unknown;
  };
}

/**
 * 解析包含 <think> 标签的内容
 * @param content 原始内容
 * @param hasUnfinishedThink 是否有未闭合的 think 标签
 * @returns 解析后的内容数组
 */
export const parseThinkContent = (content: string, hasUnfinishedThink: boolean = false): Array<ParsedContent> => {
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

  // 处理未闭合的 <think> 标签
  if (hasUnfinishedThink) {
    const lastThinkIndex = content.lastIndexOf('<think>');
    if (lastThinkIndex !== -1 && lastThinkIndex >= lastIndex) {
      // 添加最后一个完整标签到未闭合标签之间的内容（如果有的话）
      if (lastThinkIndex > lastIndex) {
        const betweenContent = content.substring(lastIndex, lastThinkIndex).trim();
        if (betweenContent) {
          result.push({
            type: 'plain',
            content: betweenContent,
          });
        }
      }

      // 添加未闭合的 think 内容
      const unfinishedThinkContent = content.substring(lastThinkIndex + 7).trim(); // +7 是 <think> 的长度
      if (unfinishedThinkContent) {
        result.push({
          type: 'think',
          content: {
            data: unfinishedThinkContent,
            status: 'error',
          },
        });
      }
      // 更新 lastIndex 到内容末尾，避免重复处理
      lastIndex = content.length;
    }
  }

  // 添加最后一个标签之后的内容（只有在没有未闭合标签的情况下）
  if (!hasUnfinishedThink && lastIndex < content.length) {
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
 * @param thinkingContent 额外的思考内容（来自 thinking 字段或 <think> 标签）
 * @param thinkingFromField 来自 thinking 字段的思考内容
 * @returns MessageType 对象
 */
export const buildMessageWithThinkContent = (responseContent: string, isComplete: boolean = false, thinkingContent?: string, thinkingFromField?: string): MessageType => {
  // 检查是否有未闭合的 <think> 标签
  const openTagCount = (responseContent.match(/<think>/g) || []).length;
  const closeTagCount = (responseContent.match(/<\/think>/g) || []).length;
  const hasUnfinishedThink = openTagCount > closeTagCount && !isComplete;

  const parsedContents = parseThinkContent(responseContent, hasUnfinishedThink);

  // 如果有来自 thinking 字段的思考内容，添加到解析结果的开头
  if (thinkingFromField && thinkingFromField.trim()) {
    parsedContents.unshift({
      type: 'think',
      content: {
        data: thinkingFromField.trim(),
        status: 'success',
      },
    });
  }

  // 如果有额外的思考内容（向后兼容），添加到解析结果中
  if (thinkingContent && thinkingContent.trim() && thinkingContent !== thinkingFromField) {
    parsedContents.push({
      type: 'think',
      content: {
        data: thinkingContent.trim(),
        status: 'success',
      },
    });
  }

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
 * 处理包含 think 标签或 thinking 字段的文本内容
 * @param data 响应数据
 * @param currentResponseContent 当前累积的响应内容
 * @param setStreamingContent 设置流式内容的函数
 * @param setStreamingThinking 设置流式思考内容的函数
 * @param requestStateRef 请求状态引用
 * @returns 更新后的响应内容
 */
export const handleTextContent = (
  data: StreamData,
  currentResponseContent: string,
  setStreamingContent: (content: string) => void,
  setStreamingThinking: (content: string) => void,
  requestStateRef: RequestState,
): string => {
  let responseContent = currentResponseContent;

  // 处理 thinking 字段的深度思考
  if (data.thinking !== undefined) {
    const isThinkingEnd = data.thinking === '' || data.thinking === null;

    if (!requestStateRef.current.content.isThinkingActive && !isThinkingEnd) {
      // 开始深度思考
      requestStateRef.current.content.isThinkingActive = true;
      requestStateRef.current.content.thinkingFromField = data.thinking;
    } else if (requestStateRef.current.content.isThinkingActive && !isThinkingEnd) {
      // 累加深度思考内容
      requestStateRef.current.content.thinkingFromField += data.thinking;
    } else if (requestStateRef.current.content.isThinkingActive && isThinkingEnd) {
      // 结束深度思考 - 保持累加的内容，但标记为结束
      requestStateRef.current.content.isThinkingActive = false;
    }

    // 更新流式思考显示内容
    let displayThinkingContent = requestStateRef.current.content.thinkingFromField || '';

    // 如果还有其他来源的思考内容，合并显示
    if (requestStateRef.current.content.thinking) {
      displayThinkingContent = displayThinkingContent ? `${displayThinkingContent}\n\n${requestStateRef.current.content.thinking}` : requestStateRef.current.content.thinking;
    }

    setStreamingThinking(displayThinkingContent);

    // 如果同时也有 content 字段，作为普通内容处理
    if (data.content) {
      setStreamingContent(data.content);
      return data.content;
    }
    // 如果只有 thinking 字段，保持当前内容不变
    return currentResponseContent;
  }

  // 根据不同的数据类型处理内容累积
  if (data.is_complete) {
    responseContent = data.content || '';
  } else if (data.type === 'answer') {
    if (responseContent.length === 0) {
      responseContent = data.content || '';
    } else {
      responseContent += data.content || '';
    }
  } else {
    if (data.content && (data.content.length > responseContent.length || !responseContent.includes(data.content.trim()))) {
      responseContent += data.content || '';
    } else if (data.content) {
      responseContent = data.content;
    }
  }

  // 检查是否有未闭合的 <think> 标签
  const openTagCount = (responseContent.match(/<think>/g) || []).length;
  const closeTagCount = (responseContent.match(/<\/think>/g) || []).length;
  const hasUnfinishedThink = openTagCount > closeTagCount;

  // 解析内容，统一处理完整和未完整的 <think> 块
  const parsedContents = parseThinkContent(responseContent, hasUnfinishedThink);

  if (responseContent.includes('<think>')) {
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

    // 更新来自 <think> 标签的思考内容
    requestStateRef.current.content.thinking = thinkContents;

    // 合并所有思考内容进行显示
    let displayThinkingContent = '';
    if (requestStateRef.current.content.thinkingFromField) {
      displayThinkingContent = requestStateRef.current.content.thinkingFromField;
    }
    if (thinkContents) {
      displayThinkingContent = displayThinkingContent ? `${displayThinkingContent}\n\n${thinkContents}` : thinkContents;
    }

    setStreamingThinking(displayThinkingContent);

    // 提取纯文本内容
    const plainContents = parsedContents
      .filter((item) => item.type === 'plain')
      .map((item) => (typeof item.content === 'string' ? item.content : ''))
      .join('\n\n');

    setStreamingContent(plainContents);
  } else {
    // 没有 <think> 标签时，只更新来自 <think> 标签的思考内容为空
    requestStateRef.current.content.thinking = '';

    // 但仍需要显示来自 thinking 字段的内容
    if (requestStateRef.current.content.thinkingFromField) {
      setStreamingThinking(requestStateRef.current.content.thinkingFromField);
    }

    // 正常设置响应内容
    setStreamingContent(responseContent);
  }

  return responseContent;
};
