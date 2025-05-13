import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';
import { notification } from 'antd';

// 创建 axios 实例
const instance = axios.create({
  baseURL: '/byze/v0.2',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应数据类型
export interface ResponseData<T = any> {
  business_code: number;
  message: string;
  data: T;
}

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    // 可以在这里添加 token 等认证信息
    // const token = localStorage.getItem('token');
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`;
    // }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  },
);

// 响应拦截器
instance.interceptors.response.use(
  (response: AxiosResponse<ResponseData>) => {
    // 如果响应成功，直接返回数据
    const { data } = response;
    // 请求成功，直接返回data数据
    console.log('response.data=====>', data);
    if (data?.data) {
      return data.data;
    } else {
      // 业务逻辑错误
      notification.error({
        message: data?.message || '请求失败',
      });
      return Promise.reject(data);
    }
  },
  (error) => {
    // 处理 HTTP 错误
    if (error?.response) {
      // 服务器返回错误状态码
      const { status, statusText } = error.response;
      notification.error({
        message: `请求错误: ${status} ${statusText}`,
      });
    } else if (error?.request) {
      // 请求发出但没有收到响应
      notification.error({
        message: '网络异常，请检查您的网络连接',
      });
    } else {
      // 请求配置出错
      notification.error({
        message: `请求错误: ${error?.message}`,
      });
    }

    return Promise.reject(error);
  },
);

/**
 * GET 请求
 * @param url 请求地址
 * @param params 请求参数
 * @param config 请求配置
 * @returns Promise
 */
const get = <T = any>(url: string, params?: any, config?: AxiosRequestConfig) => {
  return instance.get<any, T>(url, { ...config, params });
};

/**
 * POST 请求
 * @param url 请求地址
 * @param data 请求数据
 * @param config 请求配置
 * @returns Promise
 */
const post = <T = any>(url: string, data?: any, config?: AxiosRequestConfig) => {
  return instance.post<any, T>(url, data, config);
};

// 导出为对象而不是 Hook
export const httpRequest = {
  get,
  post,
  request: instance.request,
};
