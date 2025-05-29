import { useMemo, useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Tooltip } from 'antd';
import { IMcpListItem } from '../mcp-square-tab/types';
import { LocalIcon, CloudIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';
import dayjs from 'dayjs';
import defaultPng from '@/assets/favicon.png';

export interface IMcpCardProps {
  // 模型数据
  mcpData: IMcpListItem;
  handelMcpCardClick: (mcpId: string | number) => void;
}

export default function McpCard(props: IMcpCardProps) {
  const { mcpData, handelMcpCardClick } = props;

  // const tags = useMemo(
  //   () => ['深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成'],
  //   [mcpData?.tags],
  // );

  const formatUnixTime = (unixTime: number) => {
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };

  const contentRef = useRef<HTMLDivElement>(null);
  const [showTooltip, setShowTooltip] = useState(false);

  useEffect(() => {
    if (contentRef.current) {
      const isOverflowing = contentRef.current.scrollHeight > contentRef.current.offsetHeight;
      setShowTooltip(isOverflowing);
    }
  }, [mcpData?.abstract?.zh]);

  return (
    <div
      className={styles.mcpCard}
      onClick={() => handelMcpCardClick(mcpData?.id)}
    >
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
                <LocalIcon />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <CloudIcon />
                <div className={styles.localOrCloudText}>云端</div>
              </>
            )}
          </div>
        </div>
      </div>
      <div style={{ height: '24px' }}>
        <TagsRender tags={mcpData?.tags || []} />
      </div>

      <Tooltip title={showTooltip && mcpData?.abstract?.zh}>
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
        {!!mcpData.status && <div className={styles.mcpStatus}>已添加</div>}
      </div>
    </div>
  );
}
