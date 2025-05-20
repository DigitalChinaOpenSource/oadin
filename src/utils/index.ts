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
