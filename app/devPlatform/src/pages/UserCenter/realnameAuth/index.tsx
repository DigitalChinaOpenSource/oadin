import React, { useState } from 'react';
import { App, Button } from 'antd';
import styles from './index.module.scss';
import RealAuthIcon from '@/assets/realAuthIcon.svg';
import AuthUploadModal from '@/pages/UserCenter/realnameAuth/authUploadModal';
import { IAccountInfo, IAccountInfoProps } from '../types';
import { useUserCenterView } from '@/pages/UserCenter/useUserCenterView.ts';
import useAuthStore from '@/store/authStore.ts';

/**
 * 实名认证模块
 */
const RealNameAuth = ({ accountInfo, setUserInfo }: IAccountInfoProps) => {
  const { type: userType, isRealNameVerified, realNameAuth } = accountInfo;
  const { bindRealNameAuth } = useUserCenterView();
  const { message } = App.useApp();
  const { user, changeUser, token } = useAuthStore();

  const haveAuth = userType === 'person' ? isRealNameVerified : realNameAuth?.status === 'approved';
  // 控制实名认证弹窗的显示状态
  const [authModalVisible, setAuthModalVisible] = useState<boolean>(false);

  // 确认保存实名认证
  const handleConfirm = async (values: any) => {
    if (userType === 'person') {
      const params = { userId: user.uid, token: token, idCardFront: values.frontImage[0].url, idCardBack: values.backImage[0].url };
      const authRes = await bindRealNameAuth(params);
      if (authRes.code === 200) {
        changeUser({
          ...user,
          isRealNameVerified: true,
          idCardFront: values.frontImage[0].url,
          idCardBack: values.backImage[0].url,
        });
        setAuthModalVisible(false);
        message.success(authRes.message || '实名认证成功');
      } else {
        message.error(authRes.message || '保存实名认证失败');
      }
    } else {
      const params = { licenseImageUrl: values.enterpriseIcon[0].url };
      const authRes = await bindRealNameAuth(params);
      if (authRes.code === 200) {
        const { realNameAuth } = user;
        changeUser({
          ...user,
          realNameAuth: {
            ...realNameAuth,
            status: 'approved',
            licenseImageUrl: values.enterpriseIcon[0].url,
          },
        });
        setAuthModalVisible(false);
        message.success(authRes.message || '实名认证成功');
      } else {
        message.error(authRes.message || '保存实名认证失败');
      }
    }
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
        accountInfo={accountInfo}
      />
    </div>
  );
};

export default RealNameAuth;
