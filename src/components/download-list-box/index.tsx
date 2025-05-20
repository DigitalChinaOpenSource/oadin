import React, { memo } from 'react';
import styles from './index.module.scss';
import { SiderDownloadIcon, CloseIcon } from '../icons';
import DownloadItem from './download-item';
import useModelDownloadStore from '../../store/useModelDownloadStore';

export interface IDownloadListBoxProps {
  // 控制列表弹窗的展示位置
  className?: string;
  handleDownload: () => void;
}

export default function DownloadListBox(props: IDownloadListBoxProps) {
  const { className = '', handleDownload } = props;
  const { downloadList } = useModelDownloadStore();
  console.log('DownloadListBox======>', downloadList);
  return (
    <div className={`${className} ${styles.downloadListBox}`}>
      <div className={styles.header}>
        <div className={styles.title}>
          <div className={styles.downloadIcon}>
            <SiderDownloadIcon
              width={14}
              height={14}
            />
          </div>
          <div>最近下载</div>
        </div>
        <div
          className={styles.closeBtn}
          onClick={handleDownload}
        >
          <CloseIcon />
        </div>
      </div>
      <div className={styles.downloadList}>
        {downloadList.map((item) => (
          <DownloadItem
            key={`${item.modelType}-${item.id}`}
            downloadItem={item}
          />
        ))}
      </div>
    </div>
  );
}
