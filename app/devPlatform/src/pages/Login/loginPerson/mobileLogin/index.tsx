import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import LoginForm from '../../loginForm';

interface MobileLoginProps {
  onLogin: (phone: string, verificationCode: string) => void;
  loading?: boolean;
}

const MobileLogin: React.FC<MobileLoginProps> = ({ onLogin, loading = false }) => {
  const [phone, setPhone] = useState('');
  const [verificationCode, setVerificationCode] = useState('');
  const [agreed, setAgreed] = useState(false);
  const [countdown, setCountdown] = useState(0);

  // 处理倒计时
  useEffect(() => {
    if (countdown > 0) {
      const timer = setTimeout(() => {
        setCountdown(countdown - 1);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [countdown]);

  // 发送验证码
  const handleSendVerificationCode = () => {
    // 手机号验证
    if (!/^1\d{10}$/.test(phone)) {
      alert('请输入正确的手机号');
      return;
    }

    // 开始倒计时60秒
    setCountdown(60);

    // 这里可以添加发送验证码的API调用
    console.log('发送验证码到手机号：', phone);
  };

  // 提交表单
  const handleSubmit = () => {
    if (!phone || !verificationCode || !agreed) {
      alert('请填写完整信息并同意用户协议');
      return;
    }

    // 调用登录函数
    onLogin(phone, verificationCode);
  };

  // 表单是否完整且同意协议
  const isFormValid = phone && verificationCode && agreed;

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
