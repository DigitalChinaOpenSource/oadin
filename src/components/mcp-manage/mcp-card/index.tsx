import { useMemo } from 'react';
import styles from './index.module.scss';
import { Button, Progress, Tooltip } from 'antd';
import { IMcpListItem } from '../mcp-list-tab/types';
import { DOWNLOAD_STATUS } from '@/constants';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';

export interface IMcpCardProps {
  // 模型数据
  mcpData: IMcpListItem;
}

export default function McpCard(props: IMcpCardProps) {
  const { mcpData } = props;

  const tags = useMemo(
    () => ['深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成'],
    [mcpData?.tags],
  );

  return (
    <div
      className={styles.mcpCard}
      // onClick={() => onCardClick?.(true, modelData)}
    >
      {/* 推荐使用，定位右上角 */}
      {modelData?.is_recommended && <div className={styles.recommend}>推荐使用</div>}
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              src={modelData?.avatar}
              width={24}
            />
          </div>
          {/* 名称 */}
          <div className={styles.title}>{modelData?.name}</div>
          {/* 本地还是云端 */}
          <div className={styles.localOrCloud}>
            {modelData?.source === 'local' ? (
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

      <TagsRender tags={tags} />

      <Tooltip title={modelData?.desc}>
        <div className={styles.contentWrapper}>{modelData?.desc}</div>
      </Tooltip>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>深度求索</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>2025-05-19 更新</div>
        {modelData?.can_select && <div className={styles.modelStatus}>已下载</div>}
      </div>
      {Boolean(modelData?.currentDownload) && (
        <Progress
          className={styles.progress}
          percent={30}
          size="small"
          showInfo={false}
          strokeColor="#5429ff"
          strokeLinecap="butt"
        />
      )}
    </div>
  );
}
