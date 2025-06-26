import { useState } from 'react';
import { IAccountInfo } from '@/pages/UserCenter/types';
import useAuthStore from '@/store/authStore.ts';
import useLoginStore from '@/store/loginStore.ts';

export const useUserCenterView = () => {
  const [userInfo, setUserInfo] = useState<IAccountInfo>();
  const { logout } = useAuthStore();
  const { setCurrentStep } = useLoginStore();

  // 获取用户信息
  const getUserInfo = async () => {
    await new Promise((resolve) => {
      setTimeout(resolve, 2000);
    });
    const testData = {
      id: 'user-123',
      userName: 'testUser',
      companyName: 'testCompany',
      email: '123@456.com',
      phoneNumber: '1234567890',
      userType: 'enterprise', // 个人账号
      wechatBind: false,
      wechatInfo: {
        openId: 'wx-123456',
        unionId: 'wx-union-123456',
        nickname: 'Test WeChat User',
        avatarUrl: 'https://example.com/avatar.jpg',
      },
      isRealNameAuth: false,
      isEnterpriseAuth: false,
      authInfo: {},
    };
    setUserInfo(testData as IAccountInfo);
  };

  // 修改用户名
  const changeUsername = async (newUsername: string, phoneNumber: string, email?: string) => {
    // 模拟 API 调用
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // 更新用户信息
    setUserInfo(
      (prev) =>
        ({
          ...prev,
          username: newUsername,
        }) as IAccountInfo,
    );
  };

  // 确认注销
  const sureDeleteAccount = async () => {
    // 模拟 API 调用
    await new Promise((resolve) => setTimeout(resolve, 1000));
    // 清除用户信息
    setUserInfo(undefined);
    logout();
    setCurrentStep('personPhone');
  };

  return {
    userInfo,
    getUserInfo,
    changeUsername,
    sureDeleteAccount,
  };
};
