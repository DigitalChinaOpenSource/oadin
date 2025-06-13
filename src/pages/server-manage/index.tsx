import styles from './index.module.scss';
import { Tabs } from 'antd';
import type { TabsProps } from 'antd';
import ByzeServiceTab from '../../components/byze-service-tab';
import ServiceProviderManage from '../../components/service-provider-tab';
// import ServiceConvention from '../../components/service-convention-tab';

export default function ServiceManage() {
  const items: TabsProps['items'] = [
    {
      key: 'byzeServiceTab',
      label: '奥丁服务',
      children: <ByzeServiceTab />,
    },
    {
      key: 'serviceProviderTab',
      label: '服务提供商管理',
      children: <ServiceProviderManage />,
    },
    // {
    //   key: 'serviceConventionTab',
    //   label: '服务协议',
    //   children: <ServiceConvention />,
    // }
  ];

  return (
    <Tabs
      className={styles.serviceManagePage}
      defaultActiveKey="serviceProvider"
      items={items}
    />
  );
}
