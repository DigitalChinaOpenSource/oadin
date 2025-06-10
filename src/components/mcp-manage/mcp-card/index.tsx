import { useMemo, useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Popover, Tooltip } from 'antd';
import { IMcpListItem } from '../mcp-square-tab/types';
import { GlobeIcon, HardDrivesIcon } from '@phosphor-icons/react';
import TagsRender from '@/components/tags-render';
import dayjs from 'dayjs';
import defaultPng from '@/assets/favicon.png';
import { ExclamationCircleOutlined, PlusCircleOutlined } from '@ant-design/icons';
import { DotsThreeCircleIcon } from '@phosphor-icons/react';
import McpAuthModal from '@/components/mcp-manage/mcp-detail/mcp-auth-modal';
import { useMcpDetail } from '@/components/mcp-manage/mcp-detail/useMcpDetail.ts';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';

export interface IMcpCardProps {
  // 模型数据
  mcpData: IMcpListItem;
  handelMcpCardClick: (mcpId: string | number) => void;
}

export default function McpCard(props: IMcpCardProps) {
  const { mcpData, handelMcpCardClick } = props;
  const { fetchMcpDetail, mcpDetail, handleAddMcp, handleCancelMcp, cancelMcpLoading, downMcpLoading, authMcpLoading, handleAuthMcp, showMcpModal, setShowMcpModal } = useMcpDetail(mcpData.id);

  const formatUnixTime = (unixTime: number) => {
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };

  const contentRef = useRef<HTMLDivElement>(null);
  const [showTooltip, setShowTooltip] = useState(false);
  const [showOperate, setShowOperate] = useState(false);

  const clickBefore = async () => {
    if (!mcpDetail) {
      return fetchMcpDetail();
    } else {
      return mcpDetail;
    }
  };

  useEffect(() => {
    if (contentRef.current) {
      const isOverflowing = contentRef.current.scrollHeight > contentRef.current.offsetHeight;
      setShowTooltip(isOverflowing);
    }
  }, [mcpData?.abstract?.zh]);

  return (
    <div className={styles.mcpCard}>
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              src={mcpData?.logo || defaultPng}
              width={24}
            />
          </div>
          {/* 名称 */}
          <div className={styles.title}>{mcpData?.name?.zh}</div>
          {/* 本地还是云端 */}
          <div className={styles.localOrCloud}>
            {mcpData?.hosted ? (
              <>
                <HardDrivesIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <GlobeIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>云端</div>
              </>
            )}
          </div>
        </div>
      </div>
      <div style={{ height: '24px' }}>
        <TagsRender tags={mcpData?.tags || []} />
      </div>

      <Tooltip
        title={showTooltip && mcpData?.abstract?.zh}
        styles={{ root: { maxWidth: '400px' } }}
      >
        <div
          ref={contentRef}
          className={styles.contentWrapper}
        >
          {mcpData?.abstract?.zh}
        </div>
      </Tooltip>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{mcpData?.supplier}</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>{mcpData?.updatedAt && formatUnixTime(mcpData?.updatedAt) + '更新'}</div>
        {/*{!!mcpData.status && <div className={styles.mcpStatus}>已添加</div>}*/}
      </div>
      <div className={styles.cardOperate}>
        <Button
          type={'text'}
          icon={<ExclamationCircleOutlined />}
          onClick={() => handelMcpCardClick(mcpData?.id)}
        >
          查看详情
        </Button>
        <div className={styles.splitLine}></div>
        {(mcpDetail ? mcpDetail?.status === 0 : mcpData?.status === 0) && (
          <Button
            type={'link'}
            icon={<PlusCircleOutlined />}
            loading={authMcpLoading || downMcpLoading || cancelMcpLoading}
            onClick={async () => {
              const data = await clickBefore();
              if (data) await handleAddMcp(data);
            }}
          >
            立即添加
          </Button>
        )}
        {(mcpDetail ? mcpDetail?.status === 1 : mcpData?.status === 1) && (
          <Popover
            trigger="click"
            arrow={false}
            open={showOperate}
            onOpenChange={(visible) => setShowOperate(visible)}
            content={
              <div className={styles.moreOperate}>
                <div
                  className={styles.operateBtn}
                  onClick={async () => {
                    setShowOperate(false);
                    await handleAddMcp(mcpDetail as McpDetailType);
                  }}
                >
                  更新
                </div>
                <div
                  className={styles.operateBtn}
                  onClick={() => {
                    setShowOperate(false);
                    handleCancelMcp();
                  }}
                >
                  取消添加
                </div>
              </div>
            }
          >
            <Button
              type={'link'}
              icon={<DotsThreeCircleIcon />}
              onClick={async () => {
                setShowOperate(!showOperate);
                await clickBefore();
              }}
              disabled={mcpData?.envRequired == 0}
              loading={authMcpLoading || downMcpLoading || cancelMcpLoading}
            >
              更多操作
            </Button>
          </Popover>
        )}
      </div>
      {mcpDetail && (
        <McpAuthModal
          mcpDetail={mcpDetail}
          handleAuthMcp={handleAuthMcp}
          setShowMcpModal={setShowMcpModal}
          showMcpModal={showMcpModal}
          operateType={mcpDetail?.status === 0 ? 'add' : 'edit'}
        />
      )}
    </div>
  );
}
