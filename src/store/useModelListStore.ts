import { create } from 'zustand';
import { IModelDataItem } from '@/types';

interface IModelDownloadStore {
  modelListData: IModelDataItem[];
  setModelListData: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
}

const useModelListStore = create<IModelDownloadStore>((set, get) => ({
  modelListData: [],
  setModelListData: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
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
