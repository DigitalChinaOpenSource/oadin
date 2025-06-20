import { useState } from 'react';

export const useLoginView = () => {
  const [loginAccount, setLoginAccount] = useState<'personAccount' | 'enterpriseAccount'>('personAccount');

  // 登录方式切换函数
  const toggleLoginType = (loginAccount: 'personAccount' | 'enterpriseAccount') => {
    setLoginAccount(loginAccount);
  };

  // 微信登录初始化
  const initializeWeixinLogin = (redirect_uri: string) => {
    const wxApiBase = 'https://api-aipc.dcclouds.com/api';
    const state = 'aipc';
    const wx_redirect_uri = `${wxApiBase}/social-login/wechat/callback?redirect=${redirect_uri}`;
    const timeout = setTimeout(() => {
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
    }, 500);

    return () => {
      clearTimeout(timeout);
    };
  };

  return {
    loginAccount,
    setLoginAccount,
    toggleLoginType,
    initializeWeixinLogin,
  };
};
