import { useState } from 'react';
import MobileLogin from '@/pages/Login/loginPerson/mobileLogin';
import styles from './index.module.scss';
import WechatLogin from '@/pages/Login/loginPerson/wechatLogin';
import phoneSvg from '@/assets/phone.svg';
import wechatSvg from '@/assets/wechat.svg';

const LoginPerson = () => {
  const [loginType, setLoginType] = useState<'wechat' | 'phone'>('phone');

  return (
    <div className={styles.personLogin}>
      {loginType === 'wechat' && <WechatLogin />}
      {loginType === 'phone' && <MobileLogin onLogin={() => {}} />}
      <div className={styles.loginOperate}>
        <div className={styles.desc}>
          <div className={styles.line}></div>
          <div className={styles.operateDesc}>其它方式登录</div>
          <div className={styles.line}></div>
        </div>
        <div className={styles.operateType}>
          {loginType === 'phone' && (
            <div
              className={styles.wechatType}
              onClick={() => setLoginType('wechat')}
            >
              <img
                src={wechatSvg}
                alt=""
              />
            </div>
          )}
          {loginType === 'wechat' && (
            <div
              className={styles.phoneType}
              onClick={() => setLoginType('phone')}
            >
              <img
                src={phoneSvg}
                alt=""
              />
              <div>手机</div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LoginPerson;
