import { Modal } from 'antd';
import styles from './index.module.scss';
import WxFeedBackPng from '@/assets/wxFeedback.png';
import { ScanIcon } from '@phosphor-icons/react';

interface IModelProps {
  open: boolean;
  onClose: () => void;
}
export default function Feedback({ open, onClose }: IModelProps) {
  return (
    <Modal
      title={'官方客户群'}
      centered={true}
      closable={true}
      footer={null}
      onCancel={onClose}
      open={open}
    >
      <div className={styles.feedback}>
        <img
          src={WxFeedBackPng}
          alt=""
          className={styles.codeIcon}
        />
        <div className={styles.comFont}>获取实时更新与专属支持，扫码加入官方群聊！</div>
        <div className={styles.operate}>
          <ScanIcon
            size={20}
            fill={'#9DAABB'}
          />
          <span className={styles.comFont}>请使用</span>
          <span
            className={styles.comFont}
            style={{ color: '#AF5FFE' }}
          >
            微信APP
          </span>
          <span className={styles.comFont}>扫描二维码</span>
        </div>
      </div>
    </Modal>
  );
}
