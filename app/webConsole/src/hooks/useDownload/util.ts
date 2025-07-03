import { useEffect } from 'react';
import { message } from 'antd';

import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelListStore from '@/store/useModelListStore';
import embedDownloadEventBus from '@/utils/embedDownload';
import { DOWNLOAD_STATUS } from '@/constants';

// 监听浏览器刷新 并执行某些操作
export const usePageRefreshListener = (onRefresh: () => void) => {
  useEffect(() => {
    const handleBeforeUnload = () => {
      onRefresh();
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);
};

// 检查是否达到下载数量限制
export const checkIsMaxDownloadCount = ({ modelList, downList, id }: any) => {
  // 检查是否存在可选项
  if (modelList) {
    // 已下载的直接跳过
    const hasDownloadableModels = modelList.find((item: any) => item.canSelect && item.id === id);
    if (hasDownloadableModels) return false;
  }

  // 检查是否已存在相同模型
  const isModelNotInList = !downList.some((item: any) => item?.id === id);
  console.info(
    downList.filter((item: any) => !(item.canSelect && item.currentDownload === 100)),
    '过滤的下载列表',
  );
  // 验证下载数量限制
  const hasReachedLimit = downList.filter((item: any) => !(item.canSelect || item.currentDownload === 100)).length > 2;

  // 触发限制条件
  if (isModelNotInList && hasReachedLimit) {
    message.warning('您已达到同时下载模型数的上限（3 个）。若您想下载新模型，请先完成或取消部分现有下载任务。');
    return true;
  }
  return false;
};

/**
 * 同时更新下载列表和模型列表中的下载状态
 * @param id 模型ID
 * @param updates 要更新的属性
 */
// 全局变量用于节流控制
const lastUpdateTime = {
  global: 0,
  myModels: 0,
  modelSquare: 0,
};
const pendingUpdates = {
  global: new Map(),
  myModels: new Map(),
  modelSquare: new Map(),
};
const THROTTLE_INTERVAL = 500; // 节流间隔，毫秒

export function updateDownloadStatus(id: string, updates: any) {
  // 获取两个 store 的状态更新函数
  const { setDownloadList, downloadList } = useModelDownloadStore.getState();
  const { setModelListData, setMyModelsList, setModelSquareList } = useModelListStore.getState();

  const now = Date.now();

  // 检查是否是嵌入模型完成下载的特殊情况
  const isEmbedModelCompleted = updates.status === DOWNLOAD_STATUS.COMPLETED && downloadList.some((item) => item.id === id && item.name === 'quentinz/bge-large-zh-v1.5:f16');

  // 更新下载列表
  if (!downloadList || !Array.isArray(downloadList) || downloadList?.length === 0) {
    setDownloadList([]);
  } else {
    let updatedDownloadList = downloadList.map((item) => {
      if (item.id === id) {
        return { ...item, ...updates };
      }
      return item;
    });

    // 处理嵌入模型完成下载的特殊逻辑
    if (isEmbedModelCompleted) {
      embedDownloadEventBus.emit('embedDownloadComplete');
      // 移除所有已完成的下载项
      updatedDownloadList = updatedDownloadList.filter((item) => item.status !== DOWNLOAD_STATUS.COMPLETED);
    }

    setDownloadList(updatedDownloadList);
  }

  // 使用节流方式更新全局模型列表
  if (now - lastUpdateTime.global > THROTTLE_INTERVAL) {
    lastUpdateTime.global = now;
    // 清空待更新队列
    const globalUpdates = new Map(pendingUpdates.global);
    pendingUpdates.global.clear();

    setModelListData((draft: any[]): any[] => {
      if (!Array.isArray(draft) || draft.length === 0) {
        return [];
      }
      return draft.map((item) => {
        // 应用所有累积的更新
        const itemUpdates = globalUpdates.get(item.id);
        if (itemUpdates) {
          return { ...item, ...itemUpdates };
        }
        // 应用当前更新
        if (item.id === id) {
          return { ...item, ...updates };
        }
        return item;
      });
    });
  } else {
    // 将更新放入待处理队列
    const existingUpdates = pendingUpdates.global.get(id) || {};
    pendingUpdates.global.set(id, { ...existingUpdates, ...updates });

    // 设置定时器在节流间隔后处理
    if (pendingUpdates.global.size === 1) {
      setTimeout(
        () => {
          const globalUpdates = new Map(pendingUpdates.global);
          pendingUpdates.global.clear();
          lastUpdateTime.global = Date.now();

          setModelListData((draft: any[]): any[] => {
            if (!Array.isArray(draft) || draft.length === 0) {
              return [];
            }
            return draft.map((item) => {
              const itemUpdates = globalUpdates.get(item.id);
              if (itemUpdates) {
                return { ...item, ...itemUpdates };
              }
              return item;
            });
          });
        },
        THROTTLE_INTERVAL - (now - lastUpdateTime.global),
      );
    }
  }

  // 对"我的模型"和"模型广场"列表使用相同的节流逻辑
  // 这里只展示"我的模型"的实现，"模型广场"类似

  // 更新"我的模型"列表（使用节流）
  if (now - lastUpdateTime.myModels > THROTTLE_INTERVAL) {
    lastUpdateTime.myModels = now;
    const myModelsUpdates = new Map(pendingUpdates.myModels);
    pendingUpdates.myModels.clear();

    setMyModelsList((draft: any[]): any[] => {
      if (!Array.isArray(draft) || draft.length === 0) {
        return [];
      }
      return draft.map((item) => {
        const itemUpdates = myModelsUpdates.get(item.id);
        if (itemUpdates) {
          return { ...item, ...itemUpdates };
        }
        if (item.id === id) {
          return { ...item, ...updates };
        }
        return item;
      });
    });
  } else {
    const existingUpdates = pendingUpdates.myModels.get(id) || {};
    pendingUpdates.myModels.set(id, { ...existingUpdates, ...updates });

    if (pendingUpdates.myModels.size === 1) {
      setTimeout(
        () => {
          const myModelsUpdates = new Map(pendingUpdates.myModels);
          pendingUpdates.myModels.clear();
          lastUpdateTime.myModels = Date.now();

          setMyModelsList((draft: any[]): any[] => {
            if (!Array.isArray(draft) || draft.length === 0) {
              return [];
            }
            return draft.map((item) => {
              const itemUpdates = myModelsUpdates.get(item.id);
              if (itemUpdates) {
                return { ...item, ...itemUpdates };
              }
              return item;
            });
          });
        },
        THROTTLE_INTERVAL - (now - lastUpdateTime.myModels),
      );
    }
  }

  // 更新"模型广场"列表（使用节流）
  if (now - lastUpdateTime.modelSquare > THROTTLE_INTERVAL) {
    lastUpdateTime.modelSquare = now;
    const modelSquareUpdates = new Map(pendingUpdates.modelSquare);
    pendingUpdates.modelSquare.clear();

    setModelSquareList((draft: any[]): any[] => {
      if (!Array.isArray(draft) || draft.length === 0) {
        return [];
      }
      return draft.map((item) => {
        const itemUpdates = modelSquareUpdates.get(item.id);
        if (itemUpdates) {
          return { ...item, ...itemUpdates };
        }
        if (item.id === id) {
          return { ...item, ...updates };
        }
        return item;
      });
    });
  } else {
    const existingUpdates = pendingUpdates.modelSquare.get(id) || {};
    pendingUpdates.modelSquare.set(id, { ...existingUpdates, ...updates });

    if (pendingUpdates.modelSquare.size === 1) {
      setTimeout(
        () => {
          const modelSquareUpdates = new Map(pendingUpdates.modelSquare);
          pendingUpdates.modelSquare.clear();
          lastUpdateTime.modelSquare = Date.now();

          setModelSquareList((draft: any[]): any[] => {
            if (!Array.isArray(draft) || draft.length === 0) {
              return [];
            }
            return draft.map((item) => {
              const itemUpdates = modelSquareUpdates.get(item.id);
              if (itemUpdates) {
                return { ...item, ...itemUpdates };
              }
              return item;
            });
          });
        },
        THROTTLE_INTERVAL - (now - lastUpdateTime.modelSquare),
      );
    }
  }
}
