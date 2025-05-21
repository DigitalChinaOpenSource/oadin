import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelListStore from '@/store/useModelListStore';

/**
 * 同时更新下载列表和模型列表中的下载状态
 * @param id 模型ID
 * @param modelType 模型类型
 * @param updates 要更新的属性
 */
export function updateDownloadStatus(id: number, modelType: string, updates: any) {
  // 获取两个 store 的状态更新函数
  const { setDownloadList } = useModelDownloadStore.getState();
  const { setModelListData } = useModelListStore.getState();
  // 更新下载列表
  setDownloadList((draft: any[]): any[] => {
    if (!draft || !Array.isArray(draft) || draft?.length === 0) {
      return [];
    }
    return draft.map((item) => {
      if (item.modelType === modelType && item.id === id) {
        return { ...item, ...updates };
      }
      return item;
    });
  });

  // 更新模型列表
  setModelListData((draft: any[]): any[] => {
    if (!Array.isArray(draft) || draft.length === 0) {
      return [];
    }
    return draft.map((item) => {
      if (item.id === id) {
        return { ...item, ...updates };
      }
      return item;
    });
  });
}
