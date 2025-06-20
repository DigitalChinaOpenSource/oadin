import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { LoginFormValues } from '@/pages/Login/loginForm';
import { IEnterpriseFormValues } from '@/pages/Login/loginEnterprise';

type ILoginStep = 'personPhone' | 'personWechat' | 'enterpriseAccount' | 'forgetPassword' | 'personAuth' | 'enterpriseAuth' | 'bindPhone' | 'createAccount';

interface ILoginStore {
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
      currentStep: 'personPhone',
      personPhoneData: null,
      enterpriseAccountData: null,
      setCurrentStep: (step: ILoginStep) => set({ currentStep: step }),
      setPersonPhoneData: (data: LoginFormValues) => set({ personPhoneData: data }),
      setEnterpriseAccountData: (data: IEnterpriseFormValues) => set({ enterpriseAccountData: data }),
    }),
    {
      name: 'login-storage', // localStorage 中的键名
      partialize: (state) => ({
        currentStep: state.currentStep,
        personPhoneData: state.personPhoneData,
        enterpriseAccountData: state.enterpriseAccountData,
      })
    },
  ),
);

export default useLoginStore;
