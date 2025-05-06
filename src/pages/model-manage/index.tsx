import { Tabs } from 'antd';
import type { TabsProps } from 'antd';
import styles from './index.module.scss';
import ModelManageTab from '../../components/model-manage-tab';
import ModelChangeTab from '../../components/model-exchange-tab';

const ModelManage = () => {
  const items: TabsProps['items'] = [
    {
      key: 'modelManageTab',
      label: '模型管理',
      children: <ModelManageTab />,
    },
    // {
    //   key: 'modelChangeTab',
    //   label: '智能模型切换',
    //   children: <ModelChangeTab />,
    // },
  ];

  const onChange = (key: string) => {
    console.log(key);
  };

  return (
    <Tabs
      className={styles.modelManage}
      defaultActiveKey="1"
      items={items}
      onChange={onChange}
    />
  );
};

export default ModelManage;
