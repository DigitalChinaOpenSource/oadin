import { Button, Tooltip, Pagination, Col, Row } from 'antd';
import styles from './index.module.scss';
import ModelCard from './model-card';
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
}

export default function ModelListContent(props: IModelListContent) {
  const vm = useViewModel(props);
  return (
    <div className={styles.modelListContent}>
      <div className={styles.contentContainer}>
        <div className={styles.titlepath}>
          <div className={styles.title}>模型列表</div>
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
            <Row gutter={[16, 16]}>
              {Array.isArray(vm.pagenationData) &&
                vm.pagenationData.map((item, index) => {
                  return (
                    <Col
                      xs={24}
                      sm={24}
                      md={24}
                      lg={24}
                      xl={12}
                      span={4}
                      key={index}
                    >
                      <ModelCard
                        modelData={item}
                        onDetailModalVisible={vm.onDetailModalVisible}
                        onModelAuthVisible={vm.onModelAuthVisible}
                        onDownloadConfirm={vm.onDownloadConfirm}
                        onDeleteConfirm={vm.onDeleteConfirm}
                      />
                    </Col>
                  );
                })}
            </Row>
          ) : (
            <div className={styles.noData}>
              {/* <div className={styles.noDataIcon}>
                  
                </div> */}
              <div className={styles.noDataText}>暂无相关模型</div>
            </div>
          )}
        </div>
        {vm.pagenationData.length > 0 && (
          <Pagination
            className={styles.pagination}
            align="end"
            {...vm.pagination}
            pageSizeOptions={[6, 10, 30, 50]}
            showSizeChanger
            onChange={vm.onPageChange}
            onShowSizeChange={vm.onShowSizeChange}
          />
        )}
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
          modelAuthorize={vm.modelAuthorize}
          onModelAuthVisible={vm.onModelAuthVisible}
        />
      )}
      {/* 模型详情弹窗 */}
      {vm.isDetailVisible && <ModelDetailModal onDetailModalVisible={vm.onDetailModalVisible} />}
    </div>
  );
}
