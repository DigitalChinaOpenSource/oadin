import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import { notification } from 'antd';

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
        const { status, statusText } = error.response;
        notification.error({
          message: `请求错误: ${status} ${statusText}`,
        });
      } else if (error?.request) {
        notification.error({
          message: '网络异常，请检查您的网络连接',
        });
      } else {
        notification.error({
          message: `请求错误: ${error?.message}`,
        });
      }

      return Promise.reject(error);
    },
  );

  return instance;
};

const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.MODE === 'dev' ? '/byze/v0.2' : 'http://127.0.0.1:16688/byze/v0.2');

const healthBaseURL = import.meta.env.VITE_HEALTH_API_URL ?? (import.meta.env.MODE === 'dev' ? '/' : 'http://127.0.0.1:16688');

const byzeInstance = createApiInstance(apiBaseURL);
const healthInstance = createApiInstance(healthBaseURL);

const createRequestFunctions = (instance: ReturnType<typeof createApiInstance>) => ({
  get: <T = any>(url: string, params?: any, config?: any) => {
    return instance.get<any, T>(url, { ...config, params });
  },
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig, 'data'>) => {
    return instance.post<any, T>(url, data, config);
  },
  put: <T = any>(url: string, data?: any, config?: any) => {
    return instance.put<any, T>(url, data, config);
  },
  del: <T = any>(url: string, data?: any, config?: any) => {
    return instance.delete<any, T>(url, { ...config, data });
  },

  request: instance.request,
});

// 分别导出两个请求工具
export const httpRequest = createRequestFunctions(byzeInstance);
export const healthRequest = createRequestFunctions(healthInstance);
