import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
import { message } from 'antd';
import { API_PREFIX } from '@/constants';
import i18n from '@/i18n';

interface IModelChangeStore {
  needModelChangeStore?: boolean;
  setMigratingStatus?: (status: 'init' | 'pending' | 'failed') => void;
}

declare module 'axios' {
  export interface AxiosError {
    handled?: boolean;
  }
  export interface InternalAxiosRequestConfig extends IModelChangeStore {
    needMcpStore?: boolean;
    addMcpDownloadItem?: (item: { id: string; error: string; downStatus: string }) => void;
    mcpId?: string;
  }
}

export interface ResponseData<T = any> {
  business_code: number;
  message: string;
  data: T;
}

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.MODE === 'dev' ? API_PREFIX : `http://127.0.0.1:16688${API_PREFIX}`);

const healthBaseURL = import.meta.env.VITE_HEALTH_API_URL ?? (import.meta.env.MODE === 'dev' ? '/' : 'http://127.0.0.1:16688');

const createApiInstance = (baseURL: string) => {
  const instance = axios.create({
    baseURL,
    timeout: 60000,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
    },
  });

  // 请求拦截器
  instance.interceptors.request.use(
    (config) => {
      // 可以在这里添加 token 等认证信息
      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ResponseData>) => {
      const { data, config } = response;
      // 处理响应成功时，检查是否需要更新 MCP 下载状态
      if (config.needMcpStore && config.addMcpDownloadItem && config.mcpId) {
        config.addMcpDownloadItem({
          id: config.mcpId,
          error: '成功',
          downStatus: 'success',
        });
      }
      // 处理响应成功时，模型迁移状态
      if (config.needModelChangeStore && config.setMigratingStatus) {
        config.setMigratingStatus('init');
      }
      if (data?.data) {
        return data.data;
      } else {
        return data;
      }
    },
    (error) => {
      const { config } = error;
      // 处理错误时，检查是否需要更新 MCP 下载状态
      if (config.needMcpStore) {
        config.addMcpDownloadItem({
          id: config.mcpId,
          error: error.message,
          downStatus: 'error',
        });
      }
      // 处理错误时，模型迁移状态
      if (config.needModelChangeStore) {
        config.setMigratingStatus('failed');
      }
      message.destroy();
      if (error?.response) {
        const { data } = error.response;
        // 获取业务错误码
        const businessCode = data?.business_code?.toString();

        if (businessCode) {
          const errorMessage = i18n.t(`errors.${businessCode}`, {
            defaultValue: data?.message || i18n.t('errors.service_error'),
          });
          message.error(errorMessage);
          error.handled = true;
        } else {
          message.error(data?.message || i18n.t('errors.service_error'));
          error.handled = true;
        }
      } else if (error?.request) {
        message.error(i18n.t('errors.network'));
        error.handled = true;
      } else {
        message.error(error?.message || i18n.t('errors.service_error'));
        error.handled = true;
      }
      return Promise.reject(error);
    },
  );

  return instance;
};

const byzeInstance = createApiInstance(apiBaseURL);

// 健康检查包装器
async function withHealthCheck<T>(requestFn: () => Promise<T>): Promise<T> {
  const { fetchByzeServerStatus } = useByzeServerCheckStore.getState();
  await fetchByzeServerStatus();
  if (!useByzeServerCheckStore.getState().checkByzeStatus) {
    message.destroy();
    message.error('奥丁服务不可用，请确认奥丁服务启动状态');
    // 返回一个永远 pending 的 Promise，阻断后续 then/catch
    return new Promise(() => {});
  }
  return requestFn();
}

const createRequestFunctions = (instance: ReturnType<typeof createApiInstance>) => ({
  get: <T = any>(url: string, params?: any, config?: any) => withHealthCheck(() => instance.get<any, T>(url, { ...config, params })),
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig & IModelChangeStore, 'data'>) => withHealthCheck(() => instance.post<any, T>(url, data, config)),
  put: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.put<any, T>(url, data, config)),
  del: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.delete<any, T>(url, { ...config, data })),
});

export const httpRequest = createRequestFunctions(byzeInstance);

// Byze 的健康检查，请求路径不同，需要在底层做特殊处理
const createHealthApiInstance = (baseURL: string) => {
  const instance = axios.create({
    baseURL,
    timeout: 60000,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
    },
  });

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ResponseData>) => {
      const { data } = response;
      if (data?.data) {
        return data.data;
      } else {
        return data;
      }
    },
    (error) => {
      // 只要 /health 请求出错，统一提示
      message.error(i18n.t('errors.byze_unavailable'));
      error.handled = true;
      return Promise.reject(error);
    },
  );

  return instance;
};

const healthInstance = createHealthApiInstance(healthBaseURL);

export const healthRequest = {
  get: <T = any>(url: string, params?: any, config?: any) => healthInstance.get<any, T>(url, { ...config, params }),
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig & IModelChangeStore, 'data'>) => healthInstance.post<any, T>(url, data, config),
  put: <T = any>(url: string, data?: any, config?: any) => healthInstance.put<any, T>(url, data, config),
  del: <T = any>(url: string, data?: any, config?: any) => healthInstance.delete<any, T>(url, { ...config, data }),
  request: (config: AxiosRequestConfig) => healthInstance.request(config),
};
