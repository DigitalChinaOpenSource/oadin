import { useState, useEffect, useRef } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData } from './types';
import { mcpListDataMock } from './constant';
import { useNavigate } from 'react-router-dom';

export function useViewModel() {
  const navigate = useNavigate();

  const [mcpListData, setMcpListData] = useState([]);
  // const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);
  // 过滤器是否折叠了
  const [collapsed, setCollapsed] = useState(false);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 12,
    total: 20,
  });

  // 标签测试数据
  const tagsData = [
    {
      category: '图像处理',
      tags: ['增强', '压缩', '分割'],
    },
    {
      category: '文本分析',
      tags: ['情感分析', '关键词提取'],
    },
  ];

  // 初始化标签数据
  const initData = tagsData.reduce((acc: Record<string, any>, item) => {
    acc[item.category] = [];
    return acc;
  }, {});

  // 标签选中结果
  const [checkedValues, setCheckedValues] = useState(initData);
  // 查询列表所需的参数
  const [postParams, setPostParams] = useState<IMcpListRequestParams>({
    keyword: '',
    page: 1,
    size: 12,
    category: Object.keys(checkedValues).filter((key) => checkedValues[key].length > 0),
    tags: [],
  });

  // 获取 mcp 列表
  const { loading: mcpListLoading, run: fetchMcpList } = useRequest(
    async () => {
      const data = await httpRequest.post<IMcpListData>('/mcp/search', postParams, { baseURL: '/api' });
      return data?.data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('fetchMcpList===>', data);
        // setMcpListData(data?.list);
        // setPagination(
        //   ...pagination,
        //   total: data?.total,
        // )
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  const handelMcpCardClick = (serviceId: string | number) => {
    navigate(`/mcp-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };

  // 搜索框搜索
  const onMcpInputSearch = (inputSearchVal: string) => {
    console.log('inputSearchVal===>', inputSearchVal);
    if (inputSearchVal === postParams.keyword) return;
    setPagination({ ...pagination, current: 1 });
    setPostParams({
      ...postParams,
      page: 1,
      keyword: inputSearchVal,
      category: Object.keys(checkedValues).filter((key) => checkedValues[key].length > 0),
      tags: Object.values(checkedValues).flat(),
    });
  };

  const handlePageChange = (page: number, pageSize: number) => {
    console.log('page', page);
    console.log('pageSize', pageSize);
    setPagination({
      ...pagination,
      current: page,
      pageSize,
    });
    setPostParams({
      ...postParams,
      page,
      size: pageSize,
    });
  };

  // 标签选择改变
  const handleTagsChange = (category: string, list: any) => {
    const updatedCheckedValues = {
      ...checkedValues,
      [category]: list,
    };

    setCheckedValues(updatedCheckedValues);

    // 更新分页并设置新的查询参数
    setPagination({
      ...pagination,
      current: 1,
    });

    const updatedPostParams = {
      ...postParams,
      page: 1,
      category: Object.keys(updatedCheckedValues).filter((key) => updatedCheckedValues[key].length > 0),
      tags: Object.values(updatedCheckedValues).flat(),
    };

    setPostParams(updatedPostParams);
  };

  const [searchVal, setSearchVal] = useState('');

  useEffect(() => {
    fetchMcpList();
  }, [postParams]);

  return {
    mcpListLoading,
    mcpListData: mcpListDataMock,
    handelMcpCardClick,
    onMcpInputSearch,
    collapsed,
    setCollapsed,
    pagination,
    handlePageChange,
    handleTagsChange,
    tagsData,
    checkedValues,
    searchVal,
    setSearchVal,
  };
}
