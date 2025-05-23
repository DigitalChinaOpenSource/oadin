import { useState, useEffect } from 'react';
import { httpRequest } from '@/utils/httpRequest';
import { message } from 'antd';
import { useRequest } from 'ahooks';
export function useViewModel() {
  const [checkStatus, setCheckStatus] = useState<boolean>(false);
  const { loading: checkHealthtLoading, run: fetchCheckHealth } = useRequest(
    async () => {
      const data = await httpRequest.get('/health', {}, { baseURL: '' });
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (data?.status === 'UP') setCheckStatus(true);
      },
      onError: (error) => {
        message.error('检查服务健康状态失败，请重试');
        console.error('检查服务健康状态失败:', error);
      },
    },
  );

  useEffect(() => {
    fetchCheckHealth();
  }, []);
  const handleRefresh = () => {
    fetchCheckHealth();
  };
  return { handleRefresh, checkHealthtLoading, checkStatus };
}
