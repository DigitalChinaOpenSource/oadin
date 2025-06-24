import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import LoginForm from '../../loginForm';
import useLoginStore from '@/store/loginStore.ts';

interface MobileLoginProps {
  onLogin?: (phone: string, verificationCode: string) => void;
  loading?: boolean;
}

const MobileLogin: React.FC<MobileLoginProps> = () => {
  return (
    <div className={styles.mobileLoginContainer}>
      {/* 头部标题区 */}
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">手机号登录</span>
          </div>
        </div>
        <span className="subTitle">未注册的手机号将自动创建 OADIN 账号</span>
      </div>

      {/* 表单区域 */}
      <div className="formContainer">
        <LoginForm />
      </div>
    </div>
  );
};

export default MobileLogin;
