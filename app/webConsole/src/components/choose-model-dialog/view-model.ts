import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
export function useViewModel() {
  // 切换会话模型
  const { run: fetchChangeModel } = useRequest(async (params: { sessionId: string; modelId: string }) => {
    const data = await httpRequest.post('/playground/session/model', {
      ...params,
    });
    return data?.data || {};
  });
  return { fetchChangeModel };
}
