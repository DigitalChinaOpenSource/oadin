import React, { useState } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import useAuthStore from '@/store/authStore';
import styles from './index.module.scss';
import loginBg from '@/assets/login-bg.svg';
import loginDescIcon from '@/assets/login-desc-icon.svg';
import PersonLogin from './loginPerson';
import { Tabs, message } from 'antd';
import type { TabsProps } from 'antd';
import { useLoginView } from './useLoginView';

interface LoginFormValues {
  username: string;
  password: string;
}

interface MobileLoginFormValues {
  phone: string;
  verifyCode: string;
}

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login } = useAuthStore();
  const { loginAccount, setLoginAccount } = useLoginView();

  // 个人账号-企业账号
  const items: TabsProps['items'] = [
    {
      key: 'personAccount',
      label: '个人账号',
      children: <PersonLogin />,
    },
    {
      key: 'enterpriseAccount',
      label: '企业账号',
      children: <div>企业账号登录功能正在开发中</div>,
    },
  ];
  // 获取重定向来源
  const from = (location.state as { from?: string })?.from || '/app-management';

  const [searchParams] = useSearchParams();
  console.log('searchParams:', searchParams.get('needBindPhone'));

  // 如果需要绑定手机号，则显示手机号绑定页面

  return (
    <div className={styles.loginPage}>
      {/* 背景图片 */}
      <img
        src={loginBg}
        className={styles.backgroundImage}
        alt="背景"
      />

      <div className={styles.contentContainer}>
        {/* 左侧内容 */}
        <div className={styles.leftContent}>
          <div className={styles.headerContent}>
            <div className={styles.mainTitle}>
              <span className={styles.transColor}>OADIN，</span>
              <span>构建端侧 AI 新范式</span>
            </div>
            <div className={styles.secondTitle}>
              <div className={styles.secondTag}></div>
              <span>低门槛，高效率开发体验</span>
            </div>
          </div>
          {/*左侧描述图片*/}
          <div className={styles.decorationContainer}>
            <div className={styles.decoration}>
              <img
                src={loginDescIcon}
                alt=""
              />
            </div>
          </div>
        </div>
        {/* 登录卡片 */}
        <div className={styles.loginCard}>
          <div className={styles.loginCardContent}>
            <Tabs
              items={items}
              defaultActiveKey={loginAccount}
              onChange={(activeKey) => setLoginAccount(activeKey as 'personAccount' | 'enterpriseAccount')}
              tabBarGutter={24}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
