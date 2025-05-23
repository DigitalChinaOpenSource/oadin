import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData } from './types';
import { mcpListDataMock } from './constant';
import { useNavigate } from 'react-router-dom';

export function useViewModel() {
  const navigate = useNavigate();

  const [mcpListData, setMcpListData] = useState([]);
  const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);
  // 过滤器是否折叠了
  const [collapsed, setCollapsed] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 12,
    total: 0,
  });

  const lastPageSizeRef = useRef(pagination.pageSize);

  // useEffect(() => {
  //   fetchMcpList({
  //     deployment: 'hosted',
  //     page: 1,
  //     size: 12,
  //   });
  // }, []);

  useEffect(() => {
    // TODO 调列表接口，将pagination, mcpSearchVal合并发送
    // fetchMcpList();
  }, [pagination, mcpSearchVal]);

  // 获取 mcp 列表
  const { loading: mcpListLoading, run: fetchMcpList } = useRequest(
    async (params: IMcpListRequestParams) => {
      const data = await httpRequest.post<IMcpListData>('/api/mcp/search', params);
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('fetchMcpList===>', data);
        setMcpListData(mcpListData);
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  const handelMcpCardClick = (serviceId: number) => {
    navigate(`/mcp-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };

  // 搜索框搜索
  const onMcpInputSearch = () => {
    setMcpSearchVal({
      ...mcpSearchVal,
      keyword: mcpSearchVal.keyword,
    });
  };

  const onPageChange = (current: number) => {
    // 如果 pageSize 刚刚被改变，则不执行页码变更逻辑
    if (lastPageSizeRef.current !== pagination.pageSize) {
      lastPageSizeRef.current = pagination.pageSize;
      return;
    }
    setPagination({ ...pagination, current });
  };

  const onShowSizeChange = (current: number, pageSize: number) => {
    lastPageSizeRef.current = pageSize;
    setPagination({ ...pagination, current: 1, pageSize });
  };

  const onMcpInputChange = (value: string) => {
    setMcpSearchVal({ ...mcpSearchVal, keyword: value.trim() });
  };

  return {
    mcpListLoading,
    mcpListData: mcpListDataMock,
    mcpSearchVal,
    onMcpInputChange,

    handelMcpCardClick,
    onMcpInputSearch,
    collapsed,
    setCollapsed,

    onPageChange,
    onShowSizeChange,
  };
}
