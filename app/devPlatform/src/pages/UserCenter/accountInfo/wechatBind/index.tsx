import styles from './index.module.scss';
import { App, Modal } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import { ScanIcon } from '@phosphor-icons/react';
import { useLoginView } from '@/pages/Login/useLoginView.ts';
import { useEffect, useRef, useState } from 'react';
import useAuthStore from '@/store/authStore.ts';

const WechatBind = ({ visible, title, onCancel }: { visible: boolean; title?: string; onCancel: () => void }) => {
  const { setWechatInfo } = useAuthStore();
  const { getWechatLoginQRCode, pollWechatInfo } = useLoginView();
  const [qrcodeUrl, setQrcodeUrl] = useState<string>('');
  const { message } = App.useApp();

  const fetchWechatInfo = async (sessionId: string) => {
    const res = await pollWechatInfo(sessionId);
    console.log('获取微信扫码信息:', res);
    // if (res.success === false && res.message === '微信登录会话已过期，请重新获取二维码') {
    //   await initQrcode();
    // } else
    if (res && res.data && res.data.user && res.success !== false) {
      const { data } = res;
      console.log('获取微信扫码信息成功:', res.data);
      setWechatInfo(data.user);
      //TODO 这里要处理扫码成功后的逻辑，比如保存用户信息、跳转,调用后台绑定的接口等
    }
  };

  // 加载二维码
  const initQrcode = async () => {
    const res = await getWechatLoginQRCode();
    if (res && res.data) {
      setQrcodeUrl(res.data.qrcodeUrl);
      await new Promise((resolve) => setTimeout(resolve, 5000));

      await fetchWechatInfo(res.data.sessionId);
    } else {
      message.error('获取微信登录二维码失败，请刷新重试');
    }
  };

  useEffect(() => {
    if (visible) {
      initQrcode();
    }
  }, [visible]);

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>{title || '绑定微信'}</span>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      closeIcon={<CloseOutlined />}
      centered
      footer={null}
      className={styles.authModal}
      width={480}
      maskClosable={true}
      destroyOnHidden={true}
    >
      <div className={styles.wechatContent}>
        <div className={styles.top}>
          <ScanIcon
            size={16}
            fill={'#9DAABB'}
          />
          <div>请使用</div>
          <div className={styles.descFlag}>微信 APP</div>
          <div>扫码二维码</div>
        </div>
        <div className={styles.wechatBindContainer}>
          {/*<div*/}
          {/*  id="login_container_bind"*/}
          {/*  className={styles.login_content}*/}
          {/*></div>*/}
          <iframe
            className={styles.login_content}
            src={qrcodeUrl}
          />
        </div>
        <div className={styles.bottom}>绑定微信，可及时获取版本新动态和关键消息通知</div>
      </div>
    </Modal>
  );
};

export default WechatBind;
