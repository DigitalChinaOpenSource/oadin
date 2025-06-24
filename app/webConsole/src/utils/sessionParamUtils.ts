/**
 * 会话ID URL参数处理工具
 * 用于从URL中获取和设置会话ID
 */

/**
 * 从URL中获取会话ID，如果URL中不存在则尝试从sessionStorage中获取
 * @returns 当前会话ID，如果不存在则返回空字符串
 */
export const getSessionIdFromUrl = (): string => {
  const urlParams = new URLSearchParams(window.location.search);
  const sessionIdFromUrl = urlParams.get('sessionId') || '';

  // 如果URL中没有sessionId，则尝试从sessionStorage获取
  if (!sessionIdFromUrl) {
    return sessionStorage.getItem('currentSessionId') || '';
  }

  return sessionIdFromUrl;
};

/**
 * 设置会话ID到URL中并保存到sessionStorage
 * @param sessionId 要设置的会话ID
 */
export const setSessionIdToUrl = (sessionId: string): void => {
  const url = new URL(window.location.href);
  if (sessionId) {
    url.searchParams.set('sessionId', sessionId);
    // 同时保存到sessionStorage
    sessionStorage.setItem('currentSessionId', sessionId);
  } else {
    url.searchParams.delete('sessionId');
    // 清除sessionStorage中的会话ID
    sessionStorage.removeItem('currentSessionId');
  }
  // 使用history.replaceState来更新URL而不重新加载页面
  window.history.replaceState({}, '', url);
};

/**
 * 保存当前会话ID到sessionStorage中
 * 用于在页面跳转时保存会话状态
 */
export const saveSessionIdToStorage = (): void => {
  const sessionId = getSessionIdFromUrl();
  if (sessionId) {
    sessionStorage.setItem('currentSessionId', sessionId);
  }
};

/**
 * 清除保存在sessionStorage中的会话ID
 */
export const clearSessionIdFromStorage = (): void => {
  sessionStorage.removeItem('currentSessionId');
};
