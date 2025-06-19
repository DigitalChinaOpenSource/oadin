import styles from '@/components/model-manage-tab/model-list-content/index.module.scss';
import { List, Radio } from 'antd';
import type { PaginationConfig } from 'antd/es/pagination';
import type { ListGridType } from 'antd/es/list';
import GeneralCard from '@/components/model-manage-tab/model-list-content/general-card';
import noDataSvg from '@/components/icons/no-data.svg';
import { IUseViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import useSelectedModelStore from '@/store/useSelectedModel';
import { IModelDataItem } from '@/types';
import { ISelectedDialogProps } from '@/components/choose-model-dialog';

export interface IModelList extends ISelectedDialogProps {
  pagination?: PaginationConfig | false;
  dataSource?: any[];
  grid?: ListGridType;
  isSelectable?: boolean;
  vmContent?: IUseViewModel;
  selectVms?: any;
}

export const ModelList = (props: IModelList) => {
  const { vmContent, selectVms, dataSource, selectedStateModel } = props;
  const { selectedModel } = useSelectedModelStore();
  const renderVmList = () => {
    let content = (
      <div className={styles.noData}>
        <div className={styles.noDataIcon}>
          <img
            src={noDataSvg}
            alt="no-data"
          />
        </div>
        <div className={styles.noDataText}>暂无匹配的模型</div>
      </div>
    );
    if (vmContent && vmContent.pagenationData.length > 0) {
      content = (
        <Radio.Group
          value={selectedStateModel?.id}
          style={{ width: '100%' }}
        >
          <List
            grid={props?.grid}
            dataSource={dataSource ?? vmContent.pagenationData}
            pagination={typeof props?.pagination === 'undefined' ? vmContent?.pagination : props?.pagination}
            renderItem={(item) => (
              <List.Item>
                <GeneralCard
                  {...props}
                  mine={vmContent?.mine}
                  isSelectable={props.isSelectable}
                  modelData={item}
                  modelSourceVal={vmContent.modelSourceVal}
                  onCardClick={vmContent.onDetailModalVisible}
                  onModelAuthVisible={vmContent.onModelAuthVisible}
                  onDownloadConfirm={vmContent.onDownloadConfirm}
                  onDeleteConfirm={vmContent.onDeleteConfirm}
                />
              </List.Item>
            )}
          />
        </Radio.Group>
      );
    } else if (selectVms) {
      content = (
        <Radio.Group value={selectedModel?.id}>
          <List
            grid={props?.grid}
            dataSource={props.dataSource}
            pagination={false}
            renderItem={(item: IModelDataItem) => (
              <List.Item>
                <GeneralCard
                  isSelectable={props.isSelectable}
                  modelData={item}
                  modelSourceVal={'local'}
                />
              </List.Item>
            )}
          />
        </Radio.Group>
      );
    }
    return content;
  };
  return <div className={styles.modelCardList}>{renderVmList()}</div>;
};
