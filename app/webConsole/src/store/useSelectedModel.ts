import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { IModelDataItem } from '@/types';

const STORAGE_KEY = 'selected_model_store';
export type selectedModelType = IModelDataItem | null;
export interface ISelectedModelStore {
  selectedModel: selectedModelType;
  isSelectedModel: boolean;
  setSelectedModel: (model: selectedModelType | ((currentModel: selectedModelType) => selectedModelType)) => void;
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
      setSelectedModel: (model: selectedModelType | ((currentModel: selectedModelType) => selectedModelType)) => {
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
