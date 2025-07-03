import { create } from 'zustand';
import { IModelDataItem } from '@/types';
import { LOCAL_STORAGE_KEYS } from '@/constants';

interface IModelDownloadStore {
  downloadList: IModelDataItem[];
  setDownloadList: (list: IModelDataItem[]) => void;
}

// 从localStorage获取初始downloadList
const getInitialDownloadList = (): IModelDataItem[] => {
  const storedList = localStorage.getItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
  return storedList && storedList !== 'undefined' ? JSON.parse(storedList) : [];
};

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: getInitialDownloadList(),
  setDownloadList: (list: IModelDataItem[]) => {
    // 处理词嵌入模型的avatar
    const processedList = list.map((item) => {
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

    // 更新store状态
    set({ downloadList: uniqueList });
  },
}));

export default useModelDownloadStore;
