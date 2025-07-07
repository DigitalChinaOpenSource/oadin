//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

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
export const upFirstLetter = (str: string): string => {
  if (typeof str !== 'string' || str.length === 0) return str;
  return str.charAt(0).toUpperCase() + str.slice(1);
};
export const formatContentLength = (length: number): string => {
  length = length / 1000;
  const units = ['K', 'M'];
  let i = 0;

  while (length >= 1000 && i < units.length - 1) {
    length /= 1000;
    i++;
  }

  return `${Math.round(length)}${units[i]}`;
};
