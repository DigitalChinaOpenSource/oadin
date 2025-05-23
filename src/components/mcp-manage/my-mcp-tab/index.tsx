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
            <span className={styles.mcpCount}>共 {vm.myMcpListData.length} 个</span>
          </div>
          <div className={styles.searchInput}>
            <Input
              allowClear
              placeholder="请输入 MCP 服务名称"
              suffix={
                <div
                  className={styles.searchIcon}
                  onClick={() => vm.onMcpInputSearch()}
                >
                  <SearchIcon />
                </div>
              }
              value={vm.mcpSearchVal.keyword}
              onChange={(e) => vm.onMcpInputChange(e.target.value)}
              style={{ width: 380 }}
            />
          </div>
        </div>
        <McpList
          mcpListData={vm.myMcpListData}
          pagination={vm.pagination}
          onPageChange={vm.onPageChange}
          onShowSizeChange={vm.onShowSizeChange}
        />
      </div>
    </div>
  );
}
