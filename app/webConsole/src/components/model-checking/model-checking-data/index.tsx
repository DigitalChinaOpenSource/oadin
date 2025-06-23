import styles from '@/components/model-checking/index.module.scss';
import { IMyModelListViewModel } from '../view-model.ts';
import { useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';
import { ISelectedDialogProps } from '@/components/choose-model-dialog';

export interface ModelCheckingHasdata extends ISelectedDialogProps {
  vm: IMyModelListViewModel;
}

export const ModelCheckingHasdata = (props: ModelCheckingHasdata) => {
  const { vm } = props;

  const filterModelList = useMemo(() => {
    return vm.modelListData?.length > 0 ? vm.modelListData.splice(0, 2) : [];
  }, [vm.modelListData]);

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
        selectVms={vm}
        isSelectable={true}
        grid={grid()}
        dataSource={filterModelList}
        pagination={false}
      />
    </>
  );
};
