import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
import { message } from 'antd';

export interface ResponseData<T = any> {
  business_code: number;
  message: string;
  data: T;
}

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
      const { data } = response;
      if (data?.data) {
        return data.data;
      } else {
        return data;
      }
    },
    (error) => {
      if (error?.response) {
        // 配置授权参数错误的情况，不予处理
        if (error.response?.data?.business_code === 20003) {
          return Promise.reject(error);
        }
        if (error.response?.data?.error) {
          message.error(`${error.response?.data.error}`);
          return Promise.reject(error);
        }

        const { status, statusText } = error.response;
        message.error(`请求错误: ${status} ${statusText}`);
      } else if (error?.request) {
        message.error('网络异常，请检查您的网络连接');
      } else {
        message.error(`请求错误: ${error?.message}`);
      }

      return Promise.reject(error);
    },
  );

  return instance;
};

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.MODE === 'dev' ? '/byze/v0.2' : 'http://127.0.0.1:16688/byze/v0.2');

const healthBaseURL = import.meta.env.VITE_HEALTH_API_URL ?? (import.meta.env.MODE === 'dev' ? '/' : 'http://127.0.0.1:16688');

const byzeInstance = createApiInstance(apiBaseURL);

// 健康检查包装器
async function withHealthCheck<T>(requestFn: () => Promise<T>): Promise<T> {
  const { fetchByzeServerStatus } = useByzeServerCheckStore.getState();
  await fetchByzeServerStatus();
  if (!useByzeServerCheckStore.getState().checkByzeStatus) {
    message.destroy();
    message.error('白泽服务不可用，请确认白泽服务启动状态');
    // 返回一个永远 pending 的 Promise，阻断后续 then/catch
    return new Promise(() => {});
  }
  return requestFn();
}

const createRequestFunctions = (instance: ReturnType<typeof createApiInstance>) => ({
  get: <T = any>(url: string, params?: any, config?: any) => withHealthCheck(() => instance.get<any, T>(url, { ...config, params })),
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig, 'data'>) => withHealthCheck(() => instance.post<any, T>(url, data, config)),
  put: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.put<any, T>(url, data, config)),
  del: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.delete<any, T>(url, { ...config, data })),
});

export const httpRequest = createRequestFunctions(byzeInstance);

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
      message.error('白泽服务不可用，请确认白泽服务启动状态');
      return Promise.reject(error);
    },
  );

  return instance;
};

const healthInstance = createHealthApiInstance(healthBaseURL);

export const healthRequest = {
  get: <T = any>(url: string, params?: any, config?: any) => healthInstance.get<any, T>(url, { ...config, params }),
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig, 'data'>) => healthInstance.post<any, T>(url, data, config),
  put: <T = any>(url: string, data?: any, config?: any) => healthInstance.put<any, T>(url, data, config),
  del: <T = any>(url: string, data?: any, config?: any) => healthInstance.delete<any, T>(url, { ...config, data }),
  request: (config: AxiosRequestConfig) => healthInstance.request(config),
};
