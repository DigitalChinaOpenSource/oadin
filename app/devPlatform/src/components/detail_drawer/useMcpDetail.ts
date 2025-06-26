import { useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { McpDetailType } from './types.ts';

export const useMcpDetail = (id?: string | number) => {
  const serviceId = id;
  const [mcpDetail, setMcpDetail] = useState<McpDetailType | null>(null);
  const [showMcpModal, setShowMcpModal] = useState(false);

  // 获取 mcp 详情
  const { loading: mcpDetailLoading, runAsync: fetchMcpDetail } = useRequest(
    async () => {
      return await httpRequest.get<McpDetailType>(`/mcp/${serviceId}`);
    },
    {
      manual: true,
      onSuccess: (data) => {
        setMcpDetail(data);
      },
      onError: (error) => {
        console.error('获取mcp详情失败:', error);
      },
    },
  );

  return {
    mcpDetailLoading,
    mcpDetail,
    setMcpDetail,
    showMcpModal,
    setShowMcpModal,
    fetchMcpDetail,
    serviceId,
  };
};
