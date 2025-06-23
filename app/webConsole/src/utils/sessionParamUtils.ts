/**
 * 会话ID URL参数处理工具
 * 用于从URL中获取和设置会话ID
 */

/**
 * 从URL中获取会话ID
 * @returns 当前URL中的会话ID，如果不存在则返回空字符串
 */
export const getSessionIdFromUrl = (): string => {
  const urlParams = new URLSearchParams(window.location.search);
  return urlParams.get('sessionId') || '';
};

/**
 * 设置会话ID到URL中
 * @param sessionId 要设置的会话ID
 */
export const setSessionIdToUrl = (sessionId: string): void => {
  const url = new URL(window.location.href);
  if (sessionId) {
    url.searchParams.set('sessionId', sessionId);
  } else {
    url.searchParams.delete('sessionId');
  }
  // 使用history.replaceState来更新URL而不重新加载页面
  window.history.replaceState({}, '', url);
};
