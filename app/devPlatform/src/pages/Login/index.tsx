import React, { useEffect, useRef, useState } from 'react';
import { useNavigate, useLocation, useSearchParams } from 'react-router-dom';
import useAuthStore from '@/store/authStore';
import styles from './index.module.scss';
import loginBg from '@/assets/login-bg.svg';
import loginDescIcon from '@/assets/login-desc-icon.svg';
import PersonLogin from './loginPerson';
import { Tabs, message, Button } from 'antd';
import type { TabsProps } from 'antd';
import { useLoginView } from './useLoginView';
import LoginEnterprise from '@/pages/Login/loginEnterprise';
import ForgetPassword from '@/pages/Login/loginEnterprise/forgetPassword';
import useLoginStore, { LoginType } from '@/store/loginStore.ts';
import CreateNewAccount from '@/pages/Login/loginEnterprise/createNewAccount';
import './common.css';
import BindPhone from '@/pages/Login/bindPhone';
import AuthPerson from '@/pages/Login/loginPerson/authPerson';
import AuthEnterprise from '@/pages/Login/loginEnterprise/authEnterprise';

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
  const [loginAccount, setLoginAccount] = useState<'personAccount' | 'enterpriseAccount'>('personAccount');
  const { currentStep, setCurrentStep } = useLoginStore();
  const oldStep: any = useRef(null);

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
      children: <LoginEnterprise />,
    },
  ];
  // 获取重定向来源
  // const from = (location.state as { from?: string })?.from || '/app-management';

  const [searchParams] = useSearchParams();
  console.log('searchParams:', searchParams.get('needBindPhone'));

  // 如果需要绑定手机号，则显示手机号绑定页面

  // 处理tab切换
  const handleTabChange = (activeKey: LoginType) => {
    console.log(1111111);
    if (activeKey === 'enterpriseAccount') {
      setCurrentStep(activeKey);
    } else {
      setCurrentStep(oldStep.current);
    }
    setLoginAccount(activeKey);
  };

  useEffect(() => {
    if (!currentStep) setCurrentStep('personPhone');
    if (currentStep === 'personPhone' || currentStep === 'personWechat') {
      setLoginAccount('personAccount');
      oldStep.current = currentStep;
    }
    if (currentStep === 'enterpriseAccount') {
      setLoginAccount('enterpriseAccount');
    }
  }, [currentStep]);

  useEffect(() => {
    setCurrentStep('personPhone');
    const needBindPhone = searchParams.get('needBindPhone');
    const loginFrom = searchParams.get('loginFrom');
    if (needBindPhone === 'true' && loginFrom === 'personWechat') {
      setCurrentStep('bindPhone');
    } else if (needBindPhone === 'false' && loginFrom !== 'personWechat') {
      navigate(`/${loginFrom}`);
    }
  }, []);

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
            {(currentStep === 'personWechat' || currentStep === 'personPhone' || currentStep === 'enterpriseAccount') && (
              <Tabs
                items={items}
                onChange={(activeKey) => handleTabChange(activeKey as LoginType)}
                tabBarGutter={24}
                activeKey={loginAccount}
              />
            )}
            {currentStep === 'forgetPassword' && <ForgetPassword />}
            {currentStep === 'createAccount' && <CreateNewAccount />}
            {currentStep === 'bindPhone' && <BindPhone />}
            {currentStep === 'personAuth' && <AuthPerson />}
            {currentStep === 'enterpriseAuth' && <AuthEnterprise />}
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
