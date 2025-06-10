import { Checkbox, Modal, ModalProps, Space, Tabs, TabsProps } from 'antd';
import React, { useState } from 'react';
import styles from './index.module.scss';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';

export interface IChooseModelDialog extends ModalProps {}

export const ChooseMcpDialog: React.FC<IChooseModelDialog> = (options: IChooseModelDialog) => {
  // 控制是否只显示已选中的项
  const [showOnlySelectedMyMcp, setShowOnlySelectedMyMcp] = useState<boolean>(false);
  const [showOnlySelectedMcpList, setShowOnlySelectedMcpList] = useState<boolean>(false);
  const [activeKey, setActiveKey] = useState<string>('myMcp');
  console.info(activeKey, '当前选中的页码');
  const mcpDialogTabItems: TabsProps['items'] = [
    {
      key: 'mcpList',
      label: 'MCP广场',
      children: (
        <div className={styles.choose_model_tab_warp}>
          <McpSquareTab
            isDialog
            activeKey={activeKey}
            showOnlySelectedMyMcp={showOnlySelectedMyMcp}
            showOnlySelectedMcpList={showOnlySelectedMcpList}
          />
        </div>
      ),
    },
    {
      key: 'myMcp',
      label: '我的MCP',
      children: (
        <div className={styles.choose_model_tab_warp}>
          <MyMcpTab
            isDialog
            activeKey={activeKey}
            showOnlySelectedMyMcp={showOnlySelectedMyMcp}
            showOnlySelectedMcpList={showOnlySelectedMcpList}
          />
        </div>
      ),
    },
  ];
  const onChange = (activeKey: string) => {
    setActiveKey(activeKey);
  };

  return (
    <Modal
      className={styles.choose_model}
      okText="立即体验"
      title="选择 MCP"
      style={{ top: 20 }}
      width={1000}
      footer={(_, { OkBtn, CancelBtn }) => (
        <div className={styles.choose_model_footer}>
          <Checkbox
            checked={activeKey === 'myMcp' ? showOnlySelectedMyMcp : showOnlySelectedMcpList}
            onChange={(e) => {
              if (activeKey === 'myMcp') {
                setShowOnlySelectedMyMcp(e.target.checked);
              } else {
                setShowOnlySelectedMcpList(e.target.checked);
              }
            }}
          >
            仅显示已选
          </Checkbox>
          <Space>
            <CancelBtn />
            <OkBtn />
          </Space>
        </div>
      )}
      {...options}
    >
      <Tabs
        items={mcpDialogTabItems}
        onChange={onChange}
        activeKey={activeKey}
      />
    </Modal>
  );
};
