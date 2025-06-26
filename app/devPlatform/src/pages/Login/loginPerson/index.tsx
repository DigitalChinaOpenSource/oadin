import MobileLogin from '@/pages/Login/loginPerson/mobileLogin';
import styles from './index.module.scss';
import WechatLogin from '@/pages/Login/loginPerson/wechatLogin';
import phoneSvg from '@/assets/phone.svg';
import wechatSvg from '@/assets/wechat.svg';
import useLoginStore from '@/store/loginStore.ts';

const LoginPerson = () => {
  const { currentStep, setCurrentStep } = useLoginStore();

  return (
    <div className={styles.personLogin}>
      {currentStep === 'personWechat' && <WechatLogin />}
      {currentStep === 'personPhone' && <MobileLogin />}
      <div className={styles.loginOperate}>
        <div className={styles.desc}>
          <div className={styles.line}></div>
          <div className={styles.operateDesc}>其它方式登录</div>
          <div className={styles.line}></div>
        </div>
        <div className={styles.operateType}>
          {currentStep === 'personPhone' && (
            <div
              className={styles.wechatType}
              onClick={() => setCurrentStep('personWechat')}
            >
              <img
                src={wechatSvg}
                alt=""
              />
            </div>
          )}
          {currentStep === 'personWechat' && (
            <div
              className={styles.phoneType}
              onClick={() => setCurrentStep('personPhone')}
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
