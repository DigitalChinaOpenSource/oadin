import { create } from 'zustand';
import { IModelDataItem } from '@/types';

interface IModelDownloadStore {
  // 原始的全局模型列表（保留以兼容现有代码）
  modelListData: IModelDataItem[];
  // 我的模型列表
  myModelsList: IModelDataItem[];
  // 模型广场列表
  modelSquareList: IModelDataItem[];
  
  setModelListData: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
  // 设置我的模型列表
  setMyModelsList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
  // 设置模型广场列表
  setModelSquareList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
}

const useModelListStore = create<IModelDownloadStore>((set, get) => ({
  modelListData: [],
  myModelsList: [],
  modelSquareList: [],
  
  setModelListData: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ modelListData: list(state.modelListData) }));
    } else {
      // 直接设置新状态
      set({ modelListData: list });
    }
  },
  
  setMyModelsList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    if (typeof list === 'function') {
      set((state) => ({ myModelsList: list(state.myModelsList) }));
    } else {
      set({ myModelsList: list });
    }
  },
  
  setModelSquareList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    if (typeof list === 'function') {
      set((state) => ({ modelSquareList: list(state.modelSquareList) }));
    } else {
      set({ modelSquareList: list });
    }
  },
}));

export default useModelListStore;
