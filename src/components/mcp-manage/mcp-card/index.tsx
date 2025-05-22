import { useMemo } from 'react';
import styles from './index.module.scss';
import { Button, Tooltip } from 'antd';
import { IMcpListItem } from '../mcp-square-tab/types';
import { LoadingIcon, LocalIcon, CloudIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';
import ModelPng from '@/assets/model.png';
import dayjs from 'dayjs';

export interface IMcpCardProps {
  // 模型数据
  mcpData: IMcpListItem;
  handelMcpCardClick: (mcpId: number) => void;
}

export default function McpCard(props: IMcpCardProps) {
  const { mcpData, handelMcpCardClick } = props;

  const tags = useMemo(
    () => ['深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成', '999MB', '7B', '128K', '深度思考', '文本生成'],
    [mcpData?.tags],
  );

  const formateUnixTime = (unixTime: number) => {
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };
  return (
    <div
      className={styles.mcpCard}
      onClick={() => handelMcpCardClick(mcpData.id)}
    >
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              src={ModelPng}
              width={24}
            />
          </div>
          {/* 名称 */}
          <div className={styles.title}>{mcpData.name.zh}</div>
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

      <TagsRender tags={mcpData?.tags} />

      <Tooltip title={'asdad'}>
        <div className={styles.contentWrapper}>
          这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述这里是描述
        </div>
      </Tooltip>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>深度求索</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>{formateUnixTime(mcpData.updatedAt)} 更新</div>
        {!!mcpData.status && <div className={styles.mcpStatus}>已授权</div>}
      </div>
    </div>
  );
}
