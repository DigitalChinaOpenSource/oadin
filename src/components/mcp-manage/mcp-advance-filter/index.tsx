// mcp右侧过滤器
import { useState } from 'react';
import { Tooltip } from 'antd';
import styles from './index.module.scss';
import foldSvg from '@/components/icons/fold.svg';
import cleanSvg from '@/components/icons/clean.svg';

interface IMcpAdvanceFilter {
  // 是否折叠
  collapsed?: boolean;
  // 折叠通知外部
  setCollapsed?: (isCollapsed: boolean) => void;
}

export default function McpAdvanceFilter(props: IMcpAdvanceFilter) {
  const { collapsed, setCollapsed } = props;

  return (
    <div
      className={styles.mcpAdvanceFilter}
      style={{
        visibility: collapsed ? 'hidden' : 'visible',
        width: collapsed ? '0' : '200px',
        overflow: 'hidden',
        transition: 'width 0.3s ease-in-out',
        padding: collapsed ? '0' : '28px 24px 24px 0',
      }}
    >
      <div className={styles.filterTitle}>
        <div className={styles.titleOperate}>
          <Tooltip title="收起筛选">
            <div
              className={styles.foldIcon}
              onClick={() => setCollapsed?.(true)}
            >
              <img
                src={foldSvg}
                alt="折叠过滤面板"
              />
            </div>
          </Tooltip>
          MCP 服务筛选
        </div>
        <div className={styles.clean}>
          <Tooltip title="清除筛选条件">
            <img
              src={cleanSvg}
              alt="清除"
            />
          </Tooltip>
        </div>
      </div>

      <div className={styles.filterContent}>
        {/* 筛选表单等内容放这里 */}
        <p>这里是筛选项</p>
      </div>
    </div>
  );
}
