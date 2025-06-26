import ModelManageTab from '@/components/model-manage-tab';
import React from 'react';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { IUseSearchViewModelReturn, useViewModel } from '@/components/model-manage-tab/view-model.ts';
import { ISelectedDialogProps } from '@/components/choose-model-dialog/index.tsx';

export const MyModel = (props: ISelectedDialogProps) => {
  const vmSearch: IUseSearchViewModelReturn = useViewModel();
  // 获取接口数据内容
  const vmContent = useViewModelContent({
    onModelSearch: vmSearch.onModelSearch,
    modelSearchVal: vmSearch.modelSearchVal,
    modelSourceVal: vmSearch.modelSourceVal,
    mine: true,
  });
  return (
    <ModelManageTab
      {...props}
      vmContent={vmContent}
      vmSearch={vmSearch}
      currentTab="my-models"
      isDialog={props.isDialog}
    />
  );
};
