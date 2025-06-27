import { create } from 'zustand';
import { IModelDataItem } from '@/types';

interface IModelDownloadStore {
  downloadList: IModelDataItem[];
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => void;
  isDownloadEmbed: boolean;
  setIsDownloadEmbed: (isDownloadEmbed: boolean) => void;
}

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: [],
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => {
    if (typeof list === 'function') {
      set((state) => {
        const newList = list(state.downloadList);
        // 处理词嵌入模型的avatar
        const processedList = newList.map((item) => {
          if (item.name === 'quentinz/bge-large-zh-v1.5:f16') {
            return {
              ...item,
              avatar: 'http://120.232.136.73:31619/byzedev/model_avatar/BAAI.png',
            };
          }
          return item;
        });

        // 对象数组进行去重
        const uniqueList = Array.from(new Map(processedList.map((item) => [JSON.stringify(item), item])).values());
        return { downloadList: uniqueList };
      });
    } else {
      const processedList = list.map((item) => {
        // 判断是否为目标模型
        if (item.name === 'quentinz/bge-large-zh-v1.5:f16') {
          return {
            ...item,
            avatar: 'http://120.232.136.73:31619/byzedev/model_avatar/BAAI.png',
          };
        }
        return item;
      });

      const uniqueList = Array.from(new Map(processedList.map((item) => [JSON.stringify(item), item])).values());
      set({ downloadList: uniqueList });
    }
  },
  isDownloadEmbed: false,
  setIsDownloadEmbed: (isDownloadEmbed: boolean) => {
    set({ isDownloadEmbed });
  },
}));

export default useModelDownloadStore;
