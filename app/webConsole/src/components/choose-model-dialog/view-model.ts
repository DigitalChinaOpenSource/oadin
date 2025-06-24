import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
export function useViewModel() {
  // 切换会话模型
  const { run: fetchChangeModel } = useRequest(async (params: { sessionId: string; modelId: string }) => {
    if (!params?.sessionId || !params.modelId) {
      return {};
    }
    const data = await httpRequest.post('/playground/session/model', {
      ...params,
    });
    return data?.data || {};
  });

  // 选择模型后需要将所选择的通知奥丁
  const { run: fetchChooseModelNotify } = useRequest(async (params: { service_name: string; local_provider?: string; remote_provider?: string }) => {
    if (!params.service_name) return;
    const data = await httpRequest.put('/service', {
      ...params,
    });
    return data || {};
  });
  return { fetchChangeModel, fetchChooseModelNotify };
}
