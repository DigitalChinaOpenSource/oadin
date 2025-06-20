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

export type ContentItem = PlainContent | ThinkContent;

export interface OutputMessage {
  id: string;
  role: 'user' | 'assistant';
  contentList: ContentItem[];
}
/**
 *
 * @param inputArray 从接口获取输入的数组
 * @returns 输出的数组用于渲染
 */
export function convertMessageFormat(inputArray: InputMessage[]): OutputMessage[] {
  return inputArray.map((item: InputMessage): OutputMessage => {
    const result: OutputMessage = {
      id: item.id, // 直接使用原始数据的 id
      role: item.role,
      contentList: [],
    };

    const content: string = item.content;

    // 检查是否包含 <think> 标签
    const thinkRegex: RegExp = /<think>([\s\S]*?)<\/think>/;
    const match: RegExpMatchArray | null = content.match(thinkRegex);

    if (match) {
      // 包含 think 标签的情况
      const beforeThink: string = content.substring(0, match.index!).trim();
      const thinkContent: string = match[1].trim();
      const afterThink: string = content.substring(match.index! + match[0].length).trim();

      // 添加 think 前的内容（如果存在）
      if (beforeThink) {
        result.contentList.push({
          id: generateUniqueId('content'),
          type: 'plain',
          content: beforeThink,
        } as PlainContent);
      }

      // 添加 think 内容
      result.contentList.push({
        id: generateUniqueId('content'),
        type: 'think',
        content: {
          data: thinkContent,
          status: 'success',
        },
      } as ThinkContent);

      // 添加 think 后的内容（如果存在）
      if (afterThink) {
        result.contentList.push({
          id: generateUniqueId('content'),
          type: 'plain',
          content: afterThink,
        } as PlainContent);
      }
    } else {
      // 不包含 think 标签，直接作为 plain 类型
      result.contentList.push({
        id: generateUniqueId('content'),
        type: 'plain',
        content: content,
      } as PlainContent);
    }

    return result;
  });
}
