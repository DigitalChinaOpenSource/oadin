import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../sidebar';
import styles from './index.module.scss';
import useOadinServerCheckStore from '@/store/useOadinServerCheckStore';
import { Layout, Tooltip } from 'antd';
import TopHeader from '@/components/main-layout/top-header';
import { ArrowLineLeftIcon, ArrowLineRightIcon } from '@phosphor-icons/react';
import OadinErrorTip from '@/components/oadin-error-tip';

export default function MainLayout() {
  const { Header, Content, Sider } = Layout;
  const [collapsed, setCollapsed] = useState(false);

  // 获取和更新奥丁服务的健康状态
  const { fetchOadinServerStatus } = useOadinServerCheckStore();
  useEffect(() => {
    fetchOadinServerStatus();
    const interval = setInterval(() => {
      fetchOadinServerStatus();
    }, 30000);

    // 清除定时器
    return () => clearInterval(interval);
  }, [fetchOadinServerStatus]);

  const handleCollapse = (collapsed: boolean) => {
    setCollapsed(collapsed);
  };

  const TriggerIcon = () => {
    return (
      <div className={styles.triggerContent}>
        <div className={styles.triggerIconContent}>
          <Tooltip
            title={collapsed ? '展开' : '收起'}
            placement={'top'}
          >
            {collapsed ? (
              <>
                <ArrowLineRightIcon
                  size={20}
                  fill={'#71717D'}
                />
              </>
            ) : (
              <>
                <ArrowLineLeftIcon
                  size={20}
                  fill={'#71717D'}
                />
              </>
            )}
          </Tooltip>
        </div>
      </div>
    );
  };

  return (
    <Layout className={styles.mainLayout}>
      <Header className={styles.header}>
        <TopHeader />
      </Header>
      <OadinErrorTip />
      <Layout>
        <Sider
          theme={'light'}
          style={{ background: 'rgba(255, 255, 255, 0.5)' }}
          collapsible={true}
          onCollapse={handleCollapse}
          trigger={<TriggerIcon />}
        >
          <Sidebar collapsed={collapsed} />
        </Sider>
        <Layout>
          <Content className={styles.content}>
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
}
