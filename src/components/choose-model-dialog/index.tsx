import { Modal, ModalProps, Tabs, TabsProps } from 'antd';
import React from 'react';
import styles from './index.module.scss';
import ModallistContent from '@/components/model-manage-tab/model-list-content';
import { useViewModel } from '@/components/model-manage-tab/view-model.ts';
import ModelSearch from '@/components/model-manage-tab/model-search';
import noDataSvg from '@/components/icons/no-data.svg';

export interface IChooseModelDialog extends ModalProps {}

export const ChooseModelDialog: React.FC<IChooseModelDialog> = (options: IChooseModelDialog) => {
  const vm = useViewModel();
  const items: TabsProps['items'] = [
    {
      key: '1',
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
            <ModallistContent />
          </div>
        </div>
      ),
    },
    {
      key: '2',
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
  const onChange = () => {};
  return (
    <Modal
      className={styles.choose_model}
      okText="立即体验"
      title="选择模型"
      width={1000}
      {...options}
    >
      <Tabs
        defaultActiveKey="1"
        items={items}
        onChange={onChange}
      />
    </Modal>
  );
};
