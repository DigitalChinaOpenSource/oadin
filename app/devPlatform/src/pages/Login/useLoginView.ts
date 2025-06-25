import { useEffect, useState } from 'react';
import useLoginStore from '@/store/loginStore.ts';
import { IEnterpriseFormValues } from '@/pages/Login/loginEnterprise';

export const useLoginView = () => {
  // 微信登录初始化
  const initializeWeixinLogin = (redirect_uri: string, id?: string) => {
    const wxApiBase = 'https://api-aipc.dcclouds.com/api';
    const state = 'aipc';
    const wx_redirect_uri = `${wxApiBase}/social-login/wechat/callback?redirect=${redirect_uri}`;
    // 确保 WxLogin 已被正确加载
    if (window.WxLogin) {
      new window.WxLogin({
        id: id || 'login_container',
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

  // 获取手机验证码
  const getPhoneCode = async (phoneNumber: string) => {
    return new Promise((resolve, reject) => {
      // 模拟获取验证码的 API 调用
      setTimeout(() => {
        if (phoneNumber) {
          resolve({ data: true, message: '验证码已发送' });
        } else {
          reject(new Error('手机号不能为空'));
        }
      }, 1000);
    });
  };

  // 获取邮箱验证码
  const getEmailCode = async (email: string) => {
    return new Promise((resolve, reject) => {
      resolve({ data: true, message: '验证码已发送' });
    });
  };

  // 手机号登录/注册
  const loginWithPhone = async (phone: string, verificationCode: string, agreed: boolean) => {
    return new Promise((resolve, reject) => {
      // 模拟登录的 API 调用
      setTimeout(() => {
        if (phone && verificationCode) {
          resolve({ data: true, message: '登录成功' });
        } else {
          reject(new Error('手机号或验证码不能为空'));
        }
      }, 1000);
    });
  };

  // 微信扫码之后，绑定手机号
  const bindPhone = async (phone: string, verificationCode: string, wechatData: any) => {};

  // 企业账号登录
  const loginWithEnterprise = async (email: string, password: string) => {
    return new Promise((resolve, reject) => {
      // 模拟企业账号登录的 API 调用
      setTimeout(() => {
        if (email && password) {
          resolve({ data: true, message: '登录成功' });
        } else {
          reject(new Error('邮箱或密码不能为空'));
        }
      }, 1000);
    });
  };

  // 找回密码 验证code
  const getPassWordWithEmailCode = async (email: string, code: string) => {
    return true;
  };

  // 重置密码
  const resetPassword = async (email: string, newPassword: string) => {
    return new Promise((resolve, reject) => {
      // 模拟重置密码的 API 调用
      setTimeout(() => {
        if (email && newPassword) {
          resolve({ data: true, message: '密码重置成功' });
        } else {
          reject(new Error('邮箱或新密码不能为空'));
        }
      }, 1000);
    });
  };

  // 企业邮箱注册
  const createNewAccount = async (data: IEnterpriseFormValues) => {
    console.log('Creating new account with data:', data);
    return true;
  };

  // 保存个人认证照片

  return {
    initializeWeixinLogin,
    getPhoneCode,
    getEmailCode,
    loginWithPhone,
    bindPhone,
    loginWithEnterprise,
    getPassWordWithEmailCode,
    resetPassword,
    createNewAccount,
  };
};
