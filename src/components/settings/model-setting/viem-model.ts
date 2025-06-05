import { useCallback, useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { IModelPathRes } from '@/types';

export function useModelSetting() {
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');

  // 模型存储路径弹窗
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  const onModalPathChangeSuccess = useCallback(() => {
    fetchModelPath();
  }, []);

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

  useEffect(() => {
    fetchModelPath();
  }, []);

  return {
    modelPath,
    modalPathVisible,
    onModelPathVisible,
    onModalPathChangeSuccess,
  };
}
