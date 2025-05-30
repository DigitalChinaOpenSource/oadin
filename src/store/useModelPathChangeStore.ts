import { create } from 'zustand';

interface ModelPathChangeState {
  isPathMigrating: boolean; // 标识模型存储路径是否正在迁移中
  setMigratingStatus: (status: boolean) => void;
}

const useModelPathChangeStore = create<ModelPathChangeState>((set) => ({
  isPathMigrating: false,

  setMigratingStatus: (status: boolean) => set({ isPathMigrating: status }),
}));

export default useModelPathChangeStore;
