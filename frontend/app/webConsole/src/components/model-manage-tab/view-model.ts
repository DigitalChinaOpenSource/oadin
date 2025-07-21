import { useState } from 'react';
import { IModelSourceType } from '@/types';
export interface IUseSearchViewModelReturn {
  modelSourceVal: IModelSourceType;
  onModelSourceChange: (val: IModelSourceType) => void;
  modelSearchVal: string;
  onModelSearch: (val: string) => void;
}
export function useViewModel(): IUseSearchViewModelReturn {
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
