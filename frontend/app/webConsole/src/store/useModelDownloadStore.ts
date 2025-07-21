import { create } from 'zustand';
import { IModelDataItem } from '@/types';
import { LOCAL_STORAGE_KEYS } from '@/constants';

interface IModelDownloadStore {
  downloadList: IModelDataItem[];
  setDownloadList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => void;
}

// 从localStorage获取初始downloadList
const getInitialDownloadList = (): IModelDataItem[] => {
  const storedList = localStorage.getItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
  return storedList && storedList !== 'undefined' ? JSON.parse(storedList) : [];
};

const useModelDownloadStore = create<IModelDownloadStore>((set, get) => ({
  downloadList: getInitialDownloadList(),
  setDownloadList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
    // 支持函数式更新
    const newList = typeof list === 'function' ? list(get().downloadList) : list;
    
    console.log('Store更新下载列表，新列表长度:', newList.length, '项目IDs:', newList.map(item => item.id));
    
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

    // 使用 id 作为唯一标识符进行去重，而不是整个对象内容
    const uniqueMap = new Map<string, IModelDataItem>();
    processedList.forEach(item => {
      if (item.id) {
        uniqueMap.set(item.id, item);
      }
    });
    const uniqueList = Array.from(uniqueMap.values());

    // 保存到localStorage
    localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(uniqueList));

    // 更新store状态
    set({ downloadList: uniqueList });
  },
}));

export default useModelDownloadStore;
