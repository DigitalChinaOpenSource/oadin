import { Image, message, Progress } from 'antd';
import styles from './index.module.scss';
import greySpinner from '@/components/icons/greySpinner.svg';
import { ArrowClockwiseIcon, PlayPauseIcon, PauseIcon, XIcon, XCircleIcon } from '@phosphor-icons/react';
import { useModelPathChangeStore } from '@/store/useModelPathChangeStore';
import { DOWNLOAD_STATUS } from '@/constants';
import { useViewModel } from './view-model';
import { IModelDataItem } from '@/types';

export interface IDownloadItemProps {
  downloadItem: IModelDataItem;
}

export default function DownloadItem(props: IDownloadItemProps) {
  const { downloadItem } = props;
  const { migratingStatus } = useModelPathChangeStore();
  const { fetchCancelModel, fetchRemoveModel, fetchDownloadStart } = useViewModel();
  const { COMPLETED, FAILED, IN_PROGRESS, PAUSED } = DOWNLOAD_STATUS;
  return (
    <div className={styles.downloadItem}>
      <div className={styles.titleControlBar}>
        <div>
          <Image
            preview={false}
            className={styles.modelImg}
            src={downloadItem.avatar}
          />
          <span className={styles.title}>{downloadItem.name}</span>
        </div>
        <div className={styles.controlBar}>
          {downloadItem.status === IN_PROGRESS && (
            <div
              onClick={() => {
                if (migratingStatus === 'pending') {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchCancelModel(downloadItem);
              }}
            >
              <PauseIcon
                width={16}
                height={16}
                fill="#9daabb"
              />
            </div>
          )}
          {downloadItem.status === PAUSED && (
            <div
              onClick={() => {
                if (migratingStatus === 'pending') {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchDownloadStart(downloadItem);
              }}
            >
              <PlayPauseIcon
                width={16}
                height={16}
                fill="#9daabb"
              />
            </div>
          )}
          {downloadItem.status === FAILED && (
            <div
              onClick={() => {
                if (migratingStatus === 'pending') {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchDownloadStart(downloadItem);
              }}
            >
              <ArrowClockwiseIcon
                width={16}
                height={16}
                fill="#9daabb"
              />
            </div>
          )}
          <div
            className={styles.cancel}
            onClick={() => {
              if (migratingStatus === 'pending') {
                message.warning('模型存储路径正在变更中，请稍后操作');
                return;
              }
              fetchRemoveModel(downloadItem);
            }}
          >
            <XIcon
              width={16}
              height={16}
              fill="#9daabb"
            />
          </div>
        </div>
      </div>

      <div className={styles.speedSizeStatus}>
        <div className={styles.speedSize}>
          {Boolean(downloadItem.completedsize && downloadItem.totalsize) && (
            <>
              <span>{downloadItem.completedsize} MB/</span>
              <span>{downloadItem.totalsize} MB</span>
            </>
          )}
        </div>
        <div className={styles.status}>
          {downloadItem.status === FAILED && (
            <>
              <XCircleIcon
                width={16}
                height={16}
                fill="#E85951"
              />
              <span className={styles.statusError}>下载失败</span>
            </>
          )}
          {downloadItem.status === COMPLETED && <span>下载完成</span>}
          {downloadItem.status === IN_PROGRESS && (
            <>
              <img
                src={greySpinner}
                style={{ width: '18px' }}
              />
              <span className={styles.statusLoading}>正在下载</span>
            </>
          )}
          {downloadItem.status === PAUSED && <span className={styles.statusPaused}>已暂停下载</span>}
        </div>
      </div>

      <div className={styles.progress}>
        <Progress
          percent={downloadItem?.currentDownload}
          showInfo={false}
          strokeColor="#4f4dff"
          size="small"
        />
      </div>
    </div>
  );
}
