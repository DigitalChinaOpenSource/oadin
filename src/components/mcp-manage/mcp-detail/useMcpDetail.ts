import { useNavigate, useSearchParams } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import testDta from './mcp_schema.json';
import { Modal } from 'antd';

export const useMcpDetail = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const serviceId = searchParams.get('serviceId');
  const mcpFrom = searchParams.get('mcpFrom');
  const [mcpDetail, setMcpDetail] = useState<McpDetailType>();
  const [showMcpModal, setShowMcpModal] = useState(false);
  // const [authMcpParams,setAuthMcpParams] = useState<any>();

  // 获取 mcp 详情
  const { loading: mcpDetailLoading, run: fetchMcpDetail } = useRequest(
    async () => {
      const data = await httpRequest.get<McpDetailType>(`/mcp/${serviceId}`);
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
      return await httpRequest.put<McpDetailType>(`/mcp/${serviceId}/download`, null, { timeout: 2 * 60 * 1000 });
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('下载mcp===>', data);
        // 下载成功修改 下载状态值为1   1为已授权（若需授权）已下载 --- 一切就绪
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
  const { loading: authMcpLoading, runAsync: authMcp } = useRequest(
    async (authParams) => {
      return await httpRequest.put<McpDetailType>(`/mcp/${serviceId}/auth`, authParams);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('授权mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          authorized: 1, // 1为已授权
        });
        // 授权成功 开始下载
        downMcp();
      },
      onError: (error) => {
        console.error('授权mcp失败:', error);
      },
    },
  );

  //添加mcp点击
  const handleAddMcp = async () => {
    try {
      const { hosted, status, envRequired } = mcpDetail || {};
      if (envRequired === 0) {
        // 不需要授权
        Modal.confirm({
          title: '确认添加吗？',
          okText: '确认',
          centered: true,
          okButtonProps: {
            style: { backgroundColor: '#4f4dff' },
          },
          onOk() {
            downMcp(); // 直接下载mcp
          },
          onCancel() {
            console.log('Cancel');
          },
        });
      } else {
        // 显示授权弹窗
        setShowMcpModal(true);
      }
    } catch (error) {
      console.error('添加mcp失败:', error);
    }
  };

  // 取消mcp
  const { loading: cancelMcpLoading, run: handleCancelMcp } = useRequest(
    async () => {
      return await httpRequest.put(`/mcp/${serviceId}/reverse`);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('授权mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          status: 0,
        });
      },
      onError: (error) => {
        console.error('授权mcp失败:', error);
        return false;
      },
    },
  );

  // 授权mcp 确认
  const handleAuthMcp = async (authParams: any) => {
    setShowMcpModal(false);
    try {
      await authMcp(authParams);
    } catch (error) {
      console.error('授权mcp失败:', error);
    }
  };

  // 页面返回
  const handleGoBack = (): void => {
    navigate(`/mcp-service?mcpFrom=${mcpFrom}`);
  };

  useEffect(() => {
    fetchMcpDetail();
  }, [serviceId]);

  return {
    mcpDetailLoading,
    downMcpLoading,
    handleGoBack,
    mcpDetail,
    handleAddMcp,
    handleCancelMcp,
    cancelMcpLoading,
    showMcpModal,
    setShowMcpModal,
    authMcpLoading,
    authMcp,
    handleAuthMcp,
  };
};
