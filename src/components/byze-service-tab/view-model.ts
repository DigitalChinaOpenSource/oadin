import { useState } from 'react';
import { httpRequest } from '@/utils/httpRequest';
import { message } from 'antd';
import { useRequest } from 'ahooks';
export function useViewModel() {
  const [checkStatus, setCheckStatus] = useState<boolean>(false);
  const { loading: checkHealthtLoading, run: fetchCheckHealth } = useRequest(
    async () => {
      const data = await httpRequest.get('/health');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        // TODO 返回的数据层级存疑
        if (data?.status === 'up') setCheckStatus(true);
      },
      onError: (error) => {
        message.error('检查服务健康状态失败，请重试');
        console.error('检查服务健康状态失败:', error);
      },
    },
  );
  const handleRefresh = () => {
    fetchCheckHealth();
  };
  return { handleRefresh, checkHealthtLoading, checkStatus };
}
