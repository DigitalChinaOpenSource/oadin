import ModelManageTab from '@/components/model-manage-tab';
import React from 'react';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { useViewModel } from '@/components/model-manage-tab/view-model.ts';

export const MyModel = (props: { isDialog?: boolean }) => {
  const vm = useViewModel();
  // 获取接口数据内容
  const vmContent = useViewModelContent({
    onModelSearch: vm.onModelSearch,
    modelSearchVal: vm.modelSearchVal,
    modelSourceVal: vm.modelSourceVal,
    mine: true,
  });
  return (
    <ModelManageTab
      vmContent={vmContent}
      key="my-models"
      isDialog={props.isDialog}
    />
  );
};
