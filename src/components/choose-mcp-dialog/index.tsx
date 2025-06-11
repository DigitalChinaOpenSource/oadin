import { Button, Checkbox, Modal, ModalProps, Space, Tabs, TabsProps } from 'antd';
import React, { useEffect, useState } from 'react';
import styles from './index.module.scss';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';

export type IChooseMcpDialog = ModalProps & {
  onCancelProps: () => void;
};

interface ITemporaryMcpListItem extends IMcpListItem {
  checked?: boolean;
}

export const ChooseMcpDialog: React.FC<IChooseMcpDialog> = (options: IChooseMcpDialog) => {
  // 控制是否只显示已选中的项
  const [showOnlySelectedMyMcp, setShowOnlySelectedMyMcp] = useState<boolean>(false);
  const [showOnlySelectedMcpList, setShowOnlySelectedMcpList] = useState<boolean>(false);
  const [activeKey, setActiveKey] = useState<string>('myMcp');
  const [selectTemporaryMcpItems, setSelectTemporaryMcpItems] = useState<ITemporaryMcpListItem[]>([]);
  console.info(selectTemporaryMcpItems, 'selectTemporaryMcpItemsselectTemporaryMcpItems');
  const { setSelectMcpList, selectMcpList } = useSelectMcpStore();

  useEffect(() => {
    setSelectTemporaryMcpItems(selectMcpList);
  }, [selectMcpList]);
  const onSelectMcpOk = () => {
    // // 更新全局状态
    setSelectMcpList(selectTemporaryMcpItems);
    // 关闭模态框
    if (options.onCancelProps) {
      options.onCancelProps();
    }
  };
  const mcpDialogTabItems: TabsProps['items'] = [
    {
      key: 'mcpList',
      label: 'MCP广场',
      children: (
        <div className={styles.choose_model_tab_warp}>
          <McpSquareTab
            selectTemporaryMcpItems={selectTemporaryMcpItems}
            setSelectTemporaryMcpItems={setSelectTemporaryMcpItems}
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
            selectTemporaryMcpItems={selectTemporaryMcpItems}
            setSelectTemporaryMcpItems={setSelectTemporaryMcpItems}
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
  const handleCancel = () => {
    setSelectTemporaryMcpItems(selectMcpList);
    // 关闭模态框
    if (options.onCancelProps) {
      options.onCancelProps();
    }
  };

  return (
    <Modal
      className={styles.choose_model}
      okText="立即体验"
      title="选择 MCP"
      style={{ top: 20 }}
      width={1000}
      onCancel={handleCancel}
      onOk={onSelectMcpOk}
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
