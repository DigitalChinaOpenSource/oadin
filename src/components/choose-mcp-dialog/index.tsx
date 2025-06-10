import { Modal, ModalProps, Tabs, TabsProps } from 'antd';
import React from 'react';
import styles from './index.module.scss';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';

export interface IChooseModelDialog extends ModalProps {}

export const ChooseMcpDialog: React.FC<IChooseModelDialog> = (options: IChooseModelDialog) => {
  const onChange = () => {};
  const mcpDialogTabItems: TabsProps['items'] = [
    {
      key: 'mcpList',
      label: 'MCP广场',
      children: <McpSquareTab isDialog />,
    },
    {
      key: 'myMcp',
      label: '我的MCP',
      children: <MyMcpTab isDialog />,
    },
  ];

  return (
    <Modal
      className={styles.choose_model}
      okText="立即体验"
      title="选择 MCP"
      width={1000}
      {...options}
    >
      <Tabs
        defaultActiveKey="myMcp"
        items={mcpDialogTabItems}
        onChange={onChange}
      />
    </Modal>
  );
};
