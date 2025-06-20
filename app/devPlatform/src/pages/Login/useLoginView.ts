import { useState } from 'react';
// 定义登录步骤类型，包含以下可能的步骤：
// - 'loginTab': 登录选项卡
// - 'forgetPassword': 忘记密码
// - 'personAuth': 个人认证
// - 'enterpriseAuth': 企业认证
// - 'bindPhone': 绑定手机
// - 'createAccount': 创建账号
type ILoginStep = 'loginTab' | 'forgetPassword' | 'personAuth' | 'enterpriseAuth' | 'bindPhone' | 'createAccount';

export const useLoginView = () => {
  const [loginAccount, setLoginAccount] = useState<'personAccount' | 'enterpriseAccount'>('personAccount');

  const [showStep, setShowStep] = useState<ILoginStep>('loginTab');

  // 登录方式切换函数
  const toggleLoginType = (loginAccount: 'personAccount' | 'enterpriseAccount') => {
    setLoginAccount(loginAccount);
  };

  // 微信登录初始化
  const initializeWeixinLogin = (redirect_uri: string) => {
    const wxApiBase = 'https://api-aipc.dcclouds.com/api';
    const state = 'aipc';
    const wx_redirect_uri = `${wxApiBase}/social-login/wechat/callback?redirect=${redirect_uri}`;
    // 确保 WxLogin 已被正确加载
    if (window.WxLogin) {
      new window.WxLogin({
        id: 'login_container',
        appid: 'wxeb558aaad072a194',
        scope: 'snsapi_login',
        fast_login: 0,
        redirect_uri: encodeURIComponent(wx_redirect_uri),
        state,
        stylelite: 1,
      });
    } else {
      console.error('WxLogin is not loaded. Please check the script loading.');
    }
  };

  return {
    loginAccount,
    setLoginAccount,
    toggleLoginType,
    initializeWeixinLogin,
    showStep,
    setShowStep,
  };
};
