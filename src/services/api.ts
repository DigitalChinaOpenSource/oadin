import axios from 'axios';

// 创建axios实例
const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 可以在这里添加token等认证信息
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response.data;
  },
  (error) => {
    // 处理错误响应
    return Promise.reject(error);
  }
);

// 模型相关API
export const modelApi = {
  getModels: () => api.get('/models'),
  getModelById: (id: string) => api.get(`/models/${id}`),
  createModel: (data: any) => api.post('/models', data),
  updateModel: (id: string, data: any) => api.put(`/models/${id}`, data),
  deleteModel: (id: string) => api.delete(`/models/${id}`),
  downloadModel: (id: string) => api.get(`/models/${id}/download`),
};

// 服务相关API
export const serverApi = {
  getServers: () => api.get('/servers'),
  getServerById: (id: string) => api.get(`/servers/${id}`),
  createServer: (data: any) => api.post('/servers', data),
  updateServer: (id: string, data: any) => api.put(`/servers/${id}`, data),
  deleteServer: (id: string) => api.delete(`/servers/${id}`),
};

export default api;