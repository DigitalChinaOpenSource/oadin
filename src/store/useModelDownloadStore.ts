import { create } from 'zustand';

interface IModelDownloadStore {
  downloadList: any[];
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => void;
}

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: [],
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ downloadList: list(state.downloadList) }));
    } else {
      // 直接设置新状态
      set({ downloadList: list });
    }
  },
}));

export default useModelDownloadStore;
