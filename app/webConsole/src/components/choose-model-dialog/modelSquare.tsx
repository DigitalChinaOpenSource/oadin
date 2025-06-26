import ModelManageTab from '@/components/model-manage-tab';
import React from 'react';
import { IUseSearchViewModelReturn, useViewModel } from '@/components/model-manage-tab/view-model.ts';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { ISelectedDialogProps } from '@/components/choose-model-dialog/index.tsx';

// 模型广场
export const ModelSquare = (props: ISelectedDialogProps) => {
  const vmSearch: IUseSearchViewModelReturn = useViewModel();
  // 获取接口数据内容
  const vmContent = useViewModelContent({
    onModelSearch: vmSearch.onModelSearch,
    modelSearchVal: vmSearch.modelSearchVal,
    modelSourceVal: vmSearch.modelSourceVal,
    mine: false,
  });
  return (
    <ModelManageTab
      {...props}
      vmContent={vmContent}
      vmSearch={vmSearch}
      currentTab="model-square"
      isDialog={props.isDialog}
    />
  );
};
