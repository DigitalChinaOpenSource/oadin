import { Modal, ModalProps, Tabs, TabsProps } from 'antd';
import React from 'react';
import styles from './index.module.scss';

export interface IChooseModelDialog extends ModalProps {}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (options: IChooseModelDialog) => {
  const items: TabsProps['items'] = [
    {
      key: '1',
      label: '模型广场',
      children: 'Content of Tab Pane 1',
    },
    {
      key: '2',
      label: '我的模型',
      children: 'Content of Tab Pane 2',
    },
  ];
  const onChange = () => {};
  return (
    <Modal
      className={styles.choose_model}
      okText="立即体验"
      title="选择模型"
      width={1000}
      height={800}
      {...options}
    >
      <Tabs
        defaultActiveKey="1"
        items={items}
        onChange={onChange}
      />
      <p>Some contents...</p>
      <p>Some contents...</p>
      <p>Some contents...</p>
    </Modal>
  );
};
