export const DOWNLOAD_STATUS = {
  // 0：失败  1：下载中   2：下载完成   3. 暂停
  FAILED: 0,
  IN_PROGRESS: 1,
  COMPLETED: 2,
  PAUSED: 3,
};

export const AUTH_TOKEN = 'anythingllm_authToken';

export const RECOMMEND_MODEL = ['神州数码|DeepSeek-R1'];
export const PIORITY_MODEL = ['神州数码|DeepSeek-R1', '深度求索|DeepSeek-R1', '硅基流动|DeepSeek-R1', '深度求索|DeepSeek-V3', '硅基流动|DeepSeek-V3'];

// 本地存储键名常量
export const LOCAL_STORAGE_KEYS = {
  DOWN_LIST: 'downList',
};

export const API_VERSION = 'v0.2';
export const API_PREFIX = `/byze/${API_VERSION}`;
export const API_HEALTH_ENDPOINT = '/health';

// 完整的API基础URL（根据环境变量或默认值）
export const getApiBaseUrl = () => {
  return import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.MODE === 'dev' ? API_PREFIX : `http://127.0.0.1:16688${API_PREFIX}`);
};

// 健康检查URL
export const getHealthBaseUrl = () => {
  return import.meta.env.VITE_HEALTH_API_URL ?? (import.meta.env.MODE === 'dev' ? '/' : 'http://127.0.0.1:16688');
};
