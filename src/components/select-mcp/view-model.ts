// import { useRequest } from 'ahooks';
// import { httpRequest } from '@/utils/httpRequest.ts';
// import { IMcpListData, IMcpListRequestParams } from '@/components/mcp-manage/mcp-square-tab/types.ts';
// import { useState } from 'react';
//
// export function useViewModel() {
//   // 查询列表所需的参数
//   const [postParams] = useState<IMcpListRequestParams>({
//     keyword: '',
//     page: 1,
//     size: 12,
//   });
//   // 获取 mcp 列表
//   const { loading: mcpListLoading, run: fetchMcpList } = useRequest(
//     async () => {
//       return await httpRequest.post<IMcpListData>('/mcp/search', postParams);
//     },
//     {
//       manual: true,
//       onSuccess: (data) => {
//         console.log('fetchMcpList===>', data);
//       },
//       onError: (error) => {
//         console.error('获取模型列表失败:', error);
//       },
//     },
//   );
//   const onSearch = ({ page, size, keyword }) => {};
//   return {
//     fetchMcpList,
//     mcpListLoading,
//   };
// }
