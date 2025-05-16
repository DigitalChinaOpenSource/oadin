import { create } from 'zustand';

interface IModelDownloadStore {
  modelListData: any[];
  setModelListData: (list: any[] | ((currentList: any[]) => any[])) => void;
}

const useModelListStore = create<IModelDownloadStore>((set, get) => ({
  modelListData: [],
  setModelListData: (list: any[] | ((currentList: any[]) => any[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ modelListData: list(state.modelListData) }));
    } else {
      // 直接设置新状态
      set({ modelListData: list });
    }
  },
}));

export default useModelListStore;
