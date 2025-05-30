import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IMcpListRequestParams, IMcpListData, IMcpListItem, IPagination, ITagsDataItem } from './types';
import { useNavigate } from 'react-router-dom';
import usePageParamsStore from '@/store/usePageParamsStore.ts';

export function useViewModel() {
  const navigate = useNavigate();
  const { setPageParams, getPageParams, setTagsDataStore, tagsDataStore } = usePageParamsStore();
  const [mcpListData, setMcpListData] = useState<IMcpListItem[]>([]);
  // const [mcpSearchVal, setMcpSearchVal] = useState({} as IMcpListRequestParams);
  // 所有的mcp服务筛选条件tags
  const [tagsData, setTagsData] = useState<ITagsDataItem[]>([{ category: '', tags: [] }]);
  // 过滤器是否折叠了
  const [collapsed, setCollapsed] = useState(false);
  const [pagination, setPagination] = useState<IPagination>({
    current: 1,
    pageSize: 12,
    total: 0,
  });

  // 标签选中结果
  const [checkedValues, setCheckedValues] = useState<Record<string, any>>({});

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
      return await httpRequest.post<IMcpListData>('/mcp/search', postParams);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('fetchMcpList===>', data);
        setMcpListData(data?.list || []);
        setPagination({
          ...pagination,
          total: data?.total || 0,
        });
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  // 获取所有mcp服务的筛选tags
  const { loading: tagsLoading, run: getTagsData } = useRequest(
    async () => {
      return await httpRequest.get('/mcp/categories');
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('tagsData===>', data);
        setTagsData(data || []);
        // 初始化标签数据
        const initData = data.reduce((acc: Record<string, any>, item: any) => {
          acc[item.category] = [];
          return acc;
        }, {});
        setCheckedValues(initData);
        setTagsDataStore(data);
      },

      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  const handelMcpCardClick = (serviceId: string | number) => {
    // 将当前页面所有参数存入store
    setPageParams({
      fromDetail: true,
      allParams: {
        pagination,
        checkedValues,
        postParams,
        searchVal,
      },
    });
    // 执行跳转
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
    console.log('handleTagsChange===>', category, list);
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

  // 清空筛选器筛选条件
  const handleClearTags = () => {
    setPagination({
      ...pagination,
      current: 1,
    });
    setCheckedValues([]);
    const updatedPostParams = {
      ...postParams,
      page: 1,
      category: [],
      tags: [],
    };
    setPostParams(updatedPostParams);
  };

  useEffect(() => {
    if (tagsDataStore && tagsDataStore.length) {
      return setTagsData(tagsDataStore);
    }
    getTagsData();
  }, []);

  useEffect(() => {
    const pageParams = getPageParams();
    if (pageParams.fromDetail) {
      setPostParams(pageParams.allParams.postParams as IMcpListRequestParams);
      setPagination(pageParams.allParams.pagination as IPagination);
      setCheckedValues(pageParams.allParams.checkedValues as Record<string, any>);
      setSearchVal(pageParams.allParams.searchVal || '');
      setPageParams({ fromDetail: false, allParams: {} });
      return;
    }

    fetchMcpList();
  }, [postParams]);

  return {
    mcpListLoading,
    mcpListData,
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
    handleClearTags,
  };
}
