import { AUTH_TOKEN } from '../constants';

export function baseHeaders(providedToken = null) {
  const token = providedToken || window.localStorage.getItem(AUTH_TOKEN);
  return {
    Authorization: token ? `Bearer ${token}` : null,
    'Content-Type': 'application/json',
    Connection: 'keep-alive',
    'Cache-Control': 'no-cache',
  } as any;
}

// 更新localstorage的downloadList
export function updateLocalStorageDownList(key: string, data: any) {
  try {
    localStorage.setItem(key, JSON.stringify(data));
  } catch (error) {
    console.error(`Failed to set ${key} in localStorage:`, error);
  }
}

export function getLocalStorageDownList(key: string) {
  try {
    const data = localStorage.getItem(key);
    return data ? JSON.parse(data) : [];
  } catch (error) {
    console.error(`Failed to parse ${key} from localStorage:`, error);
    return [];
  }
}

/**
 * 将 snake_case 字符串转换为 HTTP Header 格式（如 API-Host）
 * @param {string} str - 输入字符串，如 'api_host'
 * @returns {string} 转换后的字符串，如 'API-Host'
 */
export function toHttpHeaderFormat(str: string) {
  if (typeof str !== 'string' || str.length === 0) return '';

  return str
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join('-');
}
