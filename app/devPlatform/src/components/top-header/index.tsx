import styles from './index.module.scss';
import favicon from '@/assets/favicon.png';
import DefaultUserIcon from '@/assets/userIcon.svg';
import CompanyIcon from '@/assets/companyIcon.svg';
import useAuthStore from '@/store/authStore.ts';
import { Button, Divider, Popover } from 'antd';
import { ExportOutlined } from '@ant-design/icons';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import useLoginStore from '@/store/loginStore.ts';

export default function TopHeader() {
  const { user, changeUser } = useAuthStore();
  const isPerson = user?.type === 'person';
  const navigate = useNavigate();
  const [showUserPopover, setShowUserPopover] = useState(false);
  const { logout } = useAuthStore();
  const { setCurrentStep } = useLoginStore();

  const UserPopover = () => {
    // 进入账号中心
    const handleUserCenter = () => {
      navigate('/user-center');
      setShowUserPopover(false);
    };
    // 退出登录
    const handleLogout = () => {
      setCurrentStep('personWechat');

      logout();
      navigate('/login');
    };

    return (
      <div className={styles.userPopover}>
        <div className={styles.popoverHeader}>
          <img
            src={user?.avatar || (isPerson ? DefaultUserIcon : CompanyIcon)}
            alt=""
          />
          <div className={styles.username}>{isPerson ? user?.nickname || user?.username : user?.enterpriseName}</div>
        </div>
        <Divider />
        <div className={styles.popoverContent}>
          <div
            className={styles.comItem}
            onClick={handleUserCenter}
          >
            账号中心
          </div>
          <div className={styles.comItem}>
            <div>服务协议</div>
            <ExportOutlined size={20} />
          </div>
          <div className={styles.comItem}>
            <div>问题反馈</div>
            <ExportOutlined size={20} />
          </div>
        </div>
        <Button
          type="default"
          onClick={handleLogout}
        >
          退出登录
        </Button>
      </div>
    );
  };

  return (
    <div className={styles.topHeader}>
      <div className={styles.headerLeft}>
        <div className={styles.project}>
          <img
            src={favicon}
            alt=""
          />
          <div>OADIN</div>
        </div>
      </div>
      <div className={styles.personIcon}>
        <Popover
          title={''}
          trigger="click"
          content={<UserPopover />}
          placement={'bottomLeft'}
          open={showUserPopover}
          onOpenChange={setShowUserPopover}
        >
          <img
            src={user?.avatar || (isPerson ? DefaultUserIcon : CompanyIcon)}
            alt=""
          />
        </Popover>
      </div>
    </div>
  );
}
