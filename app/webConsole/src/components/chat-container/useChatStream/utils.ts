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
