import styles from '@/components/model-checking/index.module.scss';
import { IMyModelListViewModel } from '../view-model.ts';
import { useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';

export interface ModelCheckingHasdata {
  vm: IMyModelListViewModel;
}

export const ModelCheckingHasdata = (props: ModelCheckingHasdata) => {
  const { vm } = props;

  const filterModelList = useMemo(() => {
    return vm.modelListData?.length > 0 ? vm.modelListData.splice(0, 2) : [];
  }, [vm.modelListData]);
  return (
    <>
      <div className={styles.recommendText}>为你推荐一下模型</div>
      <ModelList
        selectVms={vm}
        isSelectable={true}
        grid={{ gutter: 16, column: 2 }}
        dataSource={filterModelList}
        pagination={false}
      />
    </>
  );
};
