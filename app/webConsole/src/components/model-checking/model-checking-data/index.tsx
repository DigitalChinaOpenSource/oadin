import styles from '@/components/model-checking/index.module.scss';
import { IMyModelListViewModel } from '../view-model.ts';
import { useEffect, useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';
import { ISelectedDialogProps } from '@/components/choose-model-dialog';
import { selectedModelType } from '@/store/useSelectedModel.ts';

export interface ModelCheckingHasdata extends ISelectedDialogProps {
  vm: IMyModelListViewModel;
  handleAutoSetModel: (data: selectedModelType) => void;
}

export const ModelCheckingHasdata = (props: ModelCheckingHasdata) => {
  const { vm, handleAutoSetModel, selectedStateModel } = props;

  const filterModelList = useMemo(() => {
    if (!vm.modelListData?.length) {
      return [];
    }

    // 使用slice而不是splice，避免修改原始数组
    return vm.modelListData.slice(0, 2);
  }, [vm.modelListData]);

  useEffect(() => {
    // 只有在没有选中模型的情况下才自动选择第一个
    if (filterModelList.length > 0 && !selectedStateModel) {
      const modelData = filterModelList.find((item: selectedModelType) => !!item?.can_select);
      if (modelData) {
        handleAutoSetModel(modelData as selectedModelType);
      }
    }
  }, [filterModelList, selectedStateModel, handleAutoSetModel]);

  const grid = () => {
    if (filterModelList?.length > 1) {
      return { gutter: 16, column: 2 };
    }
    return { gutter: 16, column: 1 };
  };

  return (
    <>
      <div className={styles.recommendText}>为你推荐以下模型</div>
      <ModelList
        {...props}
        selectedStateModel={selectedStateModel}
        selectVms={vm}
        isSelectable={true}
        grid={grid()}
        dataSource={filterModelList}
        pagination={false}
      />
    </>
  );
};
