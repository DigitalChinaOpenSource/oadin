import { useState } from 'react';
import { IModelSourceType } from '@/types';
export function useViewModel() {
  // 模型来源
  const [modelSourceVal, setModelSourceVal] = useState<IModelSourceType>('local');
  // 模型搜索
  const [modelSearchVal, setModelSearchVal] = useState<string>('');

  const onModelSourceChange = (val: IModelSourceType) => {
    setModelSourceVal(val);
    setModelSearchVal('');
  };

  const onModelSearch = (val: string) => {
    setModelSearchVal(val);
  };

  return {
    modelSourceVal,
    onModelSourceChange,
    modelSearchVal,
    onModelSearch,
  };
}
