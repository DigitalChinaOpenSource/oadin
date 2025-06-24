import React from 'react';
import styles from './index.module.scss';
import './common.css';
import { Button } from 'antd';
import { CaretRightIcon } from '@phosphor-icons/react';
import AccountInfo from './accountInfo';
import RealnameAuth from './realnameAuth';
import ServiceAgreement from './serviceAgreement';

const UserCenter: React.FC = () => {
  return (
    <div className={styles.userCenter}>
      <div className="header">
        <div>账号信息</div>
        <div className={styles.accountOperate}>
          <Button
            type="text"
            icon={
              <CaretRightIcon
                size={16}
                style={{ display: 'flex', alignItems: 'center' }}
              />
            }
            iconPosition="end"
          >
            账号设置
          </Button>
        </div>
      </div>{' '}
      <AccountInfo />
      <div className="header">实名认证</div>
      <RealnameAuth /> <div className="header">服务协议</div>
      <ServiceAgreement />
    </div>
  );
};

export default UserCenter;
