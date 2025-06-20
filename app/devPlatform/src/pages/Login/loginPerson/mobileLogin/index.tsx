import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import LoginForm from '../../loginForm';

interface MobileLoginProps {
  onLogin: (phone: string, verificationCode: string) => void;
  loading?: boolean;
}

const MobileLogin: React.FC<MobileLoginProps> = ({ onLogin }) => {
  return (
    <div className={styles.mobileLoginContainer}>
      {/* 头部标题区 */}
      <div className={styles.headerSection}>
        <div className={styles.tabContainer}>
          <div className={styles.tabItem}>
            <span className={`${styles.tabTitle} `}>手机号登录</span>
          </div>
        </div>
        <span className={styles.subTitle}>未注册的手机号将自动创建 OADIN 账号</span>
      </div>

      {/* 表单区域 */}
      <div className={styles.formContainer}>
        <LoginForm />
      </div>
    </div>
  );
};

export default MobileLogin;
