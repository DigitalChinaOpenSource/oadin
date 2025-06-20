import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { LoginFormValues } from '@/pages/Login/loginForm';
import { IEnterpriseFormValues } from '@/pages/Login/loginEnterprise';
// 定义登录步骤类型，包含以下可能的步骤：
// - 'personPhone': 个人账号-手机号登录
// - 'personWechat': 个人账号-微信登录
// - 'forgetPassword': 忘记密码
// - 'personAuth': 个人认证
// - 'enterpriseAuth': 企业认证
// - 'bindPhone': 绑定手机
// - 'createAccount': 创建账号
export type LoginType = 'personAccount' | 'enterpriseAccount';
export type ILoginStep = 'personPhone' | 'personWechat' | 'enterpriseAccount' | 'forgetPassword' | 'personAuth' | 'enterpriseAuth' | 'bindPhone' | 'createAccount';

interface ILoginStore {
  // currentLoginType: LoginType;
  // setCurrentLoginType: (type: LoginType) => void;
  currentStep: ILoginStep;
  setCurrentStep: (step: ILoginStep) => void;
  personPhoneData: LoginFormValues | null;
  setPersonPhoneData: (data: LoginFormValues) => void;
  enterpriseAccountData: IEnterpriseFormValues | null;
  setEnterpriseAccountData: (data: IEnterpriseFormValues) => void;
}

const useLoginStore = create<ILoginStore>()(
  persist(
    (set) => ({
      // currentLoginType: 'personAccount', // 默认登录类型为个人账号
      currentStep: 'personPhone',
      personPhoneData: null,
      enterpriseAccountData: null,
      setCurrentStep: (step: ILoginStep) => set({ currentStep: step }),
      setPersonPhoneData: (data: LoginFormValues) => set({ personPhoneData: data }),
      setEnterpriseAccountData: (data: IEnterpriseFormValues) => set({ enterpriseAccountData: data }),
      // setCurrentLoginType: (type: LoginType) => set({ currentLoginType: type }),
    }),
    {
      name: 'login-storage', // localStorage 中的键名
      partialize: (state) => ({
        // currentLoginType: state.currentLoginType,
        currentStep: state.currentStep,
        personPhoneData: state.personPhoneData,
        enterpriseAccountData: state.enterpriseAccountData,
      }),
    },
  ),
);

export default useLoginStore;
