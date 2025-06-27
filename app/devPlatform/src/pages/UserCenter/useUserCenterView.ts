import { useState } from 'react';
import { IAccountInfo, IDeleteAccountProps } from '@/pages/UserCenter/types';
import useAuthStore from '@/store/authStore.ts';
import useLoginStore from '@/store/loginStore.ts';
import { httpRequest } from '@/utils/httpRequest.ts';
import { getUserToken } from '@/utils/getUserToken.ts';

export const useUserCenterView = () => {
  const [userInfo, setUserInfo] = useState<IAccountInfo>();
  const { logout, user } = useAuthStore();
  const { setCurrentStep } = useLoginStore();

  // 获取用户信息
  const getUserInfo = async () => {
    setUserInfo({ ...user, type: 'person' } as IAccountInfo);
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
  const sureDeleteAccount = async (data: IDeleteAccountProps) => {
    // 模拟 API 调用
    await new Promise((resolve) => setTimeout(resolve, 1000));
    const res = await httpRequest.del('/account', data, { Authorization: getUserToken() });
    // 清除用户信息
    setUserInfo(undefined);
    logout();
    setCurrentStep('personPhone');
  };

  // 绑定/更改实名认证
  const bindRealNameAuth = async (values: any) => {};

  // 获取用户协议与隐私政策
  const getUserAgreement = async () => {};

  return {
    userInfo,
    getUserInfo,
    changeUsername,
    sureDeleteAccount,
    bindRealNameAuth,
    setUserInfo,
    getUserAgreement,
  };
};
