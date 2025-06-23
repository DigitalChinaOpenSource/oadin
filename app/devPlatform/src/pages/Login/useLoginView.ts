import { useState } from 'react';
import useLoginStore from '@/store/loginStore.ts';

export const useLoginView = () => {
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
    initializeWeixinLogin,
  };
};
