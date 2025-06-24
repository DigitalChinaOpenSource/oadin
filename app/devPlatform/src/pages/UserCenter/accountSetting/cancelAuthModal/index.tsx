import React, { useState, useEffect } from 'react';
import { Modal, Button } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import styles from './index.module.scss';

interface CancelAuthModalProps {
  visible: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

/**
 * 账号注销确认弹窗
 */
const CancelAuthModal: React.FC<CancelAuthModalProps> = ({ visible, onCancel, onConfirm }) => {
  // 倒计时相关状态
  const [countdown, setCountdown] = useState<number>(10);
  const [isCountingDown, setIsCountingDown] = useState<boolean>(true);

  // 当弹窗打开时，开始倒计时
  useEffect(() => {
    let timer: NodeJS.Timeout;

    if (visible && isCountingDown && countdown > 0) {
      timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev === 1) {
            setIsCountingDown(false);
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    }

    // 组件卸载或弹窗关闭时清除定时器
    return () => {
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [visible, isCountingDown, countdown]);

  // 重置倒计时
  const resetCountdown = () => {
    setCountdown(10);
    setIsCountingDown(true);
  };

  // 关闭弹窗时重置倒计时
  const handleCancel = () => {
    resetCountdown();
    onCancel();
  };

  // 确认注销
  const handleConfirm = () => {
    onConfirm();
    resetCountdown();
  };

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>重要提示</span>
        </div>
      }
      open={visible}
      onCancel={handleCancel}
      footer={null}
      closeIcon={<CloseOutlined />}
      centered
      className={styles.cancelModal}
      width={480}
      maskClosable={false}
    >
      <div className={styles.contentText}>
        <strong>很遗憾即将与您告别！</strong>
        <p>在你决定注销账号之前，我们希望了解更多。请告诉我们是否遇到任何问题，或是有什么我们可以改进的地方。我们会尽全力提供更好的体验，帮助你解决问题。</p>
        <strong>注销账号会导致以下影响：</strong>
        <ul>
          <li>你将无法访问你的个人数据和设置。</li>
          <li>你将失去与平台相关的所有服务和功能。</li>
          <li>注销后若需再次使用 Oadin，需重新注册新账号。</li>
        </ul>
        <p>如果你有任何疑问或决定继续使用我们的服务，欢迎随时联系我们。</p>
        <strong>如果你仍决定注销账号，请点击下方按钮完成操作。</strong>
      </div>

      <div className={styles.buttonGroup}>
        <Button onClick={handleCancel}>暂不注销</Button>
        <Button
          type="primary"
          onClick={handleConfirm}
          disabled={isCountingDown}
          danger
        >
          {isCountingDown ? `继续注销（${countdown}s）` : '继续注销'}
        </Button>
      </div>
    </Modal>
  );
};

export default CancelAuthModal;
