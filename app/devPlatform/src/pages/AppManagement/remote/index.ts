import { httpRequest } from '@/utils/httpRequest.ts';
import { IApplicationParams } from '@/pages/AppManagement/remote/type';
import { message } from 'antd';

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
