import { create } from 'zustand';

type IStatus = 'init' | 'failed' | 'pending';

interface ModelPathChangeState {
  isPathMigrating: boolean; // 标识模型存储路径是否正在迁移中
  migratingStatus: IStatus; // 用于存储迁移状态
  // 用于更新迁移状态
  setIsPathMigrating: (status: boolean) => void;
  setMigratingStatus: (status: IStatus) => void;
}

const useModelPathChangeStore = create<ModelPathChangeState>((set) => ({
  isPathMigrating: false,
  migratingStatus: 'init',
  setMigratingStatus: (status: IStatus) => set({ migratingStatus: status }),
  setIsPathMigrating: (status: boolean) => set({ isPathMigrating: status }),
}));

export default useModelPathChangeStore;
