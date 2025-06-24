import React from 'react';
import { ArrowUpRightIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import { ExportOutlined } from '@ant-design/icons';

/**
 * 服务协议和隐私政策组件
 */
const ServiceAgreement: React.FC = () => {
  const openAgreement = (type: 'user' | 'privacy') => {
    // 这里可以根据实际需求打开对应的协议页面或弹窗
    if (type === 'user') {
      console.log('打开用户协议');
      //   window.open('/user-agreement', '_blank');
    } else {
      console.log('打开隐私协议');
      //   window.open('/privacy-policy', '_blank');
    }
  };

  return (
    <div className={styles.serviceAgreementContainer}>
      {/* 用户协议 */}
      <div
        className={styles.agreementRow}
        onClick={() => openAgreement('user')}
      >
        <div className={styles.agreementTitleWrapper}>
          <span className={styles.agreementTitle}>用户协议</span>
        </div>
        <ExportOutlined className={styles.arrowIcon} />
      </div>

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 隐私协议 */}
      <div
        className={styles.agreementRow}
        onClick={() => openAgreement('privacy')}
      >
        <div className={styles.agreementTitleWrapper}>
          <span className={styles.agreementTitle}>隐私协议</span>
        </div>
        <ExportOutlined
          size={24}
          className={styles.arrowIcon}
        />
      </div>
    </div>
  );
};

export default ServiceAgreement;
