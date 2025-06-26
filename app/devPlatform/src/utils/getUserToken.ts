import useAuthStore from '@/store/authStore.ts';

export const getUserToken = () => {
  const { token } = useAuthStore();
  // 如果 token 存在，返回它；否则返回 null
  return token ? `Bearer ${token}` : null;
};
