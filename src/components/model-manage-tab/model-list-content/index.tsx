import { Button, Tooltip, List, message } from 'antd';
import styles from './index.module.scss';
import GeneralCard from '@/components/model-manage-tab/model-list-content/general-card';
import ModelPathModal from '../modelpath-modal';
import ModelAuthorizeModal from '../model-authorize-modal';
import ModelDetailModal from '../model-detail-modal';
import { useViewModel } from './view-model';
import { SettingIcon, FailedIcon } from '../../icons';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import noDataSvg from '@/components/icons/no-data.svg';
import { IModelSourceType } from '@/types';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
export interface IModelListContent {
  modelSearchVal: string;
  modelSourceVal: IModelSourceType;
  onModelSearch: (val: string) => void;
  // onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelListContent(props: IModelListContent) {
  const vm = useViewModel(props);
  const { migratingStatus } = useModelPathChangeStore();
  return (
    <>
      {vm.modelSupportLoading ? (
        <div className={styles.loading}>
          <img
            src={realLoadingSvg}
            alt="loading"
          />
        </div>
      ) : (
        <div className={styles.modelListContent}>
          <div className={styles.contentContainer}>
            <div className={styles.titlepath}>
              <div className={styles.title}>模型列表</div>
              {vm.modelSourceVal === 'local' && (
                <div>
                  <Tooltip title={vm.modelPath}>
                    {(migratingStatus === 'init' || migratingStatus === 'failed') && (
                      <Button
                        className={styles.changePath}
                        type="text"
                        onClick={vm.onModelPathVisible}
                      >
                        <SettingIcon />
                        修改存储路径
                      </Button>
                    )}
                    {migratingStatus === 'pending' && (
                      <Button
                        className={styles.changePath}
                        type="text"
                      >
                        <img
                          src={realLoadingSvg}
                          alt="loading"
                          width={20}
                        />
                        <span className={styles.isChangingText}>正在修改至新的存储路径</span>
                      </Button>
                    )}
                  </Tooltip>
                  <>
                    {migratingStatus === 'failed' && (
                      <Tooltip title="修改存储路径失败">
                        <span className={styles.changeFailed}>
                          <FailedIcon fill="#ff6e38" />
                        </span>
                      </Tooltip>
                    )}
                  </>
                </div>
              )}
            </div>

            <div className={styles.modelCardList}>
              {vm.pagenationData.length > 0 ? (
                <List
                  grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
                  dataSource={vm.pagenationData}
                  pagination={
                    vm.modelListData.length > 12 && {
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
                        modelSourceVal={vm.modelSourceVal}
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
          </div>
          {/* 模型路径弹窗 */}
          {vm.modalPathVisible && (
            <ModelPathModal
              modalPath={vm.modelPath}
              onModelPathVisible={vm.onModelPathVisible}
              onModalPathChangeSuccess={vm.onModalPathChangeSuccess}
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
              modelSourceVal={vm.modelSourceVal}
              onDetailModalVisible={vm.onDetailModalVisible}
              selectModelData={vm.selectModelData}
            />
          )}
        </div>
      )}
    </>
  );
}
