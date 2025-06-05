import styles from '@/components/model-manage-tab/model-list-content/index.module.scss';
import { List, Radio } from 'antd';
import type { PaginationConfig } from 'antd/es/pagination';
import type { ListGridType } from 'antd/es/list';
import GeneralCard from '@/components/model-manage-tab/model-list-content/general-card';
import noDataSvg from '@/components/icons/no-data.svg';
import { useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { IModelListContent } from '@/components/model-manage-tab/model-list-content/index.tsx';

export interface IModelList extends IModelListContent {
  pagination?: PaginationConfig | false;
  dataSource?: any[];
  grid?: ListGridType;
  isSelectable?: boolean;
}

export const ModelList = (props: IModelList) => {
  const vm = useViewModel(props);
  return (
    <div className={styles.modelCardList}>
      {vm.pagenationData.length > 0 ? (
        <Radio.Group value={vm?.selectModel?.id}>
          <List
            grid={props?.grid}
            dataSource={props?.dataSource ?? vm.pagenationData}
            pagination={props?.pagination ?? vm?.pagination}
            renderItem={(item) => (
              <List.Item>
                <GeneralCard
                  isSelectable={props.isSelectable}
                  modelData={item}
                  modelSourceVal={vm.modelSourceVal}
                  onCardClick={vm.onDetailModalVisible}
                  onModelAuthVisible={vm.onModelAuthVisible}
                  onDownloadConfirm={vm.onDownloadConfirm}
                  onDeleteConfirm={vm.onDeleteConfirm}
                  setSelectModel={vm.setSelectModel}
                  selectModel={vm.selectModel}
                />
              </List.Item>
            )}
          />
        </Radio.Group>
      ) : (
        <div className={styles.noData}>
          <div className={styles.noDataIcon}>
            <img
              src={noDataSvg}
              alt="no-data"
            />
          </div>
          <div className={styles.noDataText}>暂无匹配的模型</div>
        </div>
      )}
    </div>
  );
};
