import { create } from 'zustand';
import { IModelDataItem } from '@/types';

interface ISelectMcpStore {
  selectMcpList: IModelDataItem[];
  setSelectMcpList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
}

const useSelectMcpStore = create<ISelectMcpStore>((set, get) => ({
  selectMcpList: [],
  setSelectMcpList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ selectMcpList: list(state.selectMcpList) }));
    } else {
      // 直接设置新状态
      set({ selectMcpList: list });
    }
  },
}));

export default useSelectMcpStore;
