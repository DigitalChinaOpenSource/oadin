import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import { Button, Tooltip, message, Radio } from 'antd';
import { IModelAuth } from '../../types';
import { IModelDataItem, IModelSourceType } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { DownloadSimpleIcon, GlobeIcon, ArrowClockwiseIcon, SpinnerIcon, GearSixIcon, TrashIcon, HardDrivesIcon } from '@phosphor-icons/react';
import TagsRender from '@/components/tags-render';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import React, { MouseEvent } from 'react';
import recommendedIcon from '@/components/icons/recommendIcon.png';
import { ISelectedDialogProps } from '@/components/choose-model-dialog';
import styles from './index.module.scss';

export interface IGeneralCardProps extends ISelectedDialogProps {
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
  mine?: boolean;
}

export default function GeneralCard(props: IGeneralCardProps) {
  const { isDetail, onCardClick, modelSourceVal, onDeleteConfirm, onModelAuthVisible, onDownloadConfirm, modelData, mine, setSelecteStatedModel, selectedStateModel } = props;
  const toolTipsText = props?.selectTooltip ?? '请先下载/授权，再体验';

  const { migratingStatus } = useModelPathChangeStore();
  const statusToText = (item: IModelDataItem) => {
    const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
    const { status, can_select } = item;
    if (status === IN_PROGRESS)
      return (
        <Button
          className={styles.downloadedBtn}
          icon={
            <SpinnerIcon
              width={16}
              height={16}
              fill="#344054"
            />
          }
        >
          下载中
        </Button>
      );
    if (can_select || status === COMPLETED || mine)
      return (
        <Button
          className={styles.downloadedBtn}
          onClick={(e) => {
            e.stopPropagation();
            if (migratingStatus === 'pending') {
              message.warning('模型存储路径正在变更中，请稍后操作');
              return;
            }
            onDeleteConfirm?.(modelData);
          }}
          icon={
            <TrashIcon
              width={16}
              height={16}
              fill="#344054"
            />
          }
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
            if (migratingStatus === 'pending') {
              message.warning('模型存储路径正在变更中，请稍后操作');
              return;
            }
            onDownloadConfirm?.(modelData);
          }}
          icon={
            <DownloadSimpleIcon
              width={16}
              height={16}
              fill="#ffffff"
            />
          }
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
          icon={
            <ArrowClockwiseIcon
              width={16}
              height={16}
              fill="#344054"
            />
          }
          onClick={(e) => {
            e.stopPropagation();
            if (migratingStatus === 'pending') {
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
              if (migratingStatus === 'pending') {
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
          icon={
            <GearSixIcon
              width={16}
              height={16}
              fill="#ffffff"
            />
          }
        >
          配置授权
        </Button>
      </>
    );
  };

  const handleSelectModelData = (e: MouseEvent<HTMLElement>) => {
    e.stopPropagation();
    const tempSelectedModel = selectedStateModel?.id && selectedStateModel?.id === modelData?.id ? ({} as IModelDataItem) : modelData;
    setSelecteStatedModel?.(tempSelectedModel);
  };

  const convertProviderName = (flavor: string): string => {
    const providerMap: Record<string, string> = {
      deepseek: '深度求索',
      aliyun: '阿里巴巴',
      Yi: '灵一万物',
      zhipuAi: '智谱清言',
    };

    return providerMap[flavor] || flavor;
  };

  return (
    <div
      className={`${styles.generalCard} ${!isDetail ? styles.generalCardHover : styles.generalCardDetail} `}
      onClick={(e) => {
        if (props.isSelectable) {
          if (modelData?.can_select) {
            handleSelectModelData(e);
          } else {
            message.warning(toolTipsText);
          }
        } else {
          onCardClick?.(true, modelData);
        }
      }}
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
                <HardDrivesIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <GlobeIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
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
        {isDetail ? (
          <>{ReactMarkdown({ children: modelData?.desc, remarkPlugins: [remarkGfm], rehypePlugins: [rehypeRaw] })}</>
        ) : (
          <Tooltip title={<div style={{ maxHeight: '100px', overflow: 'auto' }}>{modelData?.desc}</div>}>{modelData?.desc || '暂无模型简介'}</Tooltip>
        )}
      </div>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{convertProviderName(modelData.flavor)}</div>

        {modelData?.can_select && modelSourceVal === 'local' && <div className={styles.modelStatus}>已下载</div>}
        {modelData?.can_select && modelSourceVal === 'remote' && <div className={styles.modelStatus}>已授权</div>}
      </div>
      {!isDetail && modelSourceVal === 'local' && <div className={styles.handlebar}>{statusToText(modelData)}</div>}
      {!isDetail && modelSourceVal === 'remote' && <div className={styles.handlebar}>{remoteStatusToText(modelData)}</div>}
    </div>
  );
}
