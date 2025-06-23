import { useEffect } from 'react';
import { message } from 'antd';

import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelListStore from '@/store/useModelListStore';

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
export function updateDownloadStatus(id: string, updates: any) {
  // 获取两个 store 的状态更新函数
  const { setDownloadList, downloadList } = useModelDownloadStore.getState();
  const { setModelListData } = useModelListStore.getState();

  // 查找要更新的项目
  const itemToUpdate = downloadList?.find((item) => item.id === id);

  // 检查status字段是否发生变化
  let statusChanged = false;
  if (itemToUpdate && 'status' in updates) {
    statusChanged = itemToUpdate.status !== updates.status;
  }

  // 更新下载列表
  setDownloadList((draft: any[]): any[] => {
    if (!draft || !Array.isArray(draft) || draft?.length === 0) {
      return [];
    }
    return draft.map((item) => {
      if (item.id === id) {
        return { ...item, ...updates };
      }
      return item;
    });
  });

  // 只有当status状态发生变化时才更新模型列表
  if (statusChanged) {
    console.log('更新模型列表中的下载状态', id, updates);
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
}
