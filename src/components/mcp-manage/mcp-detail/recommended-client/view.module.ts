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
      const data = await httpRequest.get(`/mcp/${serviceId}/clients`, {}, { baseURL: '/api' });
      if (!data) throw new Error('获取推荐客户端失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('授权mcp===>', data);
        setClients(data);
      },
      onError: (error) => {
        console.error('获取推荐客户端失败:', error);
        const testData = [
          {
            id: 1,
            name: '测试',
            icon: '123',
            description: '测试内容测试内容测试内容测试内容测试内容测试内容测试内容测试内容测试内容测试内容',
            link_command: 'npx cli-launch',
            related_tags: ['命令行', 'AI', '快速部署'],
            sort_weight: 100,
            create_at: 1744566069,
            create_by: 'admin',
            update_at: 1747234202,
            update_by: 'devops',
          },
        ];
        setClients(testData);
      },
    },
  );

  useEffect(() => {
    getClients();
  }, [serviceId]);

  return { clientLoading, clients };
}
