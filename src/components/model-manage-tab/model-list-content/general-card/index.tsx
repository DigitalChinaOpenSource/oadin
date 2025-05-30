import ReactMarkdown from 'react-markdown';
import styles from './index.module.scss';
import { Button, Tooltip } from 'antd';
import { IModelAuth } from '../../types';
import { IModelDataItem, IModelSourceType } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon, SettingIcon, ArrowClockwiseIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';

export interface IGeneralCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 模型数据
  modelData: IModelDataItem;
  modelSourceVal: IModelSourceType;
  onCardClick?: (visible: boolean, selectModelData: IModelDataItem) => void;
  onModelAuthVisible?: (data: IModelAuth) => void;
  onDeleteConfirm?: (modelData: IModelDataItem) => void;
  onDownloadConfirm?: (modelData: IModelDataItem) => void;
}

export default function GeneralCard(props: IGeneralCardProps) {
  const { isDetail, onCardClick, modelSourceVal, onDeleteConfirm, onModelAuthVisible, onDownloadConfirm, modelData } = props;
  const { isPathMigrating } = useModelPathChangeStore();
  const statusToText = (item: IModelDataItem) => {
    const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
    const { status, can_select } = item;
    if (status === IN_PROGRESS)
      return (
        <Button
          className={styles.downloadedBtn}
          icon={<LoadingIcon />}
        >
          下载中
        </Button>
      );
    if (can_select || status === COMPLETED)
      return (
        <Button
          className={styles.downloadedBtn}
          disabled={isPathMigrating}
          onClick={(e) => {
            e.stopPropagation();
            onDeleteConfirm?.(modelData);
          }}
          icon={<DeleteIcon fill="#344054" />}
        >
          删除模型
        </Button>
      );
    else if (!can_select || status === FAILED || status === PAUSED)
      return (
        <Button
          type="primary"
          disabled={isPathMigrating}
          onClick={(e) => {
            e.stopPropagation();
            onDownloadConfirm?.(modelData);
          }}
          icon={<DownloadIcon />}
        >
          下载
        </Button>
      );
  };

  const remoteStatusToText = (item: IModelDataItem) => {
    const { can_select } = item;
    if (can_select) {
      return (
        <Button
          className={styles.updateSetting}
          variant="filled"
          icon={<ArrowClockwiseIcon fill="#344054" />}
          disabled={isPathMigrating}
          onClick={(e) => {
            e.stopPropagation();
            onModelAuthVisible?.({
              visible: true,
              type: 'update',
              modelData: modelData,
            });
          }}
        >
          更新授权
        </Button>
      );
    }
    return (
      <>
        <Button
          type="primary"
          disabled={isPathMigrating}
          onClick={(e) => {
            {
              e.stopPropagation();
              onModelAuthVisible?.({
                visible: true,
                type: 'config',
                modelData: modelData,
              });
            }
          }}
          icon={<SettingIcon fill="#ffffff" />}
        >
          配置授权
        </Button>
      </>
    );
  };

  return (
    <div
      className={`${styles.generalCard} ${!isDetail ? styles.generalCardHover : styles.generalCardDetail} `}
      onClick={() => onCardClick?.(true, modelData)}
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
            {modelSourceVal === 'local' ? (
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

      <TagsRender
        tags={(modelData?.class || []).concat([modelData?.size])}
        highlightNums={(modelData?.class || []).length}
      />

      <div className={`${isDetail ? styles.contentWrapperDetail : styles.contentWrapper}`}>
        {isDetail ? <ReactMarkdown>{modelData?.desc}</ReactMarkdown> : <Tooltip title={<div style={{ maxHeight: '100px', overflow: 'auto' }}>{modelData?.desc}</div>}>{modelData?.desc}</Tooltip>}
      </div>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{modelData.api_flavor}</div>

        {modelData?.can_select && modelSourceVal === 'local' && <div className={styles.modelStatus}>已下载</div>}
        {modelData?.can_select && modelSourceVal === 'remote' && <div className={styles.modelStatus}>已授权</div>}
      </div>
      {!isDetail && modelSourceVal === 'local' && <div className={styles.handlebar}>{statusToText(modelData)}</div>}
      {!isDetail && modelSourceVal === 'remote' && <div className={styles.handlebar}>{remoteStatusToText(modelData)}</div>}
    </div>
  );
}
