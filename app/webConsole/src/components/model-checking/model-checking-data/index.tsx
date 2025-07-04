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
    console.log('ModelCheckingHasdata: filterModelList 变化，长度:', filterModelList.length);
    // if (filterModelList.length > 0) {
    //   const modelData = filterModelList.find((item: selectedModelType) => !!item?.can_select);
    //   console.log('ModelCheckingHasdata: 查找可选择的模型:', modelData ? `找到模型 ${modelData.name}` : '未找到可选择模型');
    //   if (modelData) {
    //     handleAutoSetModel(modelData as selectedModelType);
    //   }
    // }
  }, [filterModelList]);

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
