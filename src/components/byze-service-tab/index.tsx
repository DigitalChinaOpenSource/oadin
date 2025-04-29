import { Button } from 'antd';
import { RedoOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import faviconPng from '../../assets/favicon.png';
import { useViewModel } from './view-model';

export default function ByzeServiceTab() {
  const vm = useViewModel();
  return (
    <div className={styles.byzeServiceTab}>
      <div className={styles.card}>
        <div className={styles.leftBlock}>
          <img src={faviconPng} alt="logo" className={styles.logoIcon} />
          <div className={styles.infoBlock}>
            <span>白泽服务状态</span>
            <span className={styles.statusName}>
              <div className={styles.dot}></div> 启用
            </span>
          </div>
        </div>
        <Button
          icon={<RedoOutlined />}
          className={styles.refreshBlock}
          onClick={vm.handleRefresh}
        >
          检查服务健康状态
        </Button>
      </div>
    </div>
  );
}
