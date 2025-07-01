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
    timeout: 120000,
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
        message.error(i18n.t('errors.service_error'));
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

// 创建请求函数，直接调用实例方法，不再包装健康检查
const createRequestFunctions = (instance: ReturnType<typeof createApiInstance>) => ({
  get: <T = any>(url: string, params?: any, config?: any) => {
    const mergedConfig = { ...config };
    // 如果配置中包含超时设置，则覆盖默认值
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return instance.get<any, T>(url, { ...mergedConfig, params });
  },

  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig & IModelChangeStore, 'data'>) => {
    const mergedConfig = { ...config };
    // 如果配置中包含超时设置，则覆盖默认值
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return instance.post<any, T>(url, data, mergedConfig);
  },

  put: <T = any>(url: string, data?: any, config?: any) => {
    const mergedConfig = { ...config };
    // 如果配置中包含超时设置，则覆盖默认值
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return instance.put<any, T>(url, data, mergedConfig);
  },

  del: <T = any>(url: string, data?: any, config?: any) => {
    const mergedConfig = { ...config };
    // 如果配置中包含超时设置，则覆盖默认值
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return instance.delete<any, T>(url, { ...mergedConfig, data });
  },
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
  get: <T = any>(url: string, params?: any, config?: any) => {
    const mergedConfig = { ...config };
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return healthInstance.get<any, T>(url, { ...mergedConfig, params });
  },

  post: <T = any>(url: string, data?: any, config?: Omit<AxiosRequestConfig & IModelChangeStore, 'data'>) => {
    const mergedConfig = { ...config };
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return healthInstance.post<any, T>(url, data, mergedConfig);
  },

  put: <T = any>(url: string, data?: any, config?: any) => {
    const mergedConfig = { ...config };
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return healthInstance.put<any, T>(url, data, mergedConfig);
  },

  del: <T = any>(url: string, data?: any, config?: any) => {
    const mergedConfig = { ...config };
    if (config?.timeout) {
      mergedConfig.timeout = config.timeout;
    }
    return healthInstance.delete<any, T>(url, { ...mergedConfig, data });
  },

  request: (config: AxiosRequestConfig) => {
    // 对于直接的request调用也支持超时配置
    return healthInstance.request(config);
  },
};

// 导出一个手动检查健康状态的函数，供需要时调用
export const checkByzeHealth = async () => {
  try {
    const storeState = useByzeServerCheckStore.getState();
    await storeState.fetchByzeServerStatus();
    return useByzeServerCheckStore.getState().checkByzeStatus;
  } catch (error) {
    console.error('Health check failed:', error);
    return false;
  }
};
