import { Outlet } from 'react-router-dom';
import Sidebar from '../sidebar';
import styles from './index.module.scss';
// import DiskInfo from '../disk-info';

const MainLayout = () => {
  return (
    <div className={styles.layout}>
      <Sidebar />
      <div className={styles.content}>
        <Outlet />
      </div>
      {/* <DiskInfo /> */}
    </div>
  );
};

export default MainLayout;