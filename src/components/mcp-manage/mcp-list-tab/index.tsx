import { Tooltip } from 'antd';
import { Input } from 'antd';
import { useViewModel } from './view-model';
import expandSvg from '@/components/icons/expand.svg';
import { SearchIcon } from '@/components/icons';
import McpAdvanceFilter from '../mcp-advance-filter';
import McpList from '../mcp-list';
import { mcpListDataMock } from './constant';
import styles from './index.module.scss';

export default function McpListTab() {
  const vm = useViewModel();

  return (
    <div className={styles.mcpListTab}>
      {/* 列表区域 */}
      <div className={styles.mcpListContent}>
        <div className={styles.mcpTitle}>
          <div className={styles.mcpTitleText}>
            <span>MCP服务</span>
            <span className={styles.mcpCount}>共 {vm.mcpListData.length} 个</span>
          </div>
          <div className={styles.searchInput}>
            <Input
              allowClear
              placeholder="请输入 MCP 服务名称"
              suffix={<SearchIcon />}
              value={vm.mcpSearchVal.keyword}
              onChange={(e) => vm.onMcpInputSearch(e.target.value.trim())}
              style={{ width: 380 }}
            />
          </div>
          {vm.collapsed && (
            <Tooltip title="展开筛选">
              <div
                className={styles.expandIcon}
                onClick={() => vm.setCollapsed(false)}
              >
                <img
                  src={expandSvg}
                  alt="折叠筛选面板"
                />
              </div>
            </Tooltip>
          )}
        </div>
        {/* vm.mcpListData ||  */}
        <McpList mcpListData={mcpListDataMock} />
      </div>
      {/* 过滤器 */}
      <McpAdvanceFilter
        collapsed={vm.collapsed}
        setCollapsed={vm.setCollapsed}
      />
    </div>
  );
}
