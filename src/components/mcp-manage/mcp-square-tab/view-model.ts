import { useState, useEffect } from 'react';
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

  // useEffect(() => {
  //   fetchMcpList({
  //     deployment: 'hosted',
  //     page: 1,
  //     size: 10,
  //   });
  // }, []);

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
  const onMcpInputSearch = (inputSearchVal: string) => {
    setMcpSearchVal({
      ...mcpSearchVal,
      keyword: inputSearchVal,
    });
  };

  return {
    mcpListLoading,
    mcpListData,
    mcpListDataMock,
    mcpSearchVal,

    handelMcpCardClick,
    onMcpInputSearch,
    collapsed,
    setCollapsed,
  };
}
