import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../sidebar';
import styles from './index.module.scss';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
import { Layout, Tooltip } from 'antd';
import TopHeader from '@/components/main-layout/top-header';
import { CaretDoubleLeftIcon, CaretDoubleRightIcon } from '@phosphor-icons/react';
import ByzeErrorTip from '@/components/byze-error-tip';

export default function MainLayout() {
  const { Header, Content, Sider } = Layout;
  const [collapsed, setCollapsed] = useState(false);

  // 获取和更新奥丁服务的健康状态
  const { fetchByzeServerStatus } = useByzeServerCheckStore();
  useEffect(() => {
    const interval = setInterval(() => {
      fetchByzeServerStatus();
    }, 30000);

    // 清除定时器
    return () => clearInterval(interval);
  }, [fetchByzeServerStatus]);

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
                <CaretDoubleRightIcon
                  width={20}
                  height={20}
                  fill="#71717D"
                />
              </>
            ) : (
              <>
                <CaretDoubleLeftIcon
                  width={20}
                  height={20}
                  fill="#71717D"
                />
              </>
            )}
          </Tooltip>
        </div>
      </div>
    );
    // return collapsed ? <span className={styles.triggerIcon}>▶</span> : <span className={styles.triggerIcon}>◀</span>;
  };

  return (
    <Layout className={styles.mainLayout}>
      <Header className={styles.header}>
        <TopHeader />
      </Header>
      <ByzeErrorTip />
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
