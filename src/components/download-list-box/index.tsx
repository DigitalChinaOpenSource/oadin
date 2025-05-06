import React from 'react';
import styles from './index.module.scss';
import { SiderDownloadIcon, CloseIcon } from '../icons';
import DownloadItem from './download-item';
export interface IDownloadListBoxProps {
  // 控制列表弹窗的展示位置
  className?: string;
  handleDownload: () => void;
}

export default function DownloadListBox(props: IDownloadListBoxProps) {
  const { className = '', handleDownload } = props;

  // 模拟5条下载数据
  const downloadItems = [
    { id: 1, name: '模型名称1', progress: 25, status: 'downloading' },
    { id: 2, name: '模型名称2', progress: 50, status: 'downloading' },
    { id: 3, name: '模型名称3', progress: 100, status: 'completed' },
    { id: 4, name: '模型名称4', progress: 0, status: 'error' },
    { id: 5, name: '模型名称5', progress: 75, status: 'downloading' },
  ];

  return (
    <div className={`${className} ${styles.downloadListBox}`}>
      <div className={styles.header}>
        <div className={styles.title}>
          <div className={styles.downloadIcon}>
            <SiderDownloadIcon width={14} height={14} />
          </div>
          <div>最近下载</div>
        </div>
        <div className={styles.closeBtn} onClick={handleDownload}>
          <CloseIcon />
        </div>
      </div>
      <div className={styles.downloadList}>
        {downloadItems.map((item) => (
          <DownloadItem downloadItem={item} />
        ))}
      </div>
    </div>
  );
}
