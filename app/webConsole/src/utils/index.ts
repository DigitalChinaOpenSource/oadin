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

/**
 * 将文件大小字符串转换为以 MB 为单位的数字
 * @param sizeStr 文件大小字符串，例如：'30MB'、'3.9GB'、'1.5 TB'
 * @returns 以 MB 为单位的数字
 */
export const convertToMB = (sizeStr: string): number => {
  if (!sizeStr) return 0;

  const normalizedStr = sizeStr.toString().replace(/\s+/g, '').toUpperCase();

  const match = normalizedStr.match(/^([\d.]+)([KMGT]B)$/i);
  if (!match) return 0;

  const value = parseFloat(match[1]);
  const unit = match[2].toUpperCase();

  // 根据单位进行转换
  switch (unit) {
    case 'KB':
      return value / 1024;
    case 'MB':
      return value;
    case 'GB':
      return value * 1024;
    case 'TB':
      return value * 1024 * 1024;
    default:
      return value;
  }
};
