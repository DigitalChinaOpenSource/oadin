import React from 'react';
import { Button } from 'antd';
import styles from './index.module.scss';
import RealAuthIcon from '@/assets/realAuthIcon.svg';

/**
 * 实名认证模块
 */
const RealnameAuth: React.FC = () => {
  return (
    <div className={styles.realnameAuthContainer}>
      <div className={styles.contentWrapper}>
        <div className={styles.infoSection}>
          {/* 左侧图标组 */}
          <div className={styles.iconContainer}>
            <img
              src={RealAuthIcon}
              alt=""
            />
          </div>

          {/* 文字说明 */}
          <div className={styles.textContainer}>
            <h3 className={styles.title}>实名认证</h3>
            <p className={styles.description}>为保证您能正常使用 OADIN 的服务,请立即进行实名认证</p>
          </div>
        </div>

        {/* 认证按钮 */}
        <Button
          type="primary"
          className={styles.authButton}
        >
          立即认证
        </Button>
      </div>
    </div>
  );
};

export default RealnameAuth;
