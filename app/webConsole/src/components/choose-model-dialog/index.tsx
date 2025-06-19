import React, { Dispatch, SetStateAction, useEffect, useState } from 'react';
import { Modal, Tabs, TabsProps, message } from 'antd';
import styles from './index.module.scss';
import useSelectedModelStore, { selectedModelType } from '@/store/useSelectedModel';
import { ModelSquare } from '@/components/choose-model-dialog/modelSquare.tsx';
import { MyModel } from '@/components/choose-model-dialog/myModel.tsx';
import useChatStore from '@/components/chat-container/store/useChatStore';
import { useViewModel } from './view-model';

export interface IChooseModelDialog {
  onCancel: () => void;
  open?: boolean;
}

export interface ISelectedDialogProps {
  isDialog?: boolean;
  selectedStateModel?: selectedModelType;
  setSelecteStatedModel?: Dispatch<SetStateAction<selectedModelType>>;
}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (props: IChooseModelDialog) => {
  const { selectedModel, setIsSelectedModel, setSelectedModel } = useSelectedModelStore();
  const [selectedStateModel, setSelecteStatedModel] = useState<selectedModelType>(null);
  const [activeKey, setActiveKey] = useState<string>('model-square');
  const onChange = (activeKey: string) => {
    setActiveKey(activeKey);
  };
  const { currentSessionId } = useChatStore();
  const { fetchChangeModel } = useViewModel();
  useEffect(() => {
    setSelecteStatedModel(selectedModel);
  }, [selectedModel]);
  const onOk = () => {
    if (selectedStateModel && Object.keys(selectedStateModel).length > 0) {
      setIsSelectedModel(true);
      setSelectedModel(selectedStateModel);
      fetchChangeModel({ sessionId: currentSessionId, modelId: String(selectedStateModel.id) });
      props.onCancel();
    } else {
      message.warning('请先选择一个模型');
    }
  };
  const items: TabsProps['items'] = [
    {
      key: 'model-square',
      label: '模型广场',
      children: (
        <ModelSquare
          selectedStateModel={selectedStateModel}
          setSelecteStatedModel={setSelecteStatedModel}
          isDialog={true}
        />
      ),
    },
    {
      key: 'my-models',
      label: '我的模型',
      children: (
        <MyModel
          selectedStateModel={selectedStateModel}
          setSelecteStatedModel={setSelecteStatedModel}
          isDialog={true}
        />
      ),
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
