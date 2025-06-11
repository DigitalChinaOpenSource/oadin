import ModelManageTab from '@/components/model-manage-tab';
import React from 'react';
import { useViewModel } from '@/components/model-manage-tab/view-model.ts';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';

export const ModelSquare = () => {
  const vm = useViewModel();
  // 获取接口数据内容
  const vmContent = useViewModelContent({
    onModelSearch: vm.onModelSearch,
    modelSearchVal: vm.modelSearchVal,
    modelSourceVal: vm.modelSourceVal,
    mine: false,
  });
  return (
    <ModelManageTab
      vmContent={vmContent}
      key="model-square"
      isDialog={true}
    />
  );
};
