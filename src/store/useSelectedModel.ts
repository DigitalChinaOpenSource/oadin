import { create } from 'zustand';

import { IModelDataItem } from '@/types';
export interface ISelectedModelStore {
  selectedModel: IModelDataItem | null;
  setSelectedModel: (model: IModelDataItem | null | ((currentModel: IModelDataItem | null) => IModelDataItem | null)) => void;
}
const useSelectedModelStore = create<ISelectedModelStore>((set) => ({
  selectedModel: null,
  setSelectedModel: (model: IModelDataItem | null | ((currentModel: IModelDataItem | null) => IModelDataItem | null)) => {
    if (typeof model === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ selectedModel: model(state.selectedModel) }));
    } else {
      // 直接设置新状态
      set({ selectedModel: model });
    }
  },
}));
export default useSelectedModelStore;
