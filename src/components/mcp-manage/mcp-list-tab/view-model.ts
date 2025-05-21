import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData } from './types';
import { mcpListDataMock } from './constant';
export function useViewModel() {
  const [mcpListData, setMcpListData] = useState([]);
  const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);
  useEffect(() => {
    fetchMcpList({
      deployment: 'hosted',
      page: 1,
      size: 10,
    });
  }, []);

  // 获取 mcp 列表
  const { loading: mcpListLoading, run: fetchMcpList } = useRequest(
    async (params: IMcpListRequestParams) => {
      const data = await httpRequest.post<IMcpListData>('/mcp/search', params, { baseURL: '/api' });
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

  return {
    mcpListLoading,
    mcpListData,
    mcpListDataMock,
  };
}
