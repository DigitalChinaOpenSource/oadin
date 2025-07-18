import { create } from 'zustand';
import { healthRequest } from '@/utils/httpRequest';
import { message } from 'antd';

interface HealthCheckState {
  checkOadinStatus: boolean; // 服务健康状态
  checkOadinServerLoading: boolean; // 请求加载状态
  setCheckOadinServerLoading: (status: boolean) => void;
  fetchOadinServerStatus: () => void; // 手动触发健康检查
}

const useOadinServerCheckStore = create<HealthCheckState>((set) => ({
  checkOadinStatus: true,
  checkOadinServerLoading: false,

  setCheckOadinServerLoading: (loading: boolean) => {
    set({ checkOadinServerLoading: loading });
  },

  fetchOadinServerStatus: async () => {
    set({ checkOadinServerLoading: true });
    try {
      const data = await healthRequest.get('/health');
      if (data?.status === 'UP') {
        set({ checkOadinStatus: true });
        return true;
      } else {
        set({ checkOadinStatus: false });
        return false;
      }
    } catch (error) {
      set({ checkOadinStatus: false });
      return false;
    } finally {
      set({ checkOadinServerLoading: false });
    }
  },
}));

export default useOadinServerCheckStore;
