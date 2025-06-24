import React, { useState } from 'react';
import { Button } from 'antd';
import styles from './index.module.scss';
import RealAuthIcon from '@/assets/realAuthIcon.svg';
import AuthUploadModal from '@/pages/UserCenter/realnameAuth/authUploadModal';

/**
 * 实名认证模块
 */
const RealnameAuth: React.FC = () => {
  // 控制实名认证弹窗的显示状态
  const [authModalVisible, setAuthModalVisible] = useState<boolean>(false);

  const handleConfirm = () => {
    setAuthModalVisible(false);
  };

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
          onClick={() => setAuthModalVisible(true)}
        >
          立即认证
        </Button>
      </div>
      {/*实名认证弹窗*/}
      <AuthUploadModal
        visible={authModalVisible}
        onCancel={() => setAuthModalVisible(false)}
        onConfirm={handleConfirm}
        userType="person"
      />
    </div>
  );
};

export default RealnameAuth;
