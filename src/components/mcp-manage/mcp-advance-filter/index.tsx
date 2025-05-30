// mcp右侧过滤器
import { Tooltip } from 'antd';
import styles from './index.module.scss';
import foldSvg from '@/components/icons/fold.svg';
import cleanSvg from '@/components/icons/clean.svg';
import McpTagsCheck from '@/components/mcp-manage/mcp-advance-filter/mcp-tags-check';
import { ITagsDataItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';

interface IMcpAdvanceFilter {
  // 是否折叠
  collapsed?: boolean;
  // 折叠通知外部
  setCollapsed?: (isCollapsed: boolean) => void;
  handleClearTags?: () => void;
  tagsData?: ITagsDataItem[];
  checkedValues?: Record<string, any>;
  handleTagsChange?: (category: string, list: any[]) => void;
}

export default function McpAdvanceFilter(props: IMcpAdvanceFilter) {
  const { collapsed, setCollapsed, handleClearTags, tagsData, checkedValues, handleTagsChange } = props;

  return (
    <div
      className={styles.mcpAdvanceFilter}
      style={{
        visibility: collapsed ? 'hidden' : 'visible',
        width: collapsed ? '0' : '224px',
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
        <div
          className={styles.clean}
          onClick={handleClearTags}
        >
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
        <McpTagsCheck
          tagsData={tagsData}
          checkedValues={checkedValues}
          handleTagsChange={handleTagsChange}
        />
      </div>
    </div>
  );
}
