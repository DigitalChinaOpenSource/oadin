import styles from '@/components/model-checking/index.module.scss';
import { IUseViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';

interface ModelCheckingHasdata {
  vm: IUseViewModel;
}

export const ModelCheckingHasdata = (props: ModelCheckingHasdata) => {
  const { vm } = props;

  const filterModelList = useMemo(() => {
    return vm.pagenationData?.length > 0 ? vm.pagenationData.splice(0, 2) : [];
  }, [vm.pagenationData]);
  return (
    <>
      <div className={styles.recommendText}>为你推荐一下模型</div>
      <ModelList
        vm={vm}
        isSelectable={true}
        grid={{ gutter: 16, column: 2 }}
        dataSource={filterModelList}
        pagination={false}
      />
    </>
  );
};
