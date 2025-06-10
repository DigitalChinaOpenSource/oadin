import { useSettingsViewModel } from '@/components/settings/view-module';
import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';

export function useAgentSettingViewModel() {
  // 是否开启代理设置
  const [agentChecked, setAgentChecked] = useState(false);
  // 代理服务器配置
  const { systemProxy, setSystemProxy, fetchSettingsLoading } = useSettingsViewModel();

  useEffect(() => {
    setAgentChecked(systemProxy.enabled);
  }, [systemProxy.enabled]);

  // 修改代理启停
  const { loading: changeProxyLoading, run: changeProxy } = useRequest(
    async (enabled) => {
      return await httpRequest.put('/system/proxy/switch', { enabled });
    },
    {
      manual: true,
      onSuccess: (data) => {
        setAgentChecked(true);
      },
      onError: (error) => {
        console.log('保存模型下载源地址失败', error);
      },
    },
  );

  return {
    agentChecked,
    setAgentChecked,
    systemProxy,
    changeProxy,
    changeProxyLoading,
  };
}
