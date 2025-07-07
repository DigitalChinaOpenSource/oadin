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

import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import useServerCheckStore from '@/store/useServerCheckStore';
import { message } from 'antd';
import { API_PREFIX } from '@/constants';
import i18n from '@/i18n';

declare module 'axios' {
  export interface AxiosError {
    handled?: boolean;
  }
}

export interface ResponseData<T = any> {
  business_code: number;
  message: string;
  data: T;
}

// const apiBaseURL = import.meta.env.VITE_API_BASE_URL ?? (import.meta.env.MODE === 'dev' ? API_PREFIX : `http://127.0.0.1:16688${API_PREFIX}`);

// const healthBaseURL = import.meta.env.VITE_HEALTH_API_URL ?? (import.meta.env.MODE === 'dev' ? '/' : 'http://127.0.0.1:16688');
const apiBaseURL = import.meta.env.VITE_API_BASE_URL;

const healthBaseURL = import.meta.env.VITE_HEALTH_API_URL;
const createApiInstance = (baseURL: string) => {
  const instance = axios.create({
    baseURL,
    // timeout: 60000,
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
      message.destroy();
      if (error?.response) {
        const { data } = error.response;
        // 获取业务错误码
        const businessCode = data?.business_code?.toString();

        if (businessCode) {
          const errorMessage = i18n.t(`errors.${businessCode}`, {
            defaultValue: data?.message || '未知错误',
          });
          message.open({
            content: errorMessage,
            type: 'error',
          });
          // message.error(errorMessage);
          error.handled = true;
        } else {
          message.open({
            content: data?.message || data?.error || i18n.t('errors.unknown'),
            type: 'error',
          });
          // message.error(data?.message || data?.error || i18n.t('errors.unknown'));
          error.handled = true;
        }
      } else if (error?.request) {
        message.open({
          content: i18n.t('errors.network'),
          type: 'error',
        });
        // message.error(i18n.t('errors.network'));
        error.handled = true;
      } else {
        message.open({
          content: error?.message || i18n.t('errors.unknown'),
          type: 'error',
        });
        // message.error(error?.message || i18n.t('errors.unknown'));
        error.handled = true;
      }
      return Promise.reject(error);
    },
  );

  return instance;
};

const apiInstance = createApiInstance(apiBaseURL);

// 健康检查包装器
async function withHealthCheck<T>(requestFn: () => Promise<T>): Promise<T> {
  const { fetchServerStatus } = useServerCheckStore.getState();
  await fetchServerStatus();
  if (!useServerCheckStore.getState().checkStatus) {
    message.destroy();

    message.open({
      content: i18n.t('errors.unavailable'),
      type: 'error',
    });
    // message.error('服务不可用，请确认服务启动状态');
    // 返回一个永远 pending 的 Promise，阻断后续 then/catch
    // return new Promise(() => {});
    return Promise.reject(new Error(i18n.t('errors.unavailable')));
  }
  return requestFn();
}

const createRequestFunctions = (instance: ReturnType<typeof createApiInstance>) => ({
  get: <T = any>(url: string, params?: any, config?: any) => withHealthCheck(() => instance.get<any, T>(url, { ...config, params })),
  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig, 'data'>) => withHealthCheck(() => instance.post<any, T>(url, data, config)),
  put: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.put<any, T>(url, data, config)),
  del: <T = any>(url: string, data?: any, config?: any) => withHealthCheck(() => instance.delete<any, T>(url, { ...config, data })),
});

export const httpRequest = createRequestFunctions(apiInstance);

// 健康检查，请求路径不同，需要在底层做特殊处理
const createHealthApiInstance = (baseURL: string) => {
  const instance = axios.create({
    baseURL,
    // timeout: 60000,
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
      message.open({
        content: i18n.t('errors.unavailable'),
        type: 'error',
      });
      // message.error(i18n.t('errors.unavailable'));
      error.handled = true;
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
