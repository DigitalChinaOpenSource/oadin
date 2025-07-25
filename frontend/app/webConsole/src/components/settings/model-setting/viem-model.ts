import { useCallback, useEffect, useState, useRef } from 'react';
import { message } from 'antd';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IModelPathRes } from '@/types';
import { IModelPathSpaceRes } from '@/components/model-manage-tab/types.ts';
import { useSettingsViewModel } from '@/components/settings/view-module';
import { useModelPathChangeStore, IStatus } from '@/store/useModelPathChangeStore.ts';

export function useModelSetting() {
  const { setMigratingStatus, migratingStatus } = useModelPathChangeStore();
  // 获取模型下载源地址
  const { ollamaRegistry } = useSettingsViewModel();
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
  // 正在查询模型路径和磁盘空间
  const [isCheckingPathSpace, setIsCheckingPathSpace] = useState<boolean>(false);

  const prevMigratingStatusRef = useRef<IStatus>('init');

  // 模型存储路径弹窗
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  // 获取模型存储路径
  const { run: fetchModelPath } = useRequest(
    async () => {
      setIsCheckingPathSpace(true);
      const res = await httpRequest.get<IModelPathRes>('/control_panel/model/filepath');
      return res || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setModelPath(data?.path || '');
        if (data?.path) {
          onCheckPathSpace(data?.path);
        }
      },
      onError: (error) => {
        console.error('获取模型存储路径失败:', error);
        setIsCheckingPathSpace(false);
      },
    },
  );

  // 获取当前路径的空间信息
  const { runAsync: onCheckPathSpace } = useRequest(
    async (path: string) => {
      const data = await httpRequest.get<IModelPathSpaceRes>('/control_panel/path/space', { path });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        setCurrentPathSpace(data);
        setIsCheckingPathSpace(false);
        return data;
      },
      onError: (error) => {
        setCurrentPathSpace({} as IModelPathSpaceRes);
        setIsCheckingPathSpace(false);
        return error;
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
      setMigratingStatus('pending');
      const data = await httpRequest.post('/control_panel/model/filepath', params, { needModelChangeStore: true, setMigratingStatus, timeout: 3600000 });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        message.success('模型存储路径修改成功');
      },
      onError: (error: Error & { handled?: boolean }) => {
        if (!error?.handled) {
          message.error(error?.message || '模型存储路径修改失败');
        }
      },
    },
  );

  useEffect(() => {
    setModelDownUrl(ollamaRegistry);
  }, [ollamaRegistry]);

  useEffect(() => {
    // 如果是迁移中到迁移完成才执行一次调用
    if (prevMigratingStatusRef.current === 'pending' && migratingStatus === 'init') {
      fetchModelPath();
    }
    // 更新之前的状态值
    prevMigratingStatusRef.current = migratingStatus;
  }, [migratingStatus]);

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
    isCheckingPathSpace,
    setIsCheckingPathSpace,
  };
}
