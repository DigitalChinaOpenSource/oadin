import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useEffect } from 'react';

export function useUpdateHistory(open: boolean) {
  const { loading: historyLoading, run: fetchUpdateHistory } = useRequest(
    async () => {
      await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get('/service/update/history');
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        if (!data) return;
      },
      onError: (error) => {
        console.error('获取更新日志失败:', error);
      },
    },
  );

  useEffect(() => {
    if (!open) return;
    fetchUpdateHistory();
  }, [open]);

  return {
    // 可以在这里添加需要暴露给外部的函数或状态
    historyLoading,
  };
}
