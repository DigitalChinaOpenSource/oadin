import React, { useState } from 'react';
import { ArrowLeftOutlined } from '@ant-design/icons';
import styles from './index.module.scss';
import { App, Button, message } from 'antd';
import CancelAuthModal from './cancelAuthModal';
import SureAuthModal from './sureAuthModal';
import { IDeleteAccountProps, IUserType } from '../types';
import { useUserCenterView } from '@/pages/UserCenter/useUserCenterView.ts';
import useAuthStore from '@/store/authStore.ts';
import useLoginStore from '@/store/loginStore.ts';
import { useNavigate } from 'react-router-dom';

const AccountSetting = ({ goBack }: { goBack: () => void }) => {
  const { message } = App.useApp();
  // 控制注销弹窗的显示状态
  const [modalVisible, setModalVisible] = useState<boolean>(false);
  // 控制身份验证弹窗的显示状态
  const [authModalVisible, setAuthModalVisible] = useState<boolean>(false);
  const navigate = useNavigate();

  const { sureDeleteAccount } = useUserCenterView();
  const { logout, user, token } = useAuthStore();
  const { setCurrentStep } = useLoginStore();

  // 注销提示弹窗确认
  const handleConfirm = () => {
    setModalVisible(false);
    setAuthModalVisible(true);
  };

  //   注销身份验证弹窗确认
  const handleAuthConfirm = async (values: IDeleteAccountProps) => {
    console.log('身份验证收集数据', values);
    let params: IDeleteAccountProps = values;
    if (user.type === 'person') {
      params = { ...values, userId: user.uid, token: token };
    }
    // 这里可以添加身份验证成功后的逻辑
    const res = await sureDeleteAccount(params);
    if (res.code === 200) {
      message.success('账号注销成功,请重新登录');
      setAuthModalVisible(false);
      // 清除用户信息
      setCurrentStep('personPhone');
      logout();
      navigate('/login');
    } else {
      message.error('账号注销失败');
    }
  };

  return (
    <div className={styles.accountSetting}>
      <div
        className={styles.goBack}
        onClick={() => goBack()}
      >
        <ArrowLeftOutlined />
        <div>返回</div>
      </div>
      <div className={styles.accountContainer}>
        <div className={styles.accountLeft}>
          <div className={styles.title}>账号注销</div>
          <div className={styles.desc}>账号注销成功后，该账号将无法使用，且账号下的数据将被永久删除</div>
        </div>
        <Button
          type="default"
          onClick={() => setModalVisible(true)}
        >
          注销
        </Button>
      </div>

      {/* 注销确认弹窗 */}
      <CancelAuthModal
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onConfirm={handleConfirm}
      />
      {/* 身份验证 */}
      <SureAuthModal
        visible={authModalVisible}
        onCancel={() => setAuthModalVisible(false)}
        onConfirm={handleAuthConfirm}
        userType={user.type}
      />
    </div>
  );
};

export default AccountSetting;
