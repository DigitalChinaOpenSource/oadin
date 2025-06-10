import React from 'react';
import { Modal, Tabs, TabsProps, message } from 'antd';
import styles from './index.module.scss';
import useSelectedModelStore from '@/store/useSelectedModel';
import ModelManageTab from '@/components/model-manage-tab';

export interface IChooseModelDialog {
  onCancel: () => void;
  open?: boolean;
}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (props: IChooseModelDialog) => {
  const { selectedModel, setIsSelectedModel } = useSelectedModelStore();
  const items: TabsProps['items'] = [
    {
      key: 'model-square',
      label: '模型广场',
      children: (
        <ModelManageTab
          isDialog={true}
          isMine={false}
        />
      ),
    },
    {
      key: 'my-models',
      label: '我的模型',
      children: (
        <ModelManageTab
          isDialog={true}
          isMine={true}
        />
      ),
    },
  ];

  const onOk = () => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      setIsSelectedModel(true);
      props.onCancel();
    } else {
      message.warning('请先选择一个模型');
    }
  };

  return (
    <Modal
      style={{ top: 20 }}
      className={styles.choose_model}
      okText="立即体验"
      title="选择模型"
      width={1000}
      onOk={onOk}
      {...props}
    >
      <Tabs
        defaultActiveKey="model-square"
        items={items}
      />
    </Modal>
  );
};
