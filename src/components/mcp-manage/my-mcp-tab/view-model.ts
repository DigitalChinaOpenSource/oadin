import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData } from '../mcp-list-tab/types';
import { mcpListDataMock } from '../mcp-list-tab/constant';
import { useNavigate } from 'react-router-dom';

export function useViewModel() {
  const navigate = useNavigate();

  const [myMcpListData, setMyMcpListData] = useState([]);
  const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);

  // useEffect(() => {
  //   fetchMyMcpList({
  //     deployment: 'hosted',
  //     page: 1,
  //     size: 10,
  //   });
  // }, []);

  // 获取我的 mcp 列表
  const { loading: myMcpListLoading, run: fetchMyMcpList } = useRequest(
    async (params: IMcpListRequestParams) => {
      const data = await httpRequest.post<IMcpListData>('/api/mcp/search', params);
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('fetchMcpList===>', data);
        setMyMcpListData(myMcpListData);
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  const handelMcpCardClick = (serviceId: number) => {
    navigate(`/mcp-detail?serviceId=${serviceId}&mcpFrom=myMcpList`);
  };

  // 搜索框搜索
  const onMcpInputSearch = (inputSearchVal: string) => {
    setMcpSearchVal({
      ...mcpSearchVal,
      keyword: inputSearchVal,
    });
  };

  return {
    myMcpListLoading,
    myMcpListData,
    mcpListDataMock,
    mcpSearchVal,

    handelMcpCardClick,
    onMcpInputSearch,
  };
}
