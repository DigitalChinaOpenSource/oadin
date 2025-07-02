import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useSearchParams } from 'react-router-dom';

interface PostParamsType {
  keyword?: string;
  page: number;
  size: number;
}

export const useMcpTools = ({ id }: { id?: string }) => {
  const [searchParams] = useSearchParams();
  const serviceId = id ?? searchParams.get('serviceId');
  const [mcpTolls, setMcpTolls] = useState<Record<string, any>[]>([]);
  // const [toolsTotal, setToolsTotal] = useState<number>(0);
  const [postParams, setPostParams] = useState<PostParamsType>({ keyword: '', page: 1, size: 10 });
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 5,
    total: 0,
  });

  // 获取工具列表
  const { loading: toolsLoading, run: getTolls } = useRequest(
    async () => {
      if (serviceId) {
        const data = await httpRequest.post(`/mcp/tools/${serviceId}`, postParams);
        if (!data) throw new Error('获取工具函数列表失败');
        return data;
      } else {
        return {};
      }
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('工具函数列表===>', data);
        setMcpTolls(data.list);
        setPagination({ ...pagination, total: data.total });
      },
      onError: (error) => {
        console.error('获取工具函数列表失败:', error);
      },
    },
  );

  // 处理开启关闭工具时的Loading
  const handleToolLoading = (tool: Record<string, any>, checked: boolean, loading: boolean) => {
    setMcpTolls((prev) => {
      return prev.map((item) => {
        if (item.name === tool.name) {
          return {
            ...item,
            enabled: checked,
            changTollLoading: loading,
          };
        }
        return item;
      });
    });
  };

  // 开启关闭工具
  const { params, run: changeTollStatus } = useRequest(
    async (tool, checked) => {
      handleToolLoading(tool, tool.enabled, true);
      return await httpRequest.put(`/mcp/setup`, {
        mcpId: serviceId,
        enabled: checked,
        toolId: tool.id,
      });
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('修改工具结果===>', data);
        handleToolLoading(params[0], params[1], false);
      },
      onError: (error) => {
        console.error('修改工具失败:', error);
        handleToolLoading(params[0], params[0].enabled, false);
      },
    },
  );

  // 改变分页相关
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

  // 搜索框输入变化
  const handleSearchChange = (keyword: string) => {
    console.log('keyword', keyword);
    if (keyword === postParams.keyword) return;
    setPagination({ ...pagination, current: 1 });
    setPostParams({
      ...postParams,
      page: 1,
      keyword,
    });
  };

  // const postParamsChange = (params: PostParamsType) => {
  //   setPostParams(params);
  // };

  useEffect(() => {
    getTolls();
  }, [serviceId, postParams]);

  return {
    toolsLoading,
    // toolsTotal,
    mcpTolls,
    pagination,
    handlePageChange,
    changeTollStatus,
    handleSearchChange,
  };
};
