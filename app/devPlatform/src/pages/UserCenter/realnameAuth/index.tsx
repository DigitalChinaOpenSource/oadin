import React, { useState } from 'react';
import { Button } from 'antd';
import styles from './index.module.scss';
import RealAuthIcon from '@/assets/realAuthIcon.svg';
import AuthUploadModal from '@/pages/UserCenter/realnameAuth/authUploadModal';
import { IAccountInfo, IUserType } from '../types';

/**
 * 实名认证模块
 */
const RealNameAuth = ({ accountInfo }: { accountInfo: IAccountInfo }) => {
  const { userType, isRealNameAuth, isEnterpriseAuth } = accountInfo;

  const haveAuth = userType === 'person' ? isRealNameAuth : isEnterpriseAuth;
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
            <h3 className={styles.title}>{userType === 'person' ? '实名认证' : '企业认证'}</h3>
            <p className={styles.description}>{userType === 'person' ? '为保证您能正常使用 OADIN 的服务,请立即进行实名认证' : '采集企业营业执照信息是为了进行资质核验与服务保障'}</p>
          </div>
        </div>

        {/* 认证按钮 */}
        <Button
          type={haveAuth ? 'default' : 'primary'}
          className={styles.authButton}
          onClick={() => setAuthModalVisible(true)}
        >
          {haveAuth ? '变更实名信息' : '立即认证'}
        </Button>
      </div>
      {/*实名认证弹窗*/}
      <AuthUploadModal
        visible={authModalVisible}
        onCancel={() => setAuthModalVisible(false)}
        onConfirm={handleConfirm}
        userType={userType}
        title={haveAuth ? '变更实名信息' : '实名认证'}
      />
    </div>
  );
};

export default RealNameAuth;
