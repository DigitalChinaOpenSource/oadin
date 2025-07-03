import React, { Dispatch, SetStateAction, useEffect, useRef, useState } from 'react';
import { message, Modal, Tabs, TabsProps } from 'antd';
import styles from './index.module.scss';
import useSelectedModelStore, { selectedModelType } from '@/store/useSelectedModel';
import { ModelSquare } from '@/components/choose-model-dialog/modelSquare.tsx';
import { MyModel } from '@/components/choose-model-dialog/myModel.tsx';
import { useViewModel } from './view-model';
import { getMessageByModel } from '@/i18n';

export interface IChooseModelDialog {
  onCancel: () => void;
  open?: boolean;
  selectedOuterStateModel?: selectedModelType;
  fromWhere?: 'model-checking' | 'chat-container';
}

export interface ISelectedDialogProps {
  isDialog?: boolean;
  selectedStateModel?: selectedModelType;
  setSelecteStatedModel?: Dispatch<SetStateAction<selectedModelType>>;
  fromWhere?: 'model-checking' | 'chat-container';
}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (props: IChooseModelDialog) => {
  const { selectedModel, setSelectedModel } = useSelectedModelStore();
  const { selectedOuterStateModel } = props;
  const [selectedStateModel, setSelecteStatedModel] = useState<selectedModelType>(null);
  const [activeKey, setActiveKey] = useState<string>(props?.fromWhere === 'model-checking' ? 'model-square' : 'my-models');
  const onChange = (activeKey: string) => {
    // 简单设置标签键，不触发额外的刷新
    setActiveKey(activeKey);
  };
  const { fetchChooseModelNotify } = useViewModel();

  // 设置选中的模型
  useEffect(() => {
    setSelecteStatedModel(selectedOuterStateModel || selectedModel);
  }, [selectedModel, selectedOuterStateModel]);

  // 对话框打开时的处理，优化以避免不必要的刷新
  const dialogOpenTimeRef = useRef<number>(0);
  useEffect(() => {
    if (props.open) {
      // 限制频率: 最多3秒一次全局刷新
      const now = Date.now();
      if (now - dialogOpenTimeRef.current > 3000) {
        dialogOpenTimeRef.current = now;
      }
    }
  }, [props.open]);
  const onOk = () => {
    if (selectedStateModel && Object.keys(selectedStateModel).length > 0) {
      setSelectedModel(null);
      const tempParams = { service_name: selectedStateModel.service_name, hybrid_policy: `always_${selectedStateModel.source}` } as any;
      if (selectedStateModel.source === 'local') {
        tempParams.local_provider = selectedStateModel.service_provider_name;
      } else if (selectedStateModel.source === 'remote') {
        tempParams.remote_provider = selectedStateModel.service_provider_name;
      }
      fetchChooseModelNotify(tempParams);
      setSelectedModel(selectedStateModel);
      props.onCancel();
    } else {
      message.warning(
        getMessageByModel('noSelectModel', {
          msg: '请先选择模型，再体验。',
        }),
      );
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
