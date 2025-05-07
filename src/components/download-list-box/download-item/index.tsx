import { Image, Progress, Tooltip } from 'antd';
import styles from './index.module.scss';
import { PauseIcon, PlayPauseIcon, CloseIcon, LoadingIcon } from '../../icons';
import ModelPng from '../../../assets/model.png';

export interface IDownloadItemProps {
  downloadItem: any;
}

export default function DownloadItem(props: IDownloadItemProps) {
  return (
    <div className={styles.downloadItem}>
      <div className={styles.titleControlBar}>
        <div>
          <Image
            className={styles.modelImg}
            src={ModelPng}
          ></Image>
          <span className={styles.title}>DeepSeek-R1-7B</span>
        </div>
        <div className={styles.controlBar}>
          <span>
            <PlayPauseIcon />
          </span>
          {/* <p><PauseIcon /></p> */}
          <span className={styles.cancel}>
            <CloseIcon />
          </span>
        </div>
      </div>

      <div className={styles.speedSizeStatus}>
        <div className={styles.speedSize}>
          <span>0</span>
          B/s-
          <span>0</span>
          MB-
          <span>4399</span>
          MB
        </div>
        <div className={styles.status}>
          <LoadingIcon />
          <span className={styles.isloading}>正在下载</span>
          {/* <span>已暂停下载</span> */}
        </div>
      </div>

      <div className={styles.progress}>
        <Progress
          percent={50}
          showInfo={false}
          strokeColor="#5429ff"
          size="small"
        />
      </div>
    </div>
  );
}
