import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface User {
  username?: string;
  nickname?: string;
  phone: string;
  wechatBind?: boolean;
  phoneBind?: boolean;
  wechatInfo?: Record<string, any>;
  email?: string;
}

interface AuthState {
  isAuthenticated: boolean;
  user: any;
  login: (userData: any, token: string) => void;
  logout: () => void;
  token: string | null; // 可选的 token 字段
  changeUser: (userData: any) => void;
}

const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      token: null,
      user: null,
      login: (userData: any, token: string) => set({ isAuthenticated: true, user: userData, token: token }),
      logout: () => set({ isAuthenticated: false, user: null, token: null }),
      changeUser: (userData: any) => set((state) => ({ user: { ...state.user, ...userData } })),
    }),
    {
      name: 'auth-storage', // localStorage 中的键名
    },
  ),
);

export default useAuthStore;
