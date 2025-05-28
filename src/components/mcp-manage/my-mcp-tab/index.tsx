import { Input } from 'antd';
import { SearchIcon } from '@/components/icons';
import { useViewModel } from './view-model';
import McpList from '../mcp-list';
import styles from './index.module.scss';

export default function MyMcpTab() {
  const vm = useViewModel();

  return (
    <div className={styles.myMcpListTab}>
      {/* 列表区域 */}
      <div className={styles.myMcpListContent}>
        <div className={styles.mcpTitle}>
          <div className={styles.mcpTitleText}>
            <span>MCP服务</span>
            <span className={styles.mcpCount}>共 {vm.pagination.total} 个</span>
          </div>
          <div className={styles.searchInput}>
            <Input
              allowClear
              placeholder="请输入 MCP 服务名称"
              suffix={
                <div
                  className={styles.searchIcon}
                  onClick={() => vm.onMcpInputSearch(vm.searchVal)}
                >
                  <SearchIcon />
                </div>
              }
              value={vm.searchVal}
              onChange={(e) => vm.setSearchVal(e.target.value)}
              onClear={() => {
                vm.setSearchVal('');
                vm.onMcpInputSearch('');
              }}
              onPressEnter={(e) => vm.onMcpInputSearch(vm.searchVal)}
              style={{ width: 380 }}
            />
          </div>
        </div>
        <McpList
          mcpListData={vm.mcpListData}
          pagination={vm.pagination}
          onPageChange={vm.handlePageChange}
          handelMcpCardClick={vm.handelMcpCardClick}
        />
      </div>
    </div>
  );
}
