import React, { useEffect, useState } from 'react';
import styles from './index.module.scss';
import './common.css';
import { Button, message, Spin } from 'antd';
import { CaretRightIcon } from '@phosphor-icons/react';
import AccountInfo from './accountInfo';
import RealNameAuth from './realnameAuth';
import ServiceAgreement from './serviceAgreement';
import AccountSetting from './accountSetting';
import { useUserCenterView } from '@/pages/UserCenter/useUserCenterView.ts';
import { IAccountInfo } from '@/pages/UserCenter/types';
import useAuthStore from '@/store/authStore.ts';

const UserCenter: React.FC = () => {
  const [showSetting, setShowSetting] = useState<boolean>(false);

  const { userInfo, getUserInfo, setUserInfo } = useUserCenterView();
  const { user } = useAuthStore();

  useEffect(() => {
    getUserInfo();
  }, [user]);

  if (!userInfo)
    return (
      <div className={styles.pageLoading}>
        <Spin />
      </div>
    );

  return (
    <div className={styles.userCenter}>
      {showSetting ? (
        <AccountSetting goBack={() => setShowSetting(false)} />
      ) : (
        <>
          <div className="header">
            <div>账号信息</div>
            <div
              className={styles.accountOperate}
              onClick={() => setShowSetting(true)}
            >
              <Button
                type="text"
                icon={
                  <CaretRightIcon
                    size={16}
                    style={{ display: 'flex', alignItems: 'center' }}
                  />
                }
                iconPosition="end"
              >
                账号设置
              </Button>
            </div>
          </div>
          <AccountInfo
            accountInfo={userInfo as IAccountInfo}
            setUserInfo={setUserInfo}
          />
          <div className="header">实名认证</div>
          <RealNameAuth
            accountInfo={userInfo as IAccountInfo}
            setUserInfo={setUserInfo}
          />
          <div className="header">服务协议</div>
          <ServiceAgreement />
        </>
      )}
    </div>
  );
};

export default UserCenter;
