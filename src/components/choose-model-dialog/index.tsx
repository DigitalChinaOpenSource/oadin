import React from 'react';
import { Modal, ModalProps, Tabs, TabsProps, message } from 'antd';
import styles from './index.module.scss';
import ModallistContent from '@/components/model-manage-tab/model-list-content';
import { useViewModel } from '@/components/model-manage-tab/view-model.ts';
import ModelSearch from '@/components/model-manage-tab/model-search';
import noDataSvg from '@/components/icons/no-data.svg';
import useSelectedModelStore from '@/store/useSelectedModel';
import { on } from 'events';

export interface IChooseModelDialog {
  onCancel: () => void;
  open?: boolean;
}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (props: IChooseModelDialog) => {
  const vm = useViewModel();
  const { selectedModel } = useSelectedModelStore();
  const items: TabsProps['items'] = [
    {
      key: 'model-square',
      label: '模型广场',
      children: (
        <div>
          <ModelSearch
            modelSearchVal={vm.modelSearchVal}
            modelSourceVal={vm.modelSourceVal}
            onModelSearch={vm.onModelSearch}
            onModelSourceChange={vm.onModelSourceChange}
          />
          <div className={styles.chooseModelList}>
            <ModallistContent isSelectable={true} />
          </div>
        </div>
      ),
    },
    {
      key: 'my-models',
      label: '我的模型',
      children: (
        <div>
          <div className={styles.noData}>
            <div className={styles.noDataIcon}>
              <img
                src={noDataSvg}
                alt="no-data"
              />
            </div>
            <div className={styles.noDataText}>暂无匹配的模型</div>
          </div>
        </div>
      ),
    },
  ];

  const onOk = () => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      props.onCancel();
    } else {
      message.warning('请先选择一个模型');
    }
  };

  return (
    <Modal
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
