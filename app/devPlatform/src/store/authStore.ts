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
  token: string;
}

interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  login: (userData: User) => void;
  logout: () => void;
}

const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      isAuthenticated: false,
      user: null,
      login: (userData: User) => set({ isAuthenticated: true, user: userData }),
      logout: () => set({ isAuthenticated: false, user: null }),
    }),
    {
      name: 'auth-storage', // localStorage 中的键名
    },
  ),
);

export default useAuthStore;
