import { useEffect, useState } from 'react';
import { useLoginView } from '../../useLoginView';
import styles from './index.module.scss';
import wechatIcon from '@/assets/wechatIcon.svg';
import { message } from 'antd';
import useLoginStore from '@/store/loginStore.ts';
import { useNavigate } from 'react-router-dom';
import useAuthStore from '@/store/authStore.ts';

const WechatLogin: React.FC = () => {
  const { getWechatLoginQRCode, pollWechatInfo } = useLoginView();
  const { setCurrentStep } = useLoginStore();
  const { setWechatInfo, login } = useAuthStore();
  const [qrcodeUrl, setQrcodeUrl] = useState<string>('');
  const navigate = useNavigate();
  let isUnmounted = false; // 标志变量

  const fetchWechatInfo = async (sessionId: string) => {
    if (isUnmounted) return; // 检查组件是否已销毁
    const res = await pollWechatInfo(sessionId);
    if (isUnmounted) return; // 再次检查组件是否已销毁
    console.log('获取微信扫码信息:', res);
    // if (res.success === false && res.message === '微信登录会话已过期，请重新获取二维码') {
    //   await initQrcode();
    // } else
    if (res && res.data && res.data.user && res.success !== false) {
      const { data } = res;
      console.log('获取微信扫码信息成功:', res.data);
      // 这里可以处理登录成功后的逻辑，比如保存用户信息、跳转等
      if (data.needBindPhone) {
        setWechatInfo(data.user);
        // 需要绑定手机号
        console.log('需要绑定手机号');
        setCurrentStep('bindPhone');
      } else {
        login(data.user, data.token);
        navigate('/app-management');
      }
    }
  };

  // 加载二维码
  const initQrcode = async () => {
    const res = await getWechatLoginQRCode();
    if (isUnmounted) return; // 检查组件是否已销毁
    if (res && res.data) {
      setQrcodeUrl(res.data.qrcodeUrl);
      await new Promise((resolve) => setTimeout(resolve, 5000));

      await fetchWechatInfo(res.data.sessionId);
    } else {
      message.error('获取微信登录二维码失败，请刷新重试');
    }
  };

  useEffect(() => {
    initQrcode();
    return () => {
      isUnmounted = true; // 设置标志变量为 true，表示组件已销毁
    };
  }, []);

  return (
    <div className={styles.wechatLogin}>
      {/* 头部标题区 */}
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">微信扫码登录</span>
          </div>
        </div>
        <span className="subTitle">未注册的微信号将自动创建 OADIN 账号</span>
      </div>
      <div className={styles.wechatLoginContainer}>
        {/*<div*/}
        {/*  id="login_container"*/}
        {/*  className={styles.login_content}*/}
        {/*></div>*/}
        <iframe
          className={styles.login_content}
          src={qrcodeUrl}
        />
        <div className={styles.wechatContainer}>
          <img
            src={wechatIcon}
            alt=""
          />
          <div>微信扫码</div>
        </div>
      </div>
    </div>
  );
};

export default WechatLogin;
