// 选择MCP相关的数据处理的函数类
import { message } from 'antd';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { getMessageByMcp } from '@/i18n';

// 定义MCP列表ID参数接口
export interface IMcpListIds {
  ids: string[];
}

// 定义MCP更新结果接口
export interface IMcpUpdateResult {
  startList: IMcpListItem[];
  stopList: IMcpListItem[];
}

/**
 * 检查MCP列表长度是否超过限制
 * @param mcpListLength MCP列表长度
 * @returns 如果长度小于等于4则返回true，否则返回false
 */
export const checkMcpLength = (mcpListLength: number): boolean => {
  if (mcpListLength > 4) {
    message.warning(
      getMessageByMcp('maxSelectMcp', {
        msg: '为保障服务稳定运行与优质体验，建议您选择的MCP工具不要超过5个。',
      }),
    );
    return false;
  }
  return true;
};

/**
 * MCP远程操作相关的钩子函数
 * @returns 返回启动和停止MCP的函数
 */
export function useSelectRemoteHelper() {
  // 启动MCP服务
  const { run: startMcps, loading: startLoading } = useRequest(
    async (params: IMcpListIds) => {
      return await httpRequest.post('/mcp/client/start', params);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.info('启动MCP服务成功:', data);
      },
      onError: (error) => {
        console.error('启动MCP服务失败:', error);
        message.error('启动MCP服务失败');
      },
    },
  );

  // 停止MCP服务
  const { run: stopMcps, loading: stopLoading } = useRequest(
    async (params: IMcpListIds) => {
      return await httpRequest.post('/mcp/client/stop', params);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.info('停止MCP服务成功:', data);
      },
      onError: (error) => {
        console.error('停止MCP服务失败:', error);
        message.error('停止MCP服务失败');
      },
    },
  );

  return {
    startMcps,
    stopMcps,
    startLoading,
    stopLoading,
  };
}

/**
 * 更新MCP列表，对比新旧列表并启动/停止相应的MCP
 * @param oldMcpList 旧的MCP列表
 * @param newMcpList 新的MCP列表
 * @returns 返回启动和停止的MCP列表
 */
export const updateMcp = (oldMcpList: IMcpListItem[], newMcpList: IMcpListItem[]): IMcpUpdateResult => {
  // 将MCP列表转换为ID Map方便比较
  const oldMcpMap = new Map(oldMcpList.map((item) => [item.id.toString(), item]));
  const newMcpMap = new Map(newMcpList.map((item) => [item.id.toString(), item]));

  // 找出新增的MCP (newMcpList中有但oldMcpList中没有的)
  const startList: IMcpListItem[] = newMcpList.filter((item) => !oldMcpMap.has(item.id.toString()));

  // 找出移除的MCP (oldMcpList中有但newMcpList中没有的)
  const stopList: IMcpListItem[] = oldMcpList.filter((item) => !newMcpMap.has(item.id.toString()));

  // 返回启动和停止的MCP列表
  return { startList, stopList };
};
// 定义数据结构类型
interface FunctionTool {
  type: string;
  function: {
    name: string;
    description: string;
    parameters: any;
  };
}

interface McpTool {
  mcpId: string;
  tools: FunctionTool[];
}

interface McpToolsResponse {
  mcpTools: McpTool[];
}

interface IToolArguments {
  [key: string]: any; // 动态键值对
}

interface IToolParams {
  toolName: string;
  toolArgs: IToolArguments; // 动态参数对象
}

interface IToolRequest extends IToolParams {
  mcpId: string;
  toolDesc?: string;
  toolLogo?: string;
}

// 转换函数
function generateToolMap(response: McpToolsResponse): Record<string, string> {
  const result: Record<string, string> = {};

  // 遍历所有 mcpTools 条目
  for (const mcpTool of response.mcpTools) {
    const mcpId = mcpTool.mcpId;

    // 遍历当前 mcpTool 中的所有工具
    for (const tool of mcpTool.tools) {
      // 仅处理类型为 "function" 的工具
      if (tool.type === 'function') {
        const functionName = tool.function.name;
        // 将函数名称作为键，mcpId 作为值
        result[functionName] = mcpId;
      }
    }
  }

  return result;
}

export const getMcpDeatilByIds = async (ids: string[]) => {
  const data: McpToolsResponse = await httpRequest.post<McpToolsResponse>('/mcp/client/getTools', { ids });
  return data;
};

// 基于返回方法从当前选择的MCP服务中获取id等信息
export const getIdByFunction = async (functionParams: IToolParams, ids: string[]): Promise<IToolRequest> => {
  const resData = await getMcpDeatilByIds(ids);
  const searchDataMap = generateToolMap(resData);
  const mcpId = searchDataMap[functionParams.toolName];
  const idTool = resData?.mcpTools.find((item) => item.mcpId === mcpId);
  const toolDesc = idTool?.tools.find((tool) => tool.function.name === functionParams.toolName)?.function.description || '';
  const toolLogo = idTool?.tools.find((tool) => tool.function.name === functionParams.toolName)?.function.parameters?.logo || '';
  if (mcpId) {
    return {
      mcpId,
      toolName: functionParams.toolName,
      toolArgs: functionParams.toolArgs,
    };
  } else {
    return {
      mcpId: '',
      toolName: functionParams.toolName,
      toolArgs: functionParams.toolArgs,
    };
  }
};
