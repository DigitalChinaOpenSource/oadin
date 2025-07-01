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
  MODEL_DOWNLOAD_LIST: 'modelDownloadList',
};

export const API_VERSION = 'v0.2';
export const API_PREFIX = `/oadin/${API_VERSION}`;
export const API_HEALTH_ENDPOINT = '/health';

export const EMBEDMODELID = '87c0b009-2d93-4f00-9662-333037666261373163373263'; // 词嵌入模型ID
