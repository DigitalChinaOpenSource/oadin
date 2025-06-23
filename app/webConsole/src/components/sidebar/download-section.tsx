import React from 'react';
import { Badge, Popover, Tooltip } from 'antd';
import { DownloadIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import DownloadListBox from '../download-list-box';
import useModelDownloadStore from '../../store/useModelDownloadStore';

const DownloadSection = ({ collapsed }: { collapsed: boolean }) => {
  const [isDownloadListOpen, setIsDownloadListOpen] = React.useState(false);
  const downloadList = useModelDownloadStore((state) => state.downloadList);

  const handleDownload = (visible: boolean) => {
    setIsDownloadListOpen(visible);
  };

  if (downloadList.length === 0) {
    return null;
  }

  return (
    <Popover
      style={{ padding: 0 }}
      zIndex={1300}
      content={
        <DownloadListBox
          className={styles.downloadListWrapper}
          handleDownload={() => handleDownload(false)}
        />
      }
      trigger={'click'}
      placement={'rightTop'}
      arrow={false}
      open={isDownloadListOpen}
      onOpenChange={(visible) => {
        handleDownload(visible);
      }}
    >
      <div className={styles.mcpDownloadBox}>
        <div className={styles.mpcDownloadBoxContent}>
          <Tooltip title={collapsed ? '正在下载' : ''}>
            <div className={styles.downloadBtnBox}>
              <Badge
                count={downloadList?.length}
                showZero={false}
                className={styles.badge}
              >
                <div className={styles.downloadBtn}>
                  <DownloadIcon
                    width={14}
                    height={14}
                    fill="#ffffff"
                  />
                </div>
              </Badge>
            </div>
          </Tooltip>
          {!collapsed && <div>正在下载</div>}
        </div>
      </div>
    </Popover>
  );
};

export default React.memo(DownloadSection);
