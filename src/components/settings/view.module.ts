import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import { ISettings, ISystemProxy } from '@/components/settings/types.ts';
import { httpRequest } from '@/utils/httpRequest.ts';

export function useSettingsViewModel() {
  const [ollamaRegistry, setOllamaRegistry] = useState<string>('');
  const [systemProxy, setSystemProxy] = useState<ISystemProxy>({ enabled: false });

  // 获取模型和代理相关设置
  const { loading: fetchSettingsLoading, run: fetchSettings } = useRequest(
    async () => {
      const res = await httpRequest.get<ISettings>('/system/information');
      return res || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setSystemProxy(data.systemProxy || { enabled: false });
        setOllamaRegistry(data.ollamaRegistry);
      },
      onError: (error) => {
        console.error('获取模型存储路径失败:', error);
        setOllamaRegistry('ccccc');
        setSystemProxy({ enabled: true });
      },
    },
  );

  useEffect(() => {
    fetchSettings();
  }, []);

  return {
    ollamaRegistry,
    systemProxy,
    setSystemProxy,
    fetchSettingsLoading,
    fetchSettings,
  };
}
