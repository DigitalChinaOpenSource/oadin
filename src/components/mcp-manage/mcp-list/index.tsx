import { List } from 'antd';
import McpCard from '../mcp-card';
import { IMcpListItem } from '../mcp-square-tab/types';
import styles from './index.module.scss';
import { useNavigate } from 'react-router-dom';

interface IMcpListProps {
  mcpListData: IMcpListItem[];
  pagination: {
    current: number;
    pageSize: number;
    total: number;
  };
  onPageChange?: (current: number, pageSize: number) => void;
  handelMcpCardClick: (serviceId: string | number) => void;
}
export default function McpList(props: IMcpListProps) {
  const { mcpListData, pagination, onPageChange, handelMcpCardClick } = props;

  return (
    <div className={styles.mcpCardList}>
      <List
        grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
        dataSource={mcpListData}
        pagination={
          pagination.total >= 12 && {
            className: styles.mcpListPagination,
            align: 'end',
            ...pagination,
            pageSizeOptions: [12, 24, 48, 96],
            showSizeChanger: true,
            onChange: onPageChange,
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
