import { useCallback, useEffect, useState } from 'react';
import { message } from 'antd';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IModelPathRes } from '@/types';
import { IModelPathSpaceRes } from '@/components/model-manage-tab/types.ts';
import { useSettingsViewModel } from '@/components/settings/view-module';
import useModelPathChangeStore from '@/store/useModelPathChangeStore.ts';

export function useModelSetting() {
  const { setMigratingStatus } = useModelPathChangeStore();
  // 获取模型下载源地址
  const { ollamaRegistry, fetchSettingsLoading } = useSettingsViewModel();
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');
  // 当前路径的空间信息
  const [currentPathSpace, setCurrentPathSpace] = useState<IModelPathSpaceRes>({} as IModelPathSpaceRes);
  // 模型下载源地址
  const [modelDownUrl, setModelDownUrl] = useState<string>('');
  // 正在进行修改的模型路径
  const [changingModelPath, setChangingModelPath] = useState<string>('');

  // 模型存储路径弹窗
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  // 获取模型存储路径
  const { run: fetchModelPath } = useRequest(
    async () => {
      const res = await httpRequest.get<IModelPathRes>('/control_panel/model/filepath');
      return res || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setModelPath(data?.path || '');
      },
      onError: (error) => {
        console.error('获取模型存储路径失败:', error);
        setModelPath('测试路径' + Math.random());
      },
    },
  );

  // 获取当前路径的空间信息
  const { run: onCheckPathSpace } = useRequest(
    async (path: string) => {
      const data = await httpRequest.get<IModelPathSpaceRes>('/control_panel/path/space', { path });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setCurrentPathSpace(data);
      },
      onError: (error) => {
        setCurrentPathSpace({} as IModelPathSpaceRes);
      },
    },
  );

  // 保存模型下载源地址
  const { loading: changeModelDownUrlLoading, run: changeModelDownUrl } = useRequest(
    async (url: string) => {
      const data = await httpRequest.put('/system/registry', { url });
      return data || url;
    },
    {
      manual: true,
      onSuccess: (data) => {
        message.success('模型下载源地址修改成功');
      },
      onError: (error) => {
        console.log('保存模型下载源地址失败', error);
      },
    },
  );

  // 修改模型存储路径
  const { run: onChangeModelPath } = useRequest(
    async (params: { source_path: string; target_path: string }) => {
      const data = await httpRequest.post('/control_panel/model/filepath', params);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        message.success('模型存储路径修改成功');
        fetchModelPath();
        setMigratingStatus('init');
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error(error?.message || '模型存储路径修改失败');
        }
        setMigratingStatus('failed');
      },
    },
  );

  useEffect(() => {
    setModelDownUrl(ollamaRegistry);
  }, [ollamaRegistry]);

  return {
    fetchModelPath,
    onChangeModelPath,
    modelPath,
    modalPathVisible,
    onModelPathVisible,
    setCurrentPathSpace,
    currentPathSpace,
    onCheckPathSpace,
    changeModelDownUrl,
    changeModelDownUrlLoading,
    modelDownUrl,
    changingModelPath,
    setChangingModelPath,
  };
}
