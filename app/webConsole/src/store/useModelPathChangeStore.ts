import { create } from 'zustand';

type IStatus = 'init' | 'failed' | 'pending';

interface ModelPathChangeState {
  migratingStatus: IStatus; // 用于存储迁移状态
  // 用于更新迁移状态
  setMigratingStatus: (status: IStatus) => void;
}

const useModelPathChangeStore = create<ModelPathChangeState>((set) => ({
  migratingStatus: 'init',
  setMigratingStatus: (status: IStatus) => set({ migratingStatus: status }),
}));

export default useModelPathChangeStore;
