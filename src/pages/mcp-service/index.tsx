import styles from './index.module.scss';
import { Tabs } from 'antd';
import type { TabsProps } from 'antd';
import McpManageList from '@/components/mcp-manage/mcp-manage-list';
import MyMcp from '@/components/mcp-manage/my-mcp';
import { useSearchParams } from 'react-router-dom';

export default function McpManage() {
  const items: TabsProps['items'] = [
    {
      key: 'mcpList',
      label: 'MCP广场',
      children: <McpManageList />,
    },
    {
      key: 'myMcp',
      label: '我的MCP',
      children: <MyMcp />,
    },
  ];

  const [searchParams] = useSearchParams();
  const mcpFrom: string | null = searchParams.get('mcpFrom');

  // const onChange = (key: string) => {
  //   // console.log(key);
  // };

  return (
    <Tabs
      className={styles.mcpManage}
      defaultActiveKey={mcpFrom ?? 'mcpList'}
      items={items}
      tabBarStyle={{ borderBottom: 'none' }}
      // onChange={onChange}
    />
    // <div>MCPg</div>
  );
}
