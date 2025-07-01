import { httpRequest } from '@/utils/httpRequest.ts';
import { IApplicationDetail, IApplicationParams, ISaveApplicationConfigParams } from '@/pages/AppManagement/remote/type';
import { message } from 'antd';
import { SearchParams } from '@/pages/AppManagement/AppConfig/types.ts';

export const getApplicationList = async (params?: IApplicationParams) => {
  const data = await httpRequest.post('/application/search', params);
  if (data?.code === 200) {
    return data?.data;
  } else {
    message.error(data?.message || '获取应用列表失败');
    return data?.data;
  }
};

export const addApplication = async (params: { name: string }) => {
  const data = await httpRequest.post('/application', params);
  console.info(data, '创建内容');
  if (data?.code === 200) {
    return true;
  } else {
    message.error(data?.message || '创建应用失败');
    return false;
  }
};

export const deleteApplication = async (id: string) => {
  const data = await httpRequest.del(`/application/${id}`);
  if (data?.code === 200) {
    return true;
  } else {
    message.error(data?.message || '删除应用失败');
    return false;
  }
};
export const getAppDetail = async (id: string): Promise<IApplicationDetail | null> => {
  const data = await httpRequest.get(`/application/${id}`);
  if (data?.code === 200) {
    return data?.data;
  } else {
    message.error(data?.message || '删除应用失败');
    return null;
  }
};

/// 更新密钥
export const updateAppSecret = async (id: string): Promise<boolean> => {
  const data = await httpRequest.put(`/application/${id}/refresh_secret`);
  if (data?.code === 200) {
    return true;
  } else {
    message.error(data?.message || '更新密钥失败');
    return false;
  }
};

/// 修改应用名称
export const updateAppName = async (id: string, name: string): Promise<boolean> => {
  const data = await httpRequest.put(`/application/${id}/rename`, { name, id });
  if (data?.code === 200) {
    return true;
  } else {
    message.error(data?.message || '修改应用名称失败');
    return false;
  }
};
/// 获取模型列表
export const getModelList = async (params: SearchParams) => {
  try {
    const data = await httpRequest.post('/model/models', params);
    if (data?.code === 200) {
      return data?.data;
    } else {
      message.error(data?.message || '获取模型列表失败');
      return [];
    }
  } catch (e) {
    return [];
  }
};

/// 获取Mcp列表
export const getMcpList = async (params: SearchParams) => {
  try {
    const data = await httpRequest.post('/mcp', params);
    if (data?.code === 200) {
      return data?.data;
    } else {
      message.error(data?.message || '获取MCP列表失败');
      return [];
    }
  } catch (e) {
    return [];
  }
};

/// 保存配置
export const saveAppConfig = async (params: ISaveApplicationConfigParams) => {
  try {
    const data = await httpRequest.put(`/application/${params.id}/config`, params);
    return data?.code === 200;
  } catch (e) {
    return false;
  }
};
