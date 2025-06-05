import styles from '@/components/model-checking/index.module.scss';
import { useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';
import { IModelListContent } from '@/components/model-manage-tab/model-list-content';

export const ModelCheckingNodata = () => {
  const vmProps: IModelListContent = {
    modelSearchVal: '',
    modelSourceVal: 'local',
    onModelSearch: () => {},
  };
  // 获取当前的列表中的全部数据
  const vm = useViewModel(vmProps);

  const filterModelList = useMemo(() => {
    return vm.pagenationData?.length > 0 ? vm.pagenationData.splice(0, 2) : [];
  }, [vm.pagenationData]);
  return (
    <>
      <div className={styles.recommendText}>当前暂无可体验的模型，请先下载后，再进行体验</div>
      <ModelList
        {...vmProps}
        isSelectable={true}
        grid={{ gutter: 16, column: 2 }}
        dataSource={filterModelList}
        pagination={false}
      />
    </>
  );
};
