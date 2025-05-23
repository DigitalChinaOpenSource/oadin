import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData } from '../mcp-square-tab/types';
import { mcpListDataMock } from '../mcp-square-tab/constant';
import { useNavigate } from 'react-router-dom';

export function useViewModel() {
  const navigate = useNavigate();

  const [myMcpListData, setMyMcpListData] = useState([]);
  const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 12,
    total: 0,
  });
  const isInitialMount = useRef(true);
  const lastPageSizeRef = useRef(pagination.pageSize);

  useEffect(() => {
    fetchMyMcpList({
      page: 1,
      size: 12,
    });
  }, []);

  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false;
      return;
    }

    const params = {
      page: pagination.current,
      size: pagination.pageSize,
    } as IMcpListRequestParams;
    if (mcpSearchVal.keyword) {
      params.keyword = mcpSearchVal.keyword;
    }
    fetchMyMcpList(params);
  }, [pagination, mcpSearchVal]);

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
  const onMcpInputSearch = () => {
    console.log('onMcpInputSearch', mcpSearchVal);
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
    console.log('onPageChange', { ...pagination, current });
    setPagination({ ...pagination, current });
  };

  const onShowSizeChange = (current: number, pageSize: number) => {
    console.log('onShowSizeChange', current, pageSize);
    lastPageSizeRef.current = pageSize;
    setPagination({ ...pagination, current: 1, pageSize });
  };

  const onMcpInputChange = (value: string) => {
    setMcpSearchVal({ ...mcpSearchVal, keyword: value.trim() });
  };
  return {
    myMcpListLoading,
    myMcpListData: mcpListDataMock,
    mcpSearchVal,
    onMcpInputChange,

    handelMcpCardClick,
    onMcpInputSearch,

    pagination,
    onPageChange,
    onShowSizeChange,
  };
}
