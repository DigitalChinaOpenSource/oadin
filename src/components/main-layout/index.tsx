import { useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../sidebar';
import styles from './index.module.scss';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
import { Layout } from 'antd';
import TopHeader from '@/components/main-layout/top-header';

export default function MainLayout() {
  const { Header, Content, Sider } = Layout;

  // 获取和更新白泽服务的健康状态
  const { fetchByzeServerStatus } = useByzeServerCheckStore();
  useEffect(() => {
    const interval = setInterval(() => {
      fetchByzeServerStatus();
    }, 30000);

    // 清除定时器
    return () => clearInterval(interval);
  }, [fetchByzeServerStatus]);

  return (
    <Layout className={styles.mainLayout}>
      <Header className={styles.header}>
        <TopHeader />
      </Header>
      <Layout>
        <Sider
          theme={'light'}
          collapsible={true}
        >
          {/*<div className={styles.layout}>*/}
          <Sidebar />
        </Sider>
        <Layout>
          <div className={styles.content}>
            <Outlet />
          </div>
        </Layout>
        {/*</div>*/}
      </Layout>
    </Layout>
  );
}
