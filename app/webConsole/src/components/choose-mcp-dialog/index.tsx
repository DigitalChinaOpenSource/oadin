import { Checkbox, message, Modal, ModalProps, Space, Tabs, TabsProps } from 'antd';
import React, { useEffect, useState } from 'react';
import styles from './index.module.scss';
import McpSquareTab from '@/components/mcp-manage/mcp-square-tab';
import MyMcpTab from '@/components/mcp-manage/my-mcp-tab';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import { useSelectRemoteHelper, updateMcp } from '@/components/select-mcp/lib/useSelectMcpHelper';
import { getMessageByMcp } from '@/i18n';

export type IChooseMcpDialog = ModalProps & {
  onCancelProps: () => void;
  onSelectMcpOkProps: () => void;
};

export interface ITemporaryMcpListItem extends IMcpListItem {
  checked?: boolean;
}

export const ChooseMcpDialog: React.FC<IChooseMcpDialog> = (options: IChooseMcpDialog) => {
  // 控制是否只显示已选中的项
  const [showOnlySelectedMyMcp, setShowOnlySelectedMyMcp] = useState<boolean>(false);
  const [showOnlySelectedMcpList, setShowOnlySelectedMcpList] = useState<boolean>(false);
  const [activeKey, setActiveKey] = useState<string>('mcpList');
  const [selectTemporaryMcpItems, setSelectTemporaryMcpItems] = useState<ITemporaryMcpListItem[]>([]);
  const { setSelectMcpList, selectMcpList } = useSelectMcpStore();
  const { startMcps, stopMcps } = useSelectRemoteHelper();

  useEffect(() => {
    setSelectTemporaryMcpItems(selectMcpList);
  }, [selectMcpList]);
  const onSelectMcpOk = () => {
    if (selectTemporaryMcpItems.length > 0) {
      /// 启用关闭模型
      const { startList, stopList } = updateMcp(selectMcpList, selectTemporaryMcpItems);

      // 仅当有需要启动的MCP时调用startMcps
      if (startList.length > 0) {
        startMcps({
          ids: startList.map((item) => item.id.toString()),
        });
      }

      // 仅当有需要停止的MCP时调用stopMcps
      if (stopList.length > 0) {
        stopMcps({
          ids: stopList.map((item) => item.id.toString()),
        });
      }
      // // 更新全局状态
      setSelectMcpList(selectTemporaryMcpItems);
      // 关闭模态框
      if (options.onSelectMcpOkProps) {
        options.onSelectMcpOkProps();
      }
      setActiveKey('mcpList');
      setShowOnlySelectedMyMcp(false);
      setShowOnlySelectedMcpList(false);
    } else {
      message.warning(getMessageByMcp('requiredMcp', { msg: '暂无添加好的MCP，请添加后，再体验' }));
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
    setActiveKey('mcpList');
    setShowOnlySelectedMyMcp(false);
    setShowOnlySelectedMcpList(false);
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
      destroyOnHidden={true}
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
