import { create } from 'zustand';
import { IModelDataItem } from '@/types';
import { LOCAL_STORAGE_KEYS } from '@/constants';

interface IModelDownloadStore {
  downloadList: IModelDataItem[];
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => void;
  isDownloadEmbed: boolean;
  setIsDownloadEmbed: (isDownloadEmbed: boolean) => void;
}

// 从localStorage获取初始downloadList
const getInitialDownloadList = (): IModelDataItem[] => {
  const storedList = localStorage.getItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
  return storedList ? JSON.parse(storedList) : [];
};

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: getInitialDownloadList(),
  isDownloadEmbed: false,
  setDownloadList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    if (typeof list === 'function') {
      set((state: IModelDownloadStore) => {
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

        // 保存到localStorage
        localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(uniqueList));
        console.log('更新后的下载列表1111:', uniqueList);
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

      // 保存到localStorage
      localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(uniqueList));
      console.log('更新后的下载列表2222:', uniqueList);
      set({ downloadList: uniqueList });
    }
  },
  setIsDownloadEmbed: (isDownloadEmbed: boolean) => {
    set({ isDownloadEmbed });
  },
}));

export default useModelDownloadStore;
