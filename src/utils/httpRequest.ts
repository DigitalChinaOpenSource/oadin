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
    const { data } = response;
    if (data?.data) {
      return data.data;
    } else {
      return data;
    }
    // 后端接口返回数据格式不统一，暂时不管业务逻辑失败的情况
    //  else {
    //   notification.error({
    //     message: data?.message || '请求失败',
    //   });
    //   return Promise.reject(data);
    // }
  },
  (error) => {
    // 处理 HTTP 错误
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
      // 请求配置出错
      notification.error({
        message: `请求错误: ${error?.message}`,
      });
    }

    return Promise.reject(error);
  },
);

/**
 * 请求
 * @param url 请求地址
 * @param params 请求参数
 * @param config 请求配置
 * @returns Promise
 */
const get = <T = any>(url: string, params?: any, config?: any) => {
  return instance.get<any, T>(url, { ...config, params, baseURL: config?.baseURL ?? instance.defaults.baseURL });
};

const post = <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig, 'data'>) => {
  return instance.post<any, T>(url, data, { ...config, baseURL: config?.baseURL || instance.defaults.baseURL });
};

const put = <T = any>(url: string, data?: any, config?: any) => {
  return instance.put<any, T>(url, data, { ...config, baseURL: config?.baseURL || instance.defaults.baseURL });
};

const del = <T = any>(url: string, data?: any, config?: any) => {
  return instance.delete<any, T>(url, { data });
};

export const httpRequest = {
  get,
  post,
  del,
  put,
  request: instance.request,
};
