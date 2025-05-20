import { Image, Progress } from 'antd';
import styles from './index.module.scss';
import { PlayPauseIcon, CloseIcon, LoadingIcon, FillCloseIcon, PauseIcon, ArrowClockwiseIcon } from '../../icons';
import ModelPng from '@/assets/model.png';
import { DOWNLOAD_STATUS } from '@/constants';
import { useViewModel } from './view-model';

export interface IDownloadItemProps {
  downloadItem: any;
}

export default function DownloadItem(props: IDownloadItemProps) {
  const { downloadItem } = props;
  const { fetchCancelModel, fetchDownloadModel, fetchRemoveModel } = useViewModel();
  const { COMPLETED, FAILED, IN_PROGRESS, PAUSED } = DOWNLOAD_STATUS;
  console.log('downloadItem===>', downloadItem);
  return (
    <div className={styles.downloadItem}>
      <div className={styles.titleControlBar}>
        <div>
          <Image
            className={styles.modelImg}
            src={ModelPng}
          ></Image>
          <span className={styles.title}>{downloadItem.name}</span>
        </div>
        <div className={styles.controlBar}>
          {downloadItem.status === IN_PROGRESS && (
            <div onClick={() => fetchCancelModel(downloadItem.name)}>
              <PauseIcon />
            </div>
          )}
          {downloadItem.status === PAUSED && (
            <div onClick={() => fetchDownloadModel(downloadItem)}>
              <PlayPauseIcon />
            </div>
          )}
          {downloadItem.status === FAILED && (
            <div onClick={() => fetchDownloadModel(downloadItem)}>
              <ArrowClockwiseIcon />
            </div>
          )}
          <div
            className={styles.cancel}
            onClick={() => fetchRemoveModel(downloadItem.name)}
          >
            <CloseIcon />
          </div>
        </div>
      </div>

      <div className={styles.speedSizeStatus}>
        <div className={styles.speedSize}>
          <span>{downloadItem.completedsize} MB/</span>
          <span>{downloadItem.totalsize} MB</span>
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
              <LoadingIcon />
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
          strokeColor="#5429ff"
          size="small"
        />
      </div>
    </div>
  );
}
