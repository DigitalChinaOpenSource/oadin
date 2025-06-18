import ModelManageTab from '@/components/model-manage-tab';
import React from 'react';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { IUseSearchViewModelReturn, useViewModel } from '@/components/model-manage-tab/view-model.ts';

export const MyModel = (props: { isDialog?: boolean }) => {
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
      vmContent={vmContent}
      vmSearch={vmSearch}
      key="my-models"
      isDialog={props.isDialog}
    />
  );
};
