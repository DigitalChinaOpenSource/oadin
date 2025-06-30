import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { IModelDataItem } from '@/types';

const STORAGE_KEY = 'selected_model_store';
export type selectedModelType = IModelDataItem | null;
export interface ISelectedModelStore {
  selectedModel: selectedModelType;
  setSelectedModel: (model: selectedModelType | ((currentModel: selectedModelType) => selectedModelType)) => void;
}

const useSelectedModelStore = create<ISelectedModelStore>()(
  persist(
    (set) => ({
      selectedModel: null,
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
      }),
    },
  ),
);

export default useSelectedModelStore;
