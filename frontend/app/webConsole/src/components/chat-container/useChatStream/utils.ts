import { httpRequest } from '@/utils/httpRequest';
import { getSessionIdFromUrl } from '@/utils/sessionParamUtils';
// 生成唯一ID的工具函数
export function generateUniqueId(prefix: string = 'id'): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

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
    await httpRequest.post(
      '/playground/session/genTitle',
      {
        sessionId: currentSessionId,
      },
      { timeout: 300000 },
    );
  } catch (error) {
    console.error('生成标题失败===>');
  }
};
