// import { useNavigate } from 'react-router-dom';

import { useSearchParams } from 'react-router-dom';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';

export function useRecommendedClient() {
  // const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const serviceId = searchParams.get('serviceId');
  const [clients, setClients] = useState<Record<string, any>[]>([]);

  // 获取推荐客户端
  const { loading: clientLoading, run: getClients } = useRequest(
    async () => {
      // await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get(`/mcp/${serviceId}/clients`);
      if (!data) throw new Error('获取推荐客户端失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('获取推荐客户端===>', data);
        setClients(data);
      },
      onError: (error) => {
        console.error('获取推荐客户端失败:', error);
      },
    },
  );

  useEffect(() => {
    getClients();
  }, [serviceId]);

  return { clientLoading, clients };
}
