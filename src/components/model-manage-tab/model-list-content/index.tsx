import { Button, Tooltip, Pagination, Col, Row } from 'antd';
import styles from './index.module.scss';
import ModelCard from './model-card';
import ModelPathModal from '../modelpath-modal';
import ModelAuthorizeModal from '../model-authorize-modal';
import ModelDetailModal from '../model-detail-modal';
import { useViewModel } from './view-model';
import {SettingIcon, FailedIcon, LoadingIcon} from '../../icons';

export default function ModelListContent() {
  const vm = useViewModel();
  const {
    modalPathVisible,
    onModelPathVisible,
    modelPath,
    modelAuthorize,
    modelAuthVisible,
    onModelAuthVisible,
    onSetModelAuthorize,
    isDetailVisible,
    onDetailModalVisible,
    modelAuthType,
  } = vm;
  return (
    <div className={styles.modelListContent}>
      <div className={styles.contentContainer}>
        <div className={styles.titlepath}>
          <div className={styles.title}>模型列表</div>
          <Tooltip title="/Users/lc/Library/Application\ Support/">
            <Button
              className={styles.changePath}
              type="text"
              onClick={onModelPathVisible}
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
          <Row gutter={[16, 16]}>
            {[1, 2, 3, 4, 5, 6, 7, 8].map((item, index) => {
              return (
                <Col
                  xs={24}
                  sm={24}
                  md={24}
                  lg={12}
                  xl={12}
                  span={4}
                  key={index}
                >
                  <ModelCard
                    onDetailModalVisible={onDetailModalVisible}
                    onModelAuthVisible={onModelAuthVisible}
                  />
                </Col>
              );
            })}
          </Row>
        </div>

        <Pagination
          className={styles.pagination}
          current={1}
          total={500}
          align="end"
          pageSizeOptions={[10, 20, 50, 100]}
          showSizeChanger
        />
      </div>
      {/* 模型路径弹窗 */}
      {modalPathVisible && (
        <ModelPathModal
          modalPath={modelPath}
          onModalPathClose={onModelPathVisible}
        />
      )}
      {/* 配置授权弹窗 */}
      {modelAuthVisible && (
        <ModelAuthorizeModal
          modelAuthType={modelAuthType}
          modelAuthorize={modelAuthorize}
          onSetModelAuthorize={onSetModelAuthorize}
          onModelAuthVisible={onModelAuthVisible}
        />
      )}
      {/* 模型详情弹窗 */}
      {isDetailVisible && (
        <ModelDetailModal onDetailModalVisible={onDetailModalVisible} />
      )}
    </div>
  );
}
