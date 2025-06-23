import { useEffect } from 'react';
import { useLoginView } from '../../useLoginView';
import styles from './index.module.scss';
import wechatIcon from '@/assets/wechatIcon.svg';

const WechatLogin: React.FC = () => {
  const { initializeWeixinLogin } = useLoginView();
  const redirect_uri = `${window.location.origin}/login`;

  useEffect(() => {
    initializeWeixinLogin(redirect_uri);
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
        <div
          id="login_container"
          className={styles.login_content}
        ></div>
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
