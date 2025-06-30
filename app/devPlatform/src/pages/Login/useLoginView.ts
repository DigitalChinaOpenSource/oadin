import { IBaseEnterpriseFormProps, IBaseRequestResProps, IEnterpriseCreateFormValues } from '@/pages/Login/types';
import { httpRequest } from '@/utils/httpRequest.ts';
import { getUserToken } from '@/utils/getUserToken.ts';
import { useEffect } from 'react';

export const useLoginView = () => {
  // 微信登录初始化(走后台接口获取二维码地址使用iframe方式)
  // const initializeWeixinLogin = (redirect_uri: string, id?: string) => {
  //   const wxApiBase = 'https://api-aipc.dcclouds.com/api';
  //   const state = 'aipc';
  //   const wx_redirect_uri = `${wxApiBase}/social-login/wechat/callback`;
  //   // 确保 WxLogin 已被正确加载
  //   if (window.WxLogin) {
  //     new window.WxLogin({
  //       id: id || 'login_container',
  //       appid: 'wxeb558aaad072a194',
  //       scope: 'snsapi_login',
  //       fast_login: 0,
  //       redirect_uri: encodeURIComponent(wx_redirect_uri),
  //       state,
  //       stylelite: 1,
  //       self_redirect: true,
  //     });
  //   } else {
  //     console.error('WxLogin is not loaded. Please check the script loading.');
  //   }
  // };

  // 获取微信登录二维码
  const getWechatLoginQRCode = async () => {
    const qrRes = await httpRequest.get('/individual-login/wechat/qrcode');
    console.log('获取微信登录二维码res', qrRes);
    return qrRes;
  };

  // 根据微信二维码sessionId 获取用户token等扫码信息
  const getWechatInfo = async (sessionId: string) => {
    const res = await httpRequest.get('/individual-login/wechat/status', { sessionId }, { timeout: 120000 });
    console.log('获取微信扫码信息res', res);
    return res;
  };

  // 轮询获取微信扫码信息
  const pollWechatInfo = async (sessionId: string) => {
    let isPolling = true;

    const poll = async () => {
      try {
        const wechatInfo = await getWechatInfo(sessionId);
        if (wechatInfo && wechatInfo.data) {
          console.log('成功获取微信扫码信息:', wechatInfo);
          isPolling = false; // 停止轮询
          return wechatInfo;
        }
      } catch (error: any) {
        console.error('获取微信扫码信息失败:', error);
        if (error.code === 500 || error.code === 400) {
          if (!error.message.includes('timeout')) {
            isPolling = false;
          }
        }
        return { data: error?.response?.data || null, success: false };
      }

      if (isPolling) {
        await poll(); // 继续轮询
      }
    };

    return await poll();
  };

  // 个人/企业 获取手机验证码
  const getPhoneCode = async (params: { phone: string }) => {
    const res = await httpRequest.post<IBaseRequestResProps>('/individual-login/phone/send-code', params);
    console.log('手机验证码res', res);
    return res;
  };

  // 企业 获取邮箱验证码
  const getEmailCode = async (params: { email: string }) => {
    return await httpRequest.post<IBaseRequestResProps>('/enterprise/email-code', params);
  };

  // 手机号登录/注册
  const loginWithPhone = async (params: { phone: string; verifyCode: string; agreed?: boolean; wechatInfo?: any }) => {
    const res = await httpRequest.post<IBaseRequestResProps>('/individual-login/phone/login', params);
    console.log('手机号登录/注册res', res);
    return res;
  };

  // 微信扫码之后，绑定手机号
  const bindPhone = async (phone: string, verificationCode: string, wechatInfo: any) => {};

  // 企业账号登录
  const loginWithEnterprise = async (params: { email: string; password: string }) => {
    return await httpRequest.post('/enterprise/login', params);
  };

  // 找回密码 验证code
  const getPassWordWithEmailCode = async (params: { email: string; emailCode: string }) => {
    return await httpRequest.post('/enterprise/password/email-auth', params);
  };

  // 重置密码
  const resetPassword = async (params: any) => {
    return httpRequest.post<IBaseRequestResProps>('/enterprise/password/reset', params);
  };

  // 企业邮箱注册
  const createNewAccount = async (params: IEnterpriseCreateFormValues) => {
    return await httpRequest.post<IBaseRequestResProps>('/enterprise/register', { ...params, agreePrivacy: params.agreed });
  };

  // 保存个人认证照片

  // 根据token获取个人用户信息
  const getUserInfo = async (params: { token: string }) => {
    const res = await httpRequest.post<IBaseRequestResProps>('/auth/user-info', params);
    console.log('获取微信扫码信息res', res);
    return res.data || null;
  };

  // 根据token获取企业信息
  const getEnterpriseInfo = async () => {
    return await httpRequest.get<IBaseRequestResProps>('/enterprise/user-info', null, { headers: { Authorization: getUserToken() } });
  };

  return {
    // initializeWeixinLogin,
    getPhoneCode,
    getEmailCode,
    loginWithPhone,
    bindPhone,
    loginWithEnterprise,
    getPassWordWithEmailCode,
    resetPassword,
    createNewAccount,
    getUserInfo,
    getWechatLoginQRCode,
    getWechatInfo,
    pollWechatInfo,
    getEnterpriseInfo,
  };
};
