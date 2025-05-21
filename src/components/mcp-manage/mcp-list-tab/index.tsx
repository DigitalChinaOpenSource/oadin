import { List } from 'antd';
import McpCard from '../mcp-card';
import { Input } from 'antd';
import { useViewModel } from './view-model';
import { SearchIcon } from '@/components/icons';
import McpAdvanceFilter from '../mcp-advance-filter';
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
              prefix={<SearchIcon />}
              allowClear
              placeholder="请输入模型名称"
              value={vm.mcpSearchVal.keyword}
              onChange={(e) => vm.onMcpInputSearch(e.target.value.trim())}
              style={{ width: 380 }}
            />
          </div>
        </div>
        <div className={styles.mcpCardList}>
          <List
            grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
            dataSource={vm.mcpListDataMock}
            pagination={
              vm.mcpListData.length > 0 && {
                // className: styles.mcpListPagination,
                align: 'end',
                // ...vm.pagination,
                pageSizeOptions: [12, 24, 48, 96],
                showSizeChanger: true,
                // onChange: vm.onPageChange,
                // onShowSizeChange: vm.onShowSizeChange,
              }
            }
            renderItem={(item) => (
              <List.Item>
                <McpCard
                  mcpData={item}
                  handelMcpCardClick={vm.handelMcpCardClick}
                />
              </List.Item>
            )}
          />
        </div>
      </div>
      {/* 过滤器 */}
      <McpAdvanceFilter />
    </div>
  );
}
