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
import { IBasePhoneFormProps } from '@/pages/Login/types';

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
  const { login } = useAuthStore();
  const [loginAccount, setLoginAccount] = useState<'personAccount' | 'enterpriseAccount'>('personAccount');
  const { currentStep, setCurrentStep } = useLoginStore();
  const oldStep: any = useRef(null);
  // 微信扫码登录的结果
  const wechatInfo = useRef<any>(null);

  const { getUserInfo, bindPhone } = useLoginView();

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

  const [searchParams] = useSearchParams();
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

  // 绑定手机号
  const handleBindPhone = (values: IBasePhoneFormProps) => {
    console.log('Binding phone with values: ', values);
    // TODO: （需要绑定手机号肯定是新用户）-绑定成功之后-设置登录状态-跳转到认证界面
    // login()
    setCurrentStep('personAuth');
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
    setCurrentStep('personWechat');
    const success = searchParams.get('success');
    if (success !== 'true') return;
    const needBindPhone = searchParams.get('needBindPhone');
    const loginFrom = searchParams.get('loginFrom');
    const token = searchParams.get('token');
    if (!token) return;
    getUserInfo({ token })
      .then((res) => {
        wechatInfo.current = res;
        if (needBindPhone === 'true' && loginFrom === 'personWechat') {
          setCurrentStep('bindPhone');
        } else if (loginFrom === 'user-center') {
          // TODO 这里需要执行手机号绑定微信操作
          // bindPhone()
          navigate(`/${loginFrom}`);
          message.success('绑定微信成功');
        } else {
          navigate('/');
        }
      })
      .catch((e) => {
        console.error('获取用户信息失败:', e);
        navigate('/login');
      });
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
            <Button onClick={() => setCurrentStep('bindPhone')}>测试按钮</Button>
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
            {currentStep === 'bindPhone' && <BindPhone onConfirm={handleBindPhone} />}
            {currentStep === 'personAuth' && <AuthPerson />}
            {currentStep === 'enterpriseAuth' && <AuthEnterprise />}
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
