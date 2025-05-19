import { useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Tag, Progress } from 'antd';
import { IModelAuth } from '../model-manage-tab/types';
import { ModelDataItem } from '@/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon } from '@/components/icons';

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

  const [isOverflow, setIsOverflow] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const checkOverflow = () => {
      const element = textRef.current;
      if (element) {
        // 检查内容是否超过两行
        const lineHeight = parseInt(window.getComputedStyle(element).lineHeight);
        const maxHeight = lineHeight * 2;
        setIsOverflow(element.scrollHeight > maxHeight);
      }
    };
    checkOverflow();
    window.addEventListener('resize', checkOverflow);
    return () => window.removeEventListener('resize', checkOverflow);
  }, [modelData]);

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
          type="text"
          className={styles.downloaded}
          onClick={() => onDeleteConfirm?.(modelData)}
        >
          已下载
          <DeleteIcon fill="#5429ff" />
        </Button>
      );
    if (status === PAUSED) return '暂停';
    else if (!can_select || status === FAILED)
      return (
        <Button
          type="primary"
          size="small"
          onClick={() => onDownloadConfirm?.(modelData)}
        >
          下载
          <DownloadIcon />
        </Button>
      );
  };

  return (
    <div className={`${isDetail ? '' : styles.generalCardStyle} ${styles.generalCard}`}>
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          <img
            src={modelData?.avatar}
            width={24}
          />
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

          {/* 推荐使用 */}
          {modelData?.is_recommended && <div className={styles.recommend}>🔥 推荐使用</div>}

          {/* 已授权 */}
          {/* <div className={styles.authorized}>已授权</div> */}
        </div>

        {!isDetail && (
          <div className={styles.cardDownload}>
            {modelData?.source === 'local' ? (
              <>
                <>{statusToText(modelData)}</>
              </>
            ) : (
              <>
                {/* 云端下载相关内容 */}
                <Button
                  type="primary"
                  size="small"
                  onClick={() =>
                    onModelAuthVisible?.({
                      visible: true,
                      type: 'config',
                      modelData: modelData,
                    })
                  }
                >
                  配置授权
                </Button>
                {/* <Button className={styles.updateSetting} variant="filled" size="small" onClick={() => onModelAuthorizeVisible?.('update')}>
                    更新配置
                  </Button> */}
              </>
            )}
          </div>
        )}
      </div>

      <div className={styles.contentWrapper}>
        <div
          ref={textRef}
          className={`${isDetail ? '' : styles.textContent}`}
        >
          {modelData?.desc}
        </div>
        {isOverflow && !isDetail && (
          <Button
            type="link"
            size="small"
            className={styles.moreButton}
            onClick={() => onDetailModalVisible?.(true, modelData)}
          >
            更多
          </Button>
        )}
      </div>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>深度求索</div>
        <div className={styles.contextLength}>
          <span className={styles.splitLine}>｜</span>
          <span>上下文长度：</span>
          <span>测试暂无</span>
        </div>
        <div className={styles.modelSize}>
          <span className={styles.splitLine}>｜</span>
          <span>大小：</span>
          <span>{modelData?.size}</span>
        </div>
        <div className={styles.beUsed}>
          <span className={styles.splitLine2}>｜</span>
          <span>使用中：</span>
          <span>测试2个</span>
        </div>
      </div>

      <div className={styles.tagWrapper}>
        {(modelData?.class || []).map((tag: string, index: number) => (
          <Tag key={index}>{tag}</Tag>
        ))}
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
    </div>
  );
}
