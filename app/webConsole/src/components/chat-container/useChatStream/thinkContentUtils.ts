import { generateUniqueId } from './utils';
import { ChatMessageItem } from '@res-utiles/ui-components';

interface ParsedContent {
  type: 'think' | 'plain';
  content: string | { data: string; status: 'progress' | 'success' | 'error' };
}

interface StreamData {
  content?: string;
  thoughts?: string;
  is_complete?: boolean;
  type?: string;
}

interface RequestState {
  current: {
    content: {
      response: string;
      thinking: string;
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
 * @param thinkingContent 来自 <think> 标签的思考内容
 * @param thinkingFromField 来自 thinking 字段的思考内容
 * @param isThinkingActive thinking 字段是否还在活跃状态
 * @returns ChatMessageItem 对象
 */
export const buildMessageWithThinkContent = (
  responseContent: string,
  isComplete: boolean = false,
  thinkingContent?: string | { data: string; status: string },
  thinkingFromField?: string,
  isThinkingActive: boolean = false,
): ChatMessageItem => {
  // 检查是否有未闭合的 <think> 标签
  const openTagCount = (responseContent.match(/<think>/g) || []).length;
  const closeTagCount = (responseContent.match(/<\/think>/g) || []).length;
  const hasUnfinishedThink = openTagCount > closeTagCount && !isComplete;
  const parsedContents = parseThinkContent(responseContent, hasUnfinishedThink);
  // 只处理来自 thinking 字段的思考内容（两种方式互斥）
  if (thinkingFromField && thinkingFromField.trim()) {
    // 根据是否还在思考中来设置状态
    const thinkingStatus = isThinkingActive ? 'progress' : isComplete ? 'success' : 'error';

    parsedContents.unshift({
      type: 'think',
      content: {
        data: thinkingFromField.trim(),
        status: thinkingStatus,
      },
    });
  }
  // 注意：这里不再处理 thinkingContent，因为它来自 <think> 标签
  // <think> 标签的内容已经在 parseThinkContent 中被处理了

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
  setStreamingThinking: (content: string | { data: string; status: string }) => void,
  requestStateRef: RequestState,
  isUserCancelled: boolean = false, // 新增参数：标识是否为用户主动取消
): string => {
  let responseContent = currentResponseContent;

  // 处理 thinking 字段的深度思考（与 <think> 标签互斥）
  if (data.thoughts !== undefined) {
    // 当content没值且thoughts有值，表示深度思考开始或继续
    // 当content有值thoughts没值或不存在时，表示深度思考结束
    const hasThoughts = data.thoughts !== '' && data.thoughts !== null;
    const hasContent = data.content !== undefined && data.content !== '' && data.content !== null;

    if (hasThoughts) {
      // 深度思考正在进行
      if (!requestStateRef.current.content.isThinkingActive) {
        // 开始深度思考
        requestStateRef.current.content.isThinkingActive = true;
        requestStateRef.current.content.thinkingFromField = data.thoughts;
      } else {
        // 累加深度思考内容
        requestStateRef.current.content.thinkingFromField += data.thoughts;
      }

      // 更新流式思考显示内容
      setStreamingThinking({
        data: requestStateRef.current.content.thinkingFromField || '',
        status: 'progress',
      });

      // 如果同时有content，则处理content
      if (hasContent) {
        setStreamingContent(data.content || '');
        responseContent = data.content || '';
      }
    } else if (requestStateRef.current.content.isThinkingActive) {
      // 结束深度思考 - 保持累加的内容，但标记为结束
      requestStateRef.current.content.isThinkingActive = false;

      // 如果有content，则更新内容
      if (hasContent) {
        setStreamingContent(data.content || '');
        responseContent = data.content || '';
      }
    }

    return responseContent;
  } else if (requestStateRef.current.content.isThinkingActive && data.content) {
    // thoughts 从有值变为 undefined，且 content 从无值变为有值
    // 结束深度思考并设置状态
    const thinkStatus = isUserCancelled ? 'error' : 'success';

    // 将思考内容状态设置为 success 或 error
    if (requestStateRef.current.content.thinkingFromField) {
      setStreamingThinking({
        data: requestStateRef.current.content.thinkingFromField,
        status: thinkStatus,
      });
    }

    // 更新状态
    requestStateRef.current.content.isThinkingActive = false;

    // 设置内容
    setStreamingContent(data.content);
    responseContent = data.content;

    return responseContent;
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

  if (responseContent.includes('<think>')) {
    // 检查是否有未闭合的 <think> 标签
    const openTagCount = (responseContent.match(/<think>/g) || []).length;
    const closeTagCount = (responseContent.match(/<\/think>/g) || []).length;
    const hasUnfinishedThink = openTagCount > closeTagCount;

    const parsedContents = parseThinkContent(responseContent, hasUnfinishedThink);
    const thinkContents = parsedContents
      .filter((item) => item.type === 'think')
      .map((item) => {
        if (typeof item.content === 'object') {
          return item.content.data;
        }
        return '';
      })
      .join('\n\n');

    requestStateRef.current.content.thinking = thinkContents;
    setStreamingThinking({
      data: thinkContents,
      status: 'progress',
    });

    // 提取纯文本内容
    const plainContents = parsedContents
      .filter((item) => item.type === 'plain')
      .map((item) => (typeof item.content === 'string' ? item.content : ''))
      .join('\n\n');

    setStreamingContent(plainContents);
  } else {
    requestStateRef.current.content.thinking = '';
    if (requestStateRef.current.content.thinkingFromField) {
      setStreamingThinking({
        data: requestStateRef.current.content.thinkingFromField,
        status: 'success',
      });
    } else {
      setStreamingThinking({
        data: '',
        status: 'success',
      });
    }

    setStreamingContent(responseContent);
  }

  return responseContent;
};
