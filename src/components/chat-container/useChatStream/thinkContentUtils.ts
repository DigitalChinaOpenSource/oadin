// utils/thinkContentUtils.ts

import { generateUniqueId } from './utils';
import { MessageType } from '@res-utiles/ui-components';

/**
 * 解析包含 <think> 标签的内容
 * @param content 原始内容
 * @returns 解析后的内容数组
 */
export const parseThinkContent = (content: string): Array<{ type: 'think' | 'plain'; content: string }> => {
  const result: Array<{ type: 'think' | 'plain'; content: string }> = [];

  // 正则表达式匹配 <think> 标签
  const thinkRegex = /<think>([\s\S]*?)<\/think>/g;
  let lastIndex = 0;
  let match;

  while ((match = thinkRegex.exec(content)) !== null) {
    // 添加 <think> 标签之前的内容（如果有）
    if (match.index > lastIndex) {
      const beforeContent = content.substring(lastIndex, match.index).trim();
      if (beforeContent) {
        result.push({
          type: 'plain',
          content: beforeContent,
        });
      }
    }

    // 添加 <think> 标签内的内容
    const thinkContent = match[1].trim();
    if (thinkContent) {
      result.push({
        type: 'think',
        content: thinkContent,
      });
    }

    lastIndex = match.index + match[0].length;
  }

  // 添加最后一个 </think> 之后的内容（如果有）
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

  // 检查是否包含 think 标签
  if (responseContent.includes('<think>')) {
    // 解析完整内容
    const parsedContents = parseThinkContent(responseContent);

    // 累积 thinking 内容（只包含 think 标签内的内容）
    const thinkContents = parsedContents
      .filter((item) => item.type === 'think')
      .map((item) => item.content)
      .join('\n\n');

    // 更新 thinking 内容的流式显示
    setStreamingThinking(thinkContents);
    if (requestStateRef) {
      requestStateRef.current.thinkingContent = thinkContents;
    }

    // 累积显示内容（只包含 plain 内容）
    const plainContents = parsedContents
      .filter((item) => item.type === 'plain')
      .map((item) => item.content)
      .join('\n\n');

    setStreamingContent(plainContents);
  } else {
    // 没有 think 标签，正常设置
    setStreamingContent(responseContent);
  }

  return responseContent;
};
