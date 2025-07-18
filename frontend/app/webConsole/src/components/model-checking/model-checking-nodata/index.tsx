import { IModelListContent, IUseViewModel, useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { useMemo } from 'react';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';
import styles from '../index.module.scss';

export const ModelCheckingNodata = () => {
  const vmProps: IModelListContent = {
    modelSearchVal: '',
    modelSourceVal: 'local',
    onModelSearch: () => {},
    pageType: 1,
  };
  // 获取当前的列表中的全部数据
  const vmContent: IUseViewModel = useViewModel(vmProps);
  const filterModelList = useMemo(() => {
    return vmContent.pagenationData?.length > 0 ? vmContent.pagenationData.splice(0, 2) : [];
  }, [vmContent.pagenationData]);
  return (
    <div className={styles.modelCheckNoDataList}>
      <div className={styles.recommendText}>当前暂无可体验的模型，请先下载后，再进行体验</div>
      <ModelList
        vmContent={vmContent}
        isSelectable={true}
        grid={{ gutter: 16, column: 2 }}
        dataSource={filterModelList}
        pagination={false}
      />
    </div>
  );
};
