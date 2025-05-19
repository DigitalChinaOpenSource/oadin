import { Button, Tooltip, Select, List } from 'antd';
import styles from './index.module.scss';
import GeneralCard from '@/components/general-card';
import ModelPathModal from '../modelpath-modal';
import ModelAuthorizeModal from '../model-authorize-modal';
import ModelDetailModal from '../model-detail-modal';
import { useViewModel } from './view-model';
import { SettingIcon, FailedIcon, LoadingIcon } from '../../icons';
import { IModelSourceType } from '@/types';

export interface IModelListContent {
  modelSearchVal: string;
  modelSourceVal: IModelSourceType;
  onModelSearch: (val: string) => void;
  onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelListContent(props: IModelListContent) {
  const vm = useViewModel(props);
  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
  ];
  return (
    <div className={styles.modelListContent}>
      <div className={styles.contentContainer}>
        <div className={styles.titlepath}>
          <Select
            className={styles.title}
            options={options}
            value={vm.modelSourceVal}
            placeholder="请选择模型来源"
            variant="borderless"
            onChange={vm.onModelSourceChange}
          />
          <Tooltip title="/Users/lc/Library/Application\ Support/">
            <Button
              className={styles.changePath}
              type="text"
              onClick={vm.onModelPathVisible}
            >
              <SettingIcon />
              修改存储路径
            </Button>
            {/* 修改路径失败提示 */}
            {/* <span className={styles.changeFailed}>
              <FailedIcon fill='#ff6e38'/>
            </span> */}
            {/* <Button className={styles.changePath} type="text">
              <LoadingIcon />
              <span className={styles.isChangingText}>正在修改至新的存储路径</span>
            </Button> */}
          </Tooltip>
        </div>

        <div className={styles.modelCardList}>
          {vm.pagenationData.length > 0 ? (
            <List
              grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
              dataSource={vm.pagenationData}
              pagination={
                vm.pagenationData.length > 0 && {
                  className: styles.pagination,
                  align: 'end',
                  ...vm.pagination,
                  pageSizeOptions: [12, 24, 48, 96],
                  showSizeChanger: true,
                  onChange: vm.onPageChange,
                  onShowSizeChange: vm.onShowSizeChange,
                }
              }
              renderItem={(item) => (
                <List.Item>
                  <GeneralCard
                    modelData={item}
                    onCardClick={vm.onDetailModalVisible}
                    onModelAuthVisible={vm.onModelAuthVisible}
                    onDownloadConfirm={vm.onDownloadConfirm}
                    onDeleteConfirm={vm.onDeleteConfirm}
                  />
                </List.Item>
              )}
            />
          ) : (
            <div className={styles.noData}>
              {/* <div className={styles.noDataIcon}>
                  
                </div> */}
              <div className={styles.noDataText}>暂无相关模型</div>
            </div>
          )}
        </div>
      </div>
      {/* 模型路径弹窗 */}
      {vm.modalPathVisible && (
        <ModelPathModal
          modalPath={vm.modelPath}
          onModalPathClose={vm.onModelPathVisible}
        />
      )}
      {/* 配置授权弹窗 */}
      {vm.modelAuthVisible && (
        <ModelAuthorizeModal
          modelDataItem={vm.selectModelData}
          modelAuthType={vm.modelAuthType}
          onModelAuthVisible={vm.onModelAuthVisible}
          onModelAuthSuccess={vm.onModelAuthSuccess}
        />
      )}
      {/* 模型详情弹窗 */}
      {vm.isDetailVisible && (
        <ModelDetailModal
          onDetailModalVisible={vm.onDetailModalVisible}
          selectModelData={vm.selectModelData}
        />
      )}
    </div>
  );
}
