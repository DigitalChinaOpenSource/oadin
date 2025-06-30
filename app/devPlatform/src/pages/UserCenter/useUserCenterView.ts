import { useState } from 'react';
import { IAccountInfo, IDeleteAccountProps } from '@/pages/UserCenter/types';
import useAuthStore from '@/store/authStore.ts';
import useLoginStore from '@/store/loginStore.ts';
import { httpRequest } from '@/utils/httpRequest.ts';
import { getUserToken } from '@/utils/getUserToken.ts';
import { IBaseRequestResProps } from '@/pages/Login/types';

export const useUserCenterView = () => {
  const [userInfo, setUserInfo] = useState<IAccountInfo>();
  const { logout, user } = useAuthStore();

  // 获取用户信息
  const getUserInfo = async () => {
    setUserInfo({ ...user } as IAccountInfo);
  };

  // 修改用户名
  const changeUsername = async (params: any) => {
    // 修改个人名称
    if (user.type === 'person') {
      return await httpRequest.post('/individual-login/user/update-nickname', params, { headers: { Authorization: getUserToken() } });
    } else if (user.type === 'enterprise') {
      // 修改企业名称
      return await httpRequest.put('/enterprise/updateName', params, { headers: { Authorization: getUserToken() } });
    }
    return { code: 400, message: '未知用户类型' };
  };

  // 确认注销
  const sureDeleteAccount = async (params: IDeleteAccountProps) => {
    if (user.type === 'person') {
      // 个人注销
      return await httpRequest.post('/individual-login/user/unregister', params, { headers: { Authorization: getUserToken() } });
    } else if (user.type === 'enterprise') {
      // 企业注销
      return await httpRequest.del('/enterprise/account', params, { headers: { Authorization: getUserToken() } });
    }
    return { code: 400, message: '未知用户类型' };
  };

  // 上传实名认证照片
  const uploadRealNameAuthPhoto = async (params: any) => {
    return await httpRequest.post<IBaseRequestResProps>('/enterprise/license', params, { headers: { Authorization: getUserToken() } });
  };

  // 绑定/更改实名认证
  const bindRealNameAuth = async (params: any) => {
    if (user.type === 'person') {
      return await httpRequest.post('/individual-login/verification/submit', params, { headers: { Authorization: getUserToken() } });
    } else if (user.type === 'enterprise') {
      return await httpRequest.post<IBaseRequestResProps>('/enterprise/licenseSubmit', params, { headers: { Authorization: getUserToken() } });
    }
    return { code: 400, message: '未知用户类型' };
  };

  // 修改密码
  const changePassword = async (params: { email: string; oldPassword: string; newPassword: string }) => {
    return await httpRequest.post('/enterprise/password/update', params, { headers: { Authorization: getUserToken() } });
  };

  // 获取用户协议与隐私政策
  const getUserAgreement = async () => {};

  // 退出登录
  const logoutUser = async () => {
    return await httpRequest.post<IBaseRequestResProps>('/individual-login/logout', undefined, { headers: { Authorization: getUserToken() } });
  };

  return {
    userInfo,
    getUserInfo,
    changeUsername,
    sureDeleteAccount,
    bindRealNameAuth,
    setUserInfo,
    getUserAgreement,
    changePassword,
    uploadRealNameAuthPhoto,
    logoutUser,
  };
};
