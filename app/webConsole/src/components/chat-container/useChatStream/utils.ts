import useChatStore from '@/components/chat-container/store/useChatStore';
import { httpRequest } from '@/utils/httpRequest';
import { getSessionIdFromUrl } from '@/utils/sessionParamUtils';
// 生成唯一ID的工具函数
export function generateUniqueId(prefix: string = 'id'): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

/**
 * 复制消息到剪贴板
 * 复制最后一条用户消息及之后的所有对话，转换为JSON格式
 * @param content 可选参数，如果提供则直接复制该内容，否则从聊天存储获取
 * @returns 复制是否成功
 */
export const copyMessageToClipboard = (content?: string) => {
  try {
    // 如果提供了具体内容，直接复制
    if (content) {
      navigator.clipboard.writeText(content);
      return true;
    }

    // 否则，从聊天存储获取最后一条用户消息及之后的对话
    const messages = useChatStore.getState().messages;
    const lastMessage = messages[messages.length - 1];
    const relevantContent = lastMessage?.contentList?.[lastMessage.contentList.length - 1]?.content || '';

    // 转换为JSON格式
    const jsonData = JSON.stringify(relevantContent, null, 2);

    // 复制到剪贴板
    navigator.clipboard.writeText(jsonData);
    return true;
  } catch (err) {
    console.error('复制到剪贴板失败:', err);
    return false;
  }
};

/**
 * 格式化错误消息
 * @param template 包含占位符的消息模板
 * @param args 要替换的参数
 */
export const formatErrorMessage = (template: string, ...args: any[]): string => {
  return template.replace(/{(\d+)}/g, (match, index) => {
    return args[index] !== undefined ? String(args[index]) : match;
  });
};

export const fetchGenChatTitle = async () => {
  const currentSessionId = getSessionIdFromUrl();
  if (!currentSessionId) return;

  try {
    await httpRequest.post('/playground/session/genTitle', {
      sessionId: currentSessionId,
    });
  } catch (error) {
    console.error('生成标题失败===>');
  }
};
