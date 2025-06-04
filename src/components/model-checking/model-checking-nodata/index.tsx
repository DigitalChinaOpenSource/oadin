import styles from '@/components/model-checking/index.module.scss';
import { List } from 'antd';
import GeneralCard from '@/components/model-manage-tab/model-list-content/general-card';
import { useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { useMemo } from 'react';

export const ModelCheckingNodata = () => {
  // 获取当前的列表中的全部数据
  const vm = useViewModel({
    modelSearchVal: '',
    modelSourceVal: 'local',
    onModelSearch: () => {},
  });

  const filterModelList = useMemo(() => {
    return vm.pagenationData?.length > 0 ? vm.pagenationData.splice(0, 2) : [];
  }, [vm.pagenationData]);
  return (
    <>
      <div className={styles.recommendText}>当前暂无可体验的模型，请先下载后，再进行体验</div>
      <div>
        <List
          grid={{ gutter: 16, column: 2 }}
          dataSource={filterModelList}
          renderItem={(item) => (
            <List.Item>
              <GeneralCard
                modelData={item}
                modelSourceVal={vm.modelSourceVal}
                onCardClick={vm.onDetailModalVisible}
                onModelAuthVisible={vm.onModelAuthVisible}
                onDownloadConfirm={vm.onDownloadConfirm}
                onDeleteConfirm={vm.onDeleteConfirm}
              />
            </List.Item>
          )}
        />
      </div>
    </>
  );
};
