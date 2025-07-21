import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import type { UploadFile } from 'antd';

const STORAGE_KEY = 'upload_file_list_store';
export interface IUseUploadFileListStore {
  uploadFileList: UploadFile[];
  setUploadFileList: (files: UploadFile[]) => void;
}

const useUploadFileListStore = create<IUseUploadFileListStore>()(
  persist(
    (set) => ({
      uploadFileList: [],
      setUploadFileList: (files: UploadFile[]) => {
        set({ uploadFileList: files });
      },
    }),
    {
      name: STORAGE_KEY,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        uploadFileList: state.uploadFileList,
      }),
    },
  ),
);

export default useUploadFileListStore;
