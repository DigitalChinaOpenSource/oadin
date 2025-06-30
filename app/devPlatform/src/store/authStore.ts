import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { IUserType } from '@/pages/UserCenter/types';

export interface User {
  // username?: string;
  // nickname?: string;
  // phone: string;
  // wechatBind?: boolean;
  // phoneBind?: boolean;
  // wechatInfo?: Record<string, any>;
  // email?: string;
  id?: string;
  username?: string;
  nickname?: string;
  type: IUserType;
  avatar?: string;
  email?: string;
  phone: string;
  enterpriseName?: string; // 企业用户特有字段
  isRealNameVerified?: boolean; // 是否实名认证
  idCardFront?: string; // 身份证人像面照片
  idCardBack?: string; // 身份证国徽面照片
  isEnterpriseAuth?: boolean; // 是否企业认证
  wechatBind?: boolean; // 是否绑定微信
  wechatInfo?: Record<string, any>;
  wechatName?: string;
  realNameAuth?: Record<string, any>;
}

interface AuthState {
  isAuthenticated: boolean;
  user: any;
  login: (userData: any, token: string) => void;
  logout: () => void;
  token: string | null; // 可选的 token 字段
  changeUser: (userData: any) => void;
  wechatInfo?: any; // 可选的微信信息字段
  setWechatInfo: (wechatInfo: any) => void;
}

const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      token: null,
      user: null,
      wechatInfo: null,
      login: (userData: any, token: string) => set({ isAuthenticated: true, user: userData, token: token }),
      logout: () => set({ isAuthenticated: false, user: null, token: null, wechatInfo: null }),
      changeUser: (userData: any) => set((state) => ({ user: { ...state.user, ...userData } })),
      setWechatInfo: (wechatInfo: any) => set({ wechatInfo }),
    }),
    {
      name: 'auth-storage', // localStorage 中的键名
    },
  ),
);

export default useAuthStore;
