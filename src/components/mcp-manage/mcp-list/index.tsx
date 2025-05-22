import { List } from 'antd';
import McpCard from '../mcp-card';
import { IMcpListItem } from '../mcp-list-tab/types';
import styles from './index.module.scss';
import { useNavigate } from 'react-router-dom';

interface IMcpListProps {
  mcpListData: IMcpListItem[];
}
export default function McpList(props: IMcpListProps) {
  const navigate = useNavigate();
  const { mcpListData } = props;

  const handelMcpCardClick = (serviceId: number) => {
    navigate(`/mcp-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };

  return (
    <div className={styles.mcpCardList}>
      <List
        grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
        dataSource={mcpListData}
        pagination={
          mcpListData.length > 0 && {
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
              handelMcpCardClick={handelMcpCardClick}
            />
          </List.Item>
        )}
      />
    </div>
  );
}
