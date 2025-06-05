import ReactMarkdown from 'react-markdown';
import styles from './index.module.scss';
import { Button, Tooltip, message, Radio } from 'antd';
import { IModelAuth } from '../../types';
import { IModelDataItem, IModelSourceType } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon, SettingIcon, ArrowClockwiseIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import React, { MouseEvent } from 'react';
import recommendedIcon from '@/components/icons/recommendIcon.png';

export interface IGeneralCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 是否可选
  isSelectable?: boolean;
  // 选择文案提示
  selectTooltip?: string;
  // 模型数据
  modelData: IModelDataItem;
  modelSourceVal: IModelSourceType;
  onCardClick?: (visible: boolean, selectModelData: IModelDataItem) => void;
  onModelAuthVisible?: (data: IModelAuth) => void;
  onDeleteConfirm?: (modelData: IModelDataItem) => void;
  onDownloadConfirm?: (modelData: IModelDataItem) => void;
  selectModel?: IModelDataItem;
  setSelectModel?: (modelData: IModelDataItem) => void;
}

export default function GeneralCard(props: IGeneralCardProps) {
  const { isDetail, onCardClick, modelSourceVal, onDeleteConfirm, onModelAuthVisible, onDownloadConfirm, modelData, setSelectModel, selectModel } = props;
  const { isPathMigrating } = useModelPathChangeStore();
  const toolTipsText = props?.selectTooltip ?? '请先下载/授权，再体验';
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
          onClick={(e) => {
            e.stopPropagation();
            if (isPathMigrating) {
              message.warning('模型存储路径正在变更中，请稍后操作');
              return;
            }
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
          onClick={(e) => {
            e.stopPropagation();
            if (isPathMigrating) {
              message.warning('模型存储路径正在变更中，请稍后操作');
              return;
            }
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
          onClick={(e) => {
            e.stopPropagation();
            if (isPathMigrating) {
              message.warning('模型存储路径正在变更中，请稍后操作');
              return;
            }
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
          onClick={(e) => {
            {
              e.stopPropagation();
              if (isPathMigrating) {
                message.warning('模型存储路径正在变更中，请稍后操作');
                return;
              }
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
  const handleSelectModelData = (e: MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    if (setSelectModel) {
      if (selectModel?.id && selectModel?.id === modelData?.id) {
        setSelectModel({} as IModelDataItem);
      } else {
        setSelectModel(modelData);
      }
    }
  };
  return (
    <div
      className={`${styles.generalCard} ${!isDetail ? styles.generalCardHover : styles.generalCardDetail} `}
      onClick={() => onCardClick?.(true, modelData)}
    >
      {/* 推荐使用，定位右上角 */}
      {modelData?.is_recommended && !props.isSelectable && <div className={styles.recommend}>推荐使用</div>}
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              alt=""
              src={modelData?.avatar}
              width={24}
            />
            {modelData?.is_recommended && props.isSelectable ? (
              <span>
                <img
                  src={recommendedIcon}
                  alt=""
                />
              </span>
            ) : null}
          </div>
          {/* 名称 */}
          <Tooltip title={modelData?.name}>
            <div className={styles.title}>{modelData?.name}</div>
          </Tooltip>
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
      {props.isSelectable ? (
        <div className={styles.actionsWarp}>
          <Tooltip title={modelData?.can_select ? null : toolTipsText}>
            <Radio
              disabled={!modelData?.can_select}
              value={modelData?.id}
              key={modelData?.id}
              onClick={(e: MouseEvent<HTMLElement>) => handleSelectModelData(e)}
            />
          </Tooltip>
        </div>
      ) : null}

      <TagsRender
        tags={(modelData?.class || []).concat([modelData?.size])}
        highlightNums={(modelData?.class || []).length}
      />

      <div className={`${isDetail ? styles.contentWrapperDetail : styles.contentWrapper}`}>
        {isDetail ? <ReactMarkdown>{modelData?.desc}</ReactMarkdown> : <Tooltip title={<div style={{ maxHeight: '100px', overflow: 'auto' }}>{modelData?.desc}</div>}>{modelData?.desc}</Tooltip>}
      </div>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{modelData.flavor}</div>

        {modelData?.can_select && modelSourceVal === 'local' && <div className={styles.modelStatus}>已下载</div>}
        {modelData?.can_select && modelSourceVal === 'remote' && <div className={styles.modelStatus}>已授权</div>}
      </div>
      {!isDetail && modelSourceVal === 'local' && <div className={styles.handlebar}>{statusToText(modelData)}</div>}
      {!isDetail && modelSourceVal === 'remote' && <div className={styles.handlebar}>{remoteStatusToText(modelData)}</div>}
    </div>
  );
}
