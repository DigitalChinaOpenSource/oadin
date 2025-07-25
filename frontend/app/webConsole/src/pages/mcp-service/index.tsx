import styles from './index.module.scss';
import { Tabs } from 'antd';
import type { TabsProps } from 'antd';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';
import { useSearchParams } from 'react-router-dom';

export default function McpManage() {
  const [searchParams] = useSearchParams();
  const mcpFrom: string | null = searchParams.get('mcpFrom');
  const mcpTabItems: TabsProps['items'] = [
    {
      key: 'mcpList',
      label: 'MCP广场',
      children: <McpSquareTab />,
    },
    {
      key: 'myMcp',
      label: '我的MCP',
      children: <MyMcpTab />,
    },
  ];

  return (
    <Tabs
      className={styles.mcpManage}
      defaultActiveKey={mcpFrom ?? 'mcpList'}
      items={mcpTabItems}
      tabBarStyle={{ borderBottom: 'none' }}
    />
  );
}
