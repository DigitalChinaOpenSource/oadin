import styles from './index.module.scss';
import ModelAuthorizeModal from '../model-authorize-modal';
import ModelDetailModal from '../model-detail-modal';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import { IModelList, ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';

export default function ModelListContent(props: IModelList) {
  const { vmContent } = props;
  return (
    vmContent && (
      <>
        {vmContent.modelSupportLoading ? (
          <div className={styles.loading}>
            <img
              src={realLoadingSvg}
              alt="loading"
            />
          </div>
        ) : (
          <div className={styles.modelListContent}>
            <div className={styles.contentContainer}>
              <ModelList
                isSelectable={props.isSelectable}
                vmContent={vmContent}
                pagination={
                  vmContent.modelListStateData.length > 12 && {
                    className: styles.pagination,
                    align: 'end',
                    ...vmContent.pagination,
                    pageSizeOptions: [12, 24, 48, 96],
                    showSizeChanger: true,
                    onChange: vmContent.onPageChange,
                    onShowSizeChange: vmContent.onShowSizeChange,
                  }
                }
                grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
              />
            </div>
            {/* 配置授权弹窗 */}
            {vmContent.modelAuthVisible && (
              <ModelAuthorizeModal
                modelDataItem={vmContent.selectModelData}
                modelAuthType={vmContent.modelAuthType}
                onModelAuthVisible={vmContent.onModelAuthVisible}
                onModelAuthSuccess={vmContent.onModelAuthSuccess}
              />
            )}
            {/* 模型详情弹窗 */}
            {vmContent.isDetailVisible && (
              <ModelDetailModal
                modelSourceVal={vmContent.modelSourceVal}
                onDetailModalVisible={vmContent.onDetailModalVisible}
                selectModelData={vmContent.selectModelData}
              />
            )}
          </div>
        )}
      </>
    )
  );
}
