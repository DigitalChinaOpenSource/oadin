import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { IModelDataItem } from '@/types';

const STORAGE_KEY = 'selected_model_store';

export interface ISelectedModelStore {
  selectedModel: IModelDataItem | null;
  isSelectedModel: boolean;
  setSelectedModel: (model: IModelDataItem | null | ((currentModel: IModelDataItem | null) => IModelDataItem | null)) => void;
  setIsSelectedModel: (isSelected: boolean) => void;
}

const useSelectedModelStore = create<ISelectedModelStore>()(
  persist(
    (set) => ({
      selectedModel: null,
      isSelectedModel: false,
      setIsSelectedModel: (isSelected: boolean) => {
        set({ isSelectedModel: isSelected });
      },
      setSelectedModel: (model: IModelDataItem | null | ((currentModel: IModelDataItem | null) => IModelDataItem | null)) => {
        if (typeof model === 'function') {
          set((state) => ({ selectedModel: model(state.selectedModel) }));
        } else {
          set({ selectedModel: model });
        }
      },
    }),
    {
      name: STORAGE_KEY,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        selectedModel: state.selectedModel,
        isSelectedModel: state.isSelectedModel,
      }),
    },
  ),
);

export default useSelectedModelStore;
