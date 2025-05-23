import { create } from 'zustand';
import { ModelDataItem } from '@/types';

interface IModelDownloadStore {
  modelListData: ModelDataItem[];
  setModelListData: (list: ModelDataItem[] | ((currentList: ModelDataItem[]) => ModelDataItem[])) => void;
}

const useModelListStore = create<IModelDownloadStore>((set, get) => ({
  modelListData: [],
  setModelListData: (list: ModelDataItem[] | ((currentList: ModelDataItem[]) => ModelDataItem[])) => {
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
