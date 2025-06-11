import React, { useState } from 'react';
import { Modal, Tabs, TabsProps, message } from 'antd';
import styles from './index.module.scss';
import useSelectedModelStore from '@/store/useSelectedModel';
import { ModelSquare } from '@/components/choose-model-dialog/modelSquare.tsx';
import { MyModel } from '@/components/choose-model-dialog/myModel.tsx';
import useChatStore from '@/components/chat-container/store/useChatStore';
import { useViewModel } from './view-model';

export interface IChooseModelDialog {
  onCancel: () => void;
  open?: boolean;
}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (props: IChooseModelDialog) => {
  const { selectedModel, setIsSelectedModel } = useSelectedModelStore();
  const [activeKey, setActiveKey] = useState<string>('model-square');
  const onChange = (activeKey: string) => {
    setActiveKey(activeKey);
  };
  const { currentSessionId } = useChatStore();
  const { fetchChangeModel } = useViewModel();

  const onOk = () => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      setIsSelectedModel(true);
      fetchChangeModel({ sessionId: currentSessionId, modelId: String(selectedModel.id) });
      props.onCancel();
    } else {
      message.warning('请先选择一个模型');
    }
  };
    const items: TabsProps['items'] = [
        {
            key: 'model-square',
            label: '模型广场',
            children: <ModelSquare />,
        },
        {
            key: 'my-models',
            label: '我的模型',
            children: <MyModel />,
        },
    ];

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
        onChange={onChange}
        activeKey={activeKey}
        items={items}
      />
    </Modal>
  );
};
