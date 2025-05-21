import { useNavigate } from 'react-router-dom';
import { List } from 'antd';
import McpCard from '../mcp-card';
import { useViewModel } from './view-model';
import styles from './index.module.scss';

export default function McpListTab() {
  const navigate = useNavigate();
  const vm = useViewModel();
  const handelClick = (serviceId: string) => {
    navigate(`/mcp-service-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };

  return (
    <div className={styles.mcpManageList}>
      <div className=""></div>
      <div className={styles.mcpListContent}>
        <List
          grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
          dataSource={vm.mcpListData}
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
              <McpCard mcpData={item} />
            </List.Item>
          )}
        />
      </div>
    </div>
  );
}
