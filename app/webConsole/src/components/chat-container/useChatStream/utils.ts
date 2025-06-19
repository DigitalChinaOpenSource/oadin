// 生成唯一ID的工具函数
export function generateUniqueId(prefix: string = 'id'): string {
  return `${prefix}_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

// 复制消息到剪贴板
export const copyMessageToClipboard = (content: string) => {
  if (!content) return false;

  try {
    navigator.clipboard.writeText(content);
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
