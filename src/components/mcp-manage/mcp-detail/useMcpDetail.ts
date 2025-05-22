import { useNavigate, useSearchParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import testDta from './mcp_schema.json';

export const useMcpDetail = (id?: string | null | number) => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const serviceId = searchParams.get('serviceId');
  const mcpFrom = searchParams.get('mcpFrom');
  const [mcpDetail, setMcpDetail] = useState<McpDetailType>();
  const [showMcpModal, setShowMcpModal] = useState(false);
  // const [authMcpParams,setAuthMcpParams] = useState<any>();

  // 获取 mcp 详情
  const { loading: mcpDetailLoading, run: fetchMcpDetail } = useRequest(
    async (serviceId) => {
      const data = await httpRequest.get<McpDetailType>(`/mcp/${serviceId}`, {}, { baseURL: '/api' });
      // if (!data) throw new Error('获取mcp详情失败');
      return data || testDta;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('fetchMcpList===>', data);
        setMcpDetail(data);
      },
      onError: (error) => {
        console.error('获取mcp详情失败:', error);
      },
    },
  );

  // 下载mcp
  const { loading: downMcpLoading, run: downMcp } = useRequest(
    async () => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get<McpDetailType>(`/mcp/${serviceId}/download`, {}, { baseURL: '/api' });
      if (!data) throw new Error('下载mcp失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('下载mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          status: 1,
        });
      },
      onError: (error) => {
        console.error('下载本地mcp失败:', error);
      },
    },
  );

  // 远端mcp授权
  const { loading: authMcpLoading, run: authMcp } = useRequest(
    async (authParams) => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.put<McpDetailType>(`/mcp/${serviceId}/auth`, authParams, { baseURL: '/api' });
      if (!data) throw new Error('授权mcp失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('授权mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          status: 1,
        });
      },
      onError: (error) => {
        console.error('授权mcp失败:', error);
      },
    },
  );

  //添加mcp点击
  const handleAddMcp = async () => {
    const { hosted } = mcpDetail || {};
    try {
      if (hosted) {
        setShowMcpModal(true);
      } else {
        downMcp();
      }
    } catch (error) {
      console.error('添加mcp失败:', error);
    }
  };

  // 授权mcp 确认
  const handleAuthMcp = async (authParams: any) => {
    setShowMcpModal(false);
    authMcp(authParams);
  };

  // 页面返回
  const handleGoBack = (): void => {
    navigate(`/mcp-service?mcpFrom=${mcpFrom}`);
  };

  useEffect(() => {
    if (id === serviceId) return;
    fetchMcpDetail(serviceId);
  }, [serviceId]);

  return {
    mcpDetailLoading,
    downMcpLoading,
    handleGoBack,
    mcpDetail,
    handleAddMcp,
    showMcpModal,
    setShowMcpModal,
    authMcpLoading,
    authMcp,
    handleAuthMcp,
  };
};
