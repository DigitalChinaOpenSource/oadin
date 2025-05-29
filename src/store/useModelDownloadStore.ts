import { create } from 'zustand';

interface IModelDownloadStore {
  downloadList: any[];
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => void;
}

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: [],
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => {
        const newList = list(state.downloadList);
        // 通过 JSON.stringify 对对象数组进行去重
        const uniqueList = Array.from(new Map(newList.map((item) => [JSON.stringify(item), item])).values());
        return { downloadList: uniqueList };
      });
    } else {
      // 直接设置新状态，同样需要去重
      const uniqueList = Array.from(new Map(list.map((item) => [JSON.stringify(item), item])).values());
      set({ downloadList: uniqueList });
    }
  },
}));

export default useModelDownloadStore;
