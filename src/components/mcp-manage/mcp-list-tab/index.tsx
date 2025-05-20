import { useNavigate } from 'react-router-dom';
import { List } from 'antd';
import McpCard from '../mcp-card';
import { useViewModel } from './view-model';
import Styles from './index.module.scss';

type cardType = {
  title: string;
  content: string;
  icon?: string;
  tags: string[];
  serviceId: string;
};
export default function McpListTab() {
  const navigate = useNavigate();
  const vm = useViewModel();
  const handelClick = (serviceId: string) => {
    navigate(`/mcp-service-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };
  return (
    <div className={Styles.mcpManageList}>
      <div className={Styles.mcpManageListContent}>
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
              <McpCard modelData={item} />
            </List.Item>
          )}
        />
      </div>
    </div>
  );
}
