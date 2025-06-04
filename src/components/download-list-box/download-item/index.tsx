import { Image, message, Progress } from 'antd';
import styles from './index.module.scss';
import { PlayPauseIcon, CloseIcon, FillCloseIcon, PauseIcon, ArrowClockwiseIcon } from '../../icons';
import greySpinner from '@/components/icons/greySpinner.svg';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import { DOWNLOAD_STATUS } from '@/constants';
import { useViewModel } from './view-model';
import { IModelDataItem } from '@/types';

export interface IDownloadItemProps {
  downloadItem: IModelDataItem;
}

export default function DownloadItem(props: IDownloadItemProps) {
  const { downloadItem } = props;
  const { isPathMigrating } = useModelPathChangeStore();
  const { fetchCancelModel, fetchRemoveModel, fetchDownloadStart } = useViewModel();
  const { COMPLETED, FAILED, IN_PROGRESS, PAUSED } = DOWNLOAD_STATUS;
  return (
    <div className={styles.downloadItem}>
      <div className={styles.titleControlBar}>
        <div>
          <Image
            className={styles.modelImg}
            src={downloadItem.avatar}
          />
          <span className={styles.title}>{downloadItem.name}</span>
        </div>
        <div className={styles.controlBar}>
          {downloadItem.status === IN_PROGRESS && (
            <div
              onClick={() => {
                if (isPathMigrating) {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchCancelModel(downloadItem);
              }}
            >
              <PauseIcon />
            </div>
          )}
          {downloadItem.status === PAUSED && (
            <div
              onClick={() => {
                if (isPathMigrating) {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchDownloadStart(downloadItem);
              }}
            >
              <PlayPauseIcon />
            </div>
          )}
          {downloadItem.status === FAILED && (
            <div
              onClick={() => {
                if (isPathMigrating) {
                  message.warning('模型存储路径正在变更中，请稍后操作');
                  return;
                }
                fetchDownloadStart(downloadItem);
              }}
            >
              <ArrowClockwiseIcon />
            </div>
          )}
          <div
            className={styles.cancel}
            onClick={() => {
              if (isPathMigrating) {
                message.warning('模型存储路径正在变更中，请稍后操作');
                return;
              }
              fetchRemoveModel(downloadItem);
            }}
          >
            <CloseIcon />
          </div>
        </div>
      </div>

      <div className={styles.speedSizeStatus}>
        <div className={styles.speedSize}>
          {downloadItem.completedsize && downloadItem.totalsize && (
            <>
              <span>{downloadItem.completedsize} MB/</span>
              <span>{downloadItem.totalsize} MB</span>
            </>
          )}
        </div>
        <div className={styles.status}>
          {downloadItem.status === FAILED && (
            <>
              <FillCloseIcon />
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
