import { useRef, useMemo } from 'react';
import styles from './index.module.scss';
import { Button, Progress, Tooltip } from 'antd';
import { IModelAuth } from '../model-manage-tab/types';
import { ModelDataItem } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon } from '@/components/icons';
import TagsRender from './tags-render';

export interface IGeneralCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 模型数据
  modelData: ModelDataItem;
  onDetailModalVisible?: (visible: boolean, selectModelData: ModelDataItem) => void;
  onModelAuthVisible?: (data: IModelAuth) => void;
  onDeleteConfirm?: (modelData: ModelDataItem) => void;
  onDownloadConfirm?: (modelData: ModelDataItem) => void;
}

export default function GeneralCard(props: IGeneralCardProps) {
  const { isDetail, onDetailModalVisible, onModelAuthVisible, onDeleteConfirm, onDownloadConfirm, modelData } = props;

  // modelData?.class
  const tags = useMemo(
    () => ['深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成'],
    [modelData?.class],
  );

  const statusToText = (item: ModelDataItem) => {
    const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
    const { status, can_select } = item;
    if (status === IN_PROGRESS)
      return (
        <Button
          type="text"
          style={{ color: '#344054', padding: 'unset' }}
        >
          下载中
          <LoadingIcon />
        </Button>
      );
    if (can_select || status === COMPLETED)
      return (
        <Button
          className={styles.downloadedBtn}
          onClick={() => onDeleteConfirm?.(modelData)}
        >
          <DeleteIcon fill="#344054" />
          已下载
        </Button>
      );
    if (status === PAUSED) return '暂停';
    else if (!can_select || status === FAILED)
      return (
        <Button
          type="primary"
          onClick={() => onDownloadConfirm?.(modelData)}
        >
          <DownloadIcon />
          下载
        </Button>
      );
  };

  return (
    <div
      className={`${isDetail ? styles.generalCardDetail : styles.generalCardHover} ${styles.generalCard}`}
      onClick={() => onDetailModalVisible?.(true, modelData)}
    >
      {/* 推荐使用，定位右上角 */}
      {modelData?.is_recommended && <div className={styles.recommend}>推荐使用</div>}
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              src={modelData?.avatar}
              width={24}
            />
          </div>
          {/* 名称 */}
          <div className={styles.title}>{modelData?.name}</div>
          {/* 本地还是云端 */}
          <div className={styles.localOrCloud}>
            {modelData?.source === 'local' ? (
              <>
                <LocalIcon />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <CloudIcon />
                <div className={styles.localOrCloudText}>云端</div>
              </>
            )}
          </div>
        </div>
      </div>

      <TagsRender tags={tags} />

      <Tooltip title={modelData?.desc}>
        <div className={`${isDetail ? styles.contentWrapperDetail : styles.contentWrapper}`}>{modelData?.desc}</div>
      </Tooltip>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>深度求索</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>2025-05-19 更新</div>
        {modelData?.can_select && <div className={styles.modelStatus}>已下载</div>}
      </div>
      {Boolean(modelData?.currentDownload) && (
        <Progress
          className={styles.progress}
          percent={30}
          size="small"
          showInfo={false}
          strokeColor="#5429ff"
          strokeLinecap="butt"
        />
      )}

      <div className={styles.handlebar}>{statusToText(modelData)}</div>
    </div>
  );
}
