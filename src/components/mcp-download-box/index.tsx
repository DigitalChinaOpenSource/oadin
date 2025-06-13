import styles from './index.module.scss';
import mcpAddSvg from '@/components/icons/mcpAdd.svg';
import { CloseOutlined } from '@ant-design/icons';
import useMcpDownloadStore from '@/store/useMcpDownloadStore.ts';
import favicon from '@/assets/favicon.png';
import { ArrowClockwiseIcon } from '@phosphor-icons/react';
import { useMcpDetail } from '@/components/mcp-manage/mcp-detail/useMcpDetail.ts';
import McpAuthModal from '@/components/mcp-manage/mcp-detail/mcp-auth-modal';
import React, { useRef, useState } from 'react';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import loadingSvg from '@/components/icons/real-loading.svg';
import failedSvg from '@/components/icons/failed.svg';
import { Tooltip } from 'antd';

export default function McpDownloadBox() {
  const { mcpDownloadList, setMcpAddModalShow, delMcpDownloadItem } = useMcpDownloadStore();
  const { handleAuthMcp, setShowMcpModal, showMcpModal, handleAddMcp } = useMcpDetail();
  // const [curMcpDetail, setCurMcpDetail] = useState<McpDetailType | null>(null);
  const currentMcpDetail = useRef<McpDetailType | null>(null);

  return (
    <div className={styles.mcpDownloadBox}>
      <div className={styles.header}>
        <div className={styles.title}>
          <img
            src={mcpAddSvg}
            alt="最近添加"
          />
          <div>最近添加</div>
        </div>
        <CloseOutlined
          className={styles.closeIcon}
          onClick={() => setMcpAddModalShow(false)}
        />
      </div>
      <div className={styles.content}>
        {mcpDownloadList.map((item) => (
          <div
            key={item.mcpDetail?.id}
            className={styles.contentItem}
          >
            <div className={styles.header}>
              <div className={styles.title}>
                <img
                  src={item.mcpDetail?.logo || favicon}
                  alt={item.mcpDetail?.name?.zh || ''}
                  className={styles.mcpIcon}
                />
                <div>{item.mcpDetail?.name?.zh}</div>
              </div>
              <div className={styles.operate}>
                {item.downStatus === 'error' && (
                  <>
                    <Tooltip title={'重新添加'}>
                      <ArrowClockwiseIcon
                        className={styles.closeIcon}
                        onClick={() => {
                          console.log('重新添加的mcpDetail===>', item.mcpDetail);
                          currentMcpDetail.current = item.mcpDetail;
                          const { envRequired } = item.mcpDetail!;

                          if (envRequired !== 0) {
                            setShowMcpModal(true);
                          } else {
                            handleAddMcp(item.mcpDetail as McpDetailType);
                          }
                        }}
                      />
                    </Tooltip>
                    <Tooltip title={'删除'}>
                      <CloseOutlined
                        className={styles.closeIcon}
                        onClick={() => delMcpDownloadItem(item.mcpDetail?.id as string)}
                      />
                    </Tooltip>
                  </>
                )}
              </div>
            </div>
            <div className={styles.contentStatus}>
              {item.downStatus === 'downloading' && (
                <>
                  <img
                    src={loadingSvg}
                    alt=""
                    style={{ width: '16px', height: '16px' }}
                  />
                  <div>正在添加</div>
                </>
              )}
              {item.downStatus === 'error' && (
                <>
                  <img
                    src={failedSvg}
                    alt=""
                    style={{ width: '16px', height: '16px' }}
                  />
                  <div>添加失败</div>
                </>
              )}
            </div>
          </div>
        ))}
      </div>
      {currentMcpDetail.current && (
        <McpAuthModal
          mcpDetail={currentMcpDetail.current as McpDetailType}
          handleAuthMcp={handleAuthMcp}
          setShowMcpModal={setShowMcpModal}
          showMcpModal={showMcpModal}
          operateType={currentMcpDetail.current?.status === 0 ? 'add' : 'edit'}
        />
      )}
    </div>
  );
}
