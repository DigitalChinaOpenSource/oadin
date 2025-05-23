import { Tooltip } from 'antd';
import { Input } from 'antd';
import { useViewModel } from './view-model';
import expandSvg from '@/components/icons/expand.svg';
import { SearchIcon } from '@/components/icons';
import McpAdvanceFilter from '../mcp-advance-filter';
import McpList from '../mcp-list';
import styles from './index.module.scss';

export default function McpSquareTab() {
  const vm = useViewModel();

  return (
    <div className={styles.mcpSquareTab}>
      {/* 列表区域 */}
      <div className={styles.mcpSquareContent}>
        <div className={styles.mcpTitle}>
          <div className={styles.mcpTitleText}>
            <span>MCP服务</span>
            <span className={styles.mcpCount}>共 {vm.mcpListData.length} 个</span>
          </div>
          <div className={styles.searchInput}>
            <Input.Search
              allowClear
              placeholder="请输入 MCP 服务名称"
              onSearch={(value) => vm.onMcpInputSearch(value.trim())}
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
        <McpList
          mcpListData={vm.mcpListData}
          onPageChange={vm.handlePageChange}
          handelMcpCardClick={vm.handelMcpCardClick}
          pagination={vm.pagination}
        />
      </div>
      {/* 过滤器 */}
      <McpAdvanceFilter
        collapsed={vm.collapsed}
        setCollapsed={vm.setCollapsed}
        vm={vm}
      />
    </div>
  );
}
