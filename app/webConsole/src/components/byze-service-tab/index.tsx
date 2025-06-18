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
          <img
            src={faviconPng}
            alt="logo"
            className={styles.logoIcon}
          />
          <div className={styles.infoBlock}>
            <span>奥丁服务状态</span>
            <div className={styles.statusName}>
              {vm.checkByzeStatus ? (
                <>
                  <div className={styles.dot}></div>
                  <span className={styles.used}>启用</span>
                </>
              ) : (
                <>
                  <div className={styles.noCheck}></div>
                  <span className={styles.stopped}>停用</span>
                </>
              )}
            </div>
          </div>
        </div>
        <Button
          icon={vm.checkByzeServerLoading ? null : <RedoOutlined />}
          loading={vm.checkByzeServerLoading}
          className={styles.refreshBlock}
          onClick={vm.handleRefresh}
        >
          检查服务健康状态
        </Button>
      </div>
    </div>
  );
}
