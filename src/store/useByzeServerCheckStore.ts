import { create } from 'zustand';
import { healthRequest } from '@/utils/httpRequest';
import { message } from 'antd';

interface HealthCheckState {
  checkByzeStatus: boolean; // 服务健康状态
  checkByzeServerLoading: boolean; // 请求加载状态
  fetchByzeServerStatus: () => void; // 手动触发健康检查
}

const useByzeServerCheckStore = create<HealthCheckState>((set) => ({
  checkByzeStatus: false,
  checkByzeServerLoading: false,

  fetchByzeServerStatus: async () => {
    set({ checkByzeServerLoading: true });
    try {
      const data = await healthRequest.get('/health');
      if (data?.status === 'UP') {
        set({ checkByzeStatus: true });
      } else {
        set({ checkByzeStatus: false });
      }
    } catch (error) {
      message.error('白泽服务不可用，请检查服务状态');
      console.error('检查服务健康状态失败:', error);
      set({ checkByzeStatus: false });
    } finally {
      set({ checkByzeServerLoading: false });
    }
  },
}));

export default useByzeServerCheckStore;
