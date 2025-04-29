import { create } from "zustand";

interface IModelDownloadStore {
  downloadList: any[];
  setDownloadList: (list: any[]) => void;
}

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: [],
  setDownloadList: (list: any[]) => {
    
  }
}))

export default useModelDownloadStore;