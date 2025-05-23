import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useSearchParams } from 'react-router-dom';
interface PostParamsType {
  keyword?: string;
  page: number;
  size: number;
}

export const useMcpTools = () => {
  const [searchParams] = useSearchParams();
  const serviceId = searchParams.get('serviceId');
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
      // await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.post(`/mcp/${serviceId}/clients`, postParams, { baseURL: '/api' });
      if (!data) throw new Error('获取工具函数列表失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('工具函数列表===>', data);
        setMcpTolls(data.list);
        // setToolsTotal(data.total);
        setPagination({ ...pagination, total: data.total });
      },
      onError: (error) => {
        console.error('获取工具函数列表失败:', error);
        const testData = {
          total: 2,
          list: [
            {
              description: "Execute a terminal command with timeout. Command will continue running in background if it doesn't complete within timeout.",
              inputSchema: {
                $schema: 'http://json-schema.org/draft-07/schema#',
                additionalProperties: false,
                properties: {
                  command: {
                    type: 'string',
                    required: true,
                  },
                  timeout_ms: {
                    type: 'number',
                    required: false,
                  },
                },
                type: 'object',
              },
              name: 'execute_command',
              server: 'filesystem',
              tool: 'execute_command',
              tags: ['123', '456'],
              enabled: true,
            },
            {
              description: 'List all currently blocked commands.',
              inputSchema: {
                properties: {},
                type: 'object',
              },
              name: 'list_blocked_commands',
              server: 'filesystem',
              tool: 'list_blocked_commands',
              tags: [],
              enabled: false,
            },
          ],
        };
        setMcpTolls(testData.list);
        // setToolsTotal(testData.total);
        setPagination({ ...pagination, total: testData.total });
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
      const data = await httpRequest.put(`/mcp/tool/${serviceId}/enabled`, { checked }, { baseURL: '/api' });
      // if (!data) throw new Error('获取工具函数列表失败');
      return data;
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
