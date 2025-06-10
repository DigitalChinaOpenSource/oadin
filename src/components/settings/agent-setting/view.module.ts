import { useSettingsViewModel } from '@/components/settings/view.module.ts';
import { useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { AgentSettingFormValues } from '@/components/settings/agent-setting/index.tsx';

export function useAgentSettingViewModel() {
  // 是否开启代理设置
  const [agentChecked, setAgentChecked] = useState<boolean>(false);
  // 代理服务器配置
  const { systemProxy, setSystemProxy, fetchSettingsLoading } = useSettingsViewModel();

  useEffect(() => {
    setAgentChecked(systemProxy.enabled);
  }, [systemProxy.enabled]);

  // 修改代理启停
  const {
    params,
    loading: changeProxyLoading,
    run: changeProxy,
  } = useRequest(
    async (enabled: boolean) => {
      return await httpRequest.put('/system/proxy/switch', { enabled });
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('paramsparams', params);
        setAgentChecked(params[0] as boolean);
      },
      onError: (error) => {
        console.log('paramsparams222222', params);
        console.log('保存模型下载源地址失败', error);
      },
    },
  );

  // 保存代理设置
  const { loading: saveProxyLoading, run: saveProxy } = useRequest(
    async (params: AgentSettingFormValues) => {
      return await httpRequest.put('/system/proxy', params);
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('paramsparams', params);
      },
      onError: (error) => {
        console.log('paramsparams222222', params);
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
    saveProxy,
    saveProxyLoading,
  };
}
