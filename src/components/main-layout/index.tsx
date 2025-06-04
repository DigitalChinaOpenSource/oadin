import { useEffect } from 'react';
import { Outlet } from 'react-router-dom';
import Sidebar from '../sidebar';
import styles from './index.module.scss';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';

export default function MainLayout() {
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
    <div className={styles.layout}>
      <Sidebar />
      <div className={styles.content}>
        <Outlet />
      </div>
    </div>
  );
}
