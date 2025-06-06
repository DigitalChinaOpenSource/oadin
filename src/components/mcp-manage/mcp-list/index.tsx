import { List } from 'antd';
import McpCard from '../mcp-card';
import { IMcpListItem } from '../mcp-square-tab/types';
import styles from './index.module.scss';
import realLoadingSvg from '@/components/icons/real-loading.svg';

interface IMcpListProps {
  mcpListData: IMcpListItem[];
  pagination: {
    current: number;
    pageSize: number;
    total: number;
  };
  onPageChange?: (current: number, pageSize: number) => void;
  handelMcpCardClick: (serviceId: string | number) => void;
  pageLoading?: boolean;
}
export default function McpList(props: IMcpListProps) {
  const { mcpListData, pagination, onPageChange, handelMcpCardClick, pageLoading = false } = props;

  return (
    <div className={styles.mcpCardList}>
      {pageLoading ? (
        <div className={styles.loading}>
          <img
            src={realLoadingSvg}
            alt="loading"
          />
        </div>
      ) : (
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
      )}
    </div>
  );
}
