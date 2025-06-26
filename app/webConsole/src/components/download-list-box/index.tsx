import React, { memo } from 'react';
import { XIcon, DownloadIcon } from '@phosphor-icons/react';
import DownloadItem from './download-item';
import useModelDownloadStore from '../../store/useModelDownloadStore';
import styles from './index.module.scss';

export interface IDownloadListBoxProps {
  // 控制列表弹窗的展示位置
  className?: string;
  handleDownload: () => void;
}

const DownloadListBox = memo((props: IDownloadListBoxProps) => {
  const { className = '', handleDownload } = props;
  const downloadList = useModelDownloadStore.getState().downloadList;
  return (
    <div className={`${className} ${styles.downloadListBox}`}>
      <div className={styles.header}>
        <div className={styles.title}>
          <div className={styles.downloadIcon}>
            <DownloadIcon
              width={14}
              height={14}
              fill="#ffffff"
            />
          </div>
          <div>最近下载</div>
        </div>
        <div
          className={styles.closeBtn}
          onClick={handleDownload}
        >
          <XIcon
            width={16}
            height={16}
            fill="#9daabb"
          />
        </div>
      </div>
      <div className={styles.downloadList}>
        {downloadList.map((item) => (
          <DownloadItem
            key={`${item.id}`}
            downloadItem={item}
          />
        ))}
      </div>
    </div>
  );
});

export default DownloadListBox;
