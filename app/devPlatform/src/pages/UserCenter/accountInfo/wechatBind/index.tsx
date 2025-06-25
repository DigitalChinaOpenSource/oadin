import styles from './index.module.scss';
import { Modal } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import { ScanIcon } from '@phosphor-icons/react';
import { useLoginView } from '@/pages/Login/useLoginView.ts';
import { useEffect } from 'react';

const WechatBind = ({ visible, title, onCancel }: { visible: boolean; title?: string; onCancel: () => void }) => {
  const { initializeWeixinLogin } = useLoginView();

  useEffect(() => {
    const redirect_uri = `${window.location.origin}/login?loginFrom='user-center'`;
    initializeWeixinLogin(redirect_uri, 'login_container_bind');
  }, []);

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
          <div
            id="login_container_bind"
            className={styles.login_content}
          ></div>
        </div>
        <div className={styles.bottom}>绑定微信，可及时获取版本新动态和关键消息通知</div>
      </div>
    </Modal>
  );
};

export default WechatBind;
