import { useCallback, useRef, useEffect, useMemo } from 'react';
import { modelDownloadStream, abortDownload } from './download';
import { usePageRefreshListener, checkIsMaxDownloadCount } from './util';
import { DOWNLOAD_STATUS, LOCAL_STORAGE_KEYS } from '../../constants';
import { IModelDataItem } from '../../types';
import useModelDownloadStore from '../../store/useModelDownloadStore';
import useModelListStore from '../../store/useModelListStore';
import { updateDownloadStatus } from './updateDownloadStatus';
import { getLocalStorageDownList } from '@/utils';

/**
 * 下载操作
 * request.body => {
 *   modelName: // 要下载的模型名称
 *   serviceName: "chat" | "embed" | "text_to_image"
 *   serviceSource: "local"
 *   providerName: "local_ollama_chat" | "local_ollama_embed" | "aliyun" // "text_to_image"时传"baidu"
 * }
 * @returns {Object} - 下载相关的状态和方法
 */
export const useDownLoad = () => {
  const { downloadList, setDownloadList } = useModelDownloadStore();
  const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
  const downListRef = useRef<any[]>([]);
  downListRef.current = downloadList;

  // 计算当前下载中的项目
  const downloadingItems = useMemo(() => downloadList.filter((item) => item.status === IN_PROGRESS), [downloadList]);

  // 开始下载
  const fetchDownloadStart = useCallback(
    (params: IModelDataItem) => {
      // modelType: 模型提供商类型
      const { id, type, modelType, source, service_provider_name, service_name, name } = params;

      // 最大下载数量
      const isMaxNum = checkIsMaxDownloadCount({
        downList: downListRef.current,
        modelType,
        id,
      } as any);
      // 检查是否超过最大下载数量
      if (isMaxNum) return;
      // 兼容处理第一条数据id===0的场景
      if (id === undefined || id === null) return;

      const paramsTemp = {
        model_name: name,
        service_name: service_name || 'chat',
        service_source: source || 'local',
        provider_name: service_provider_name || 'local_ollama_chat',
      };

      // 更新下载列表
      setDownloadList([
        ...downloadList.map((item) => (item.type === type && item.modelType === modelType && item.id === id ? { ...item, status: IN_PROGRESS } : item)),
        // 如果不存在则添加新项
        ...(!downloadList.some((item) => item.type === type && item.modelType === modelType && item.id === id) ? [{ ...params, status: IN_PROGRESS }] : []),
      ]);

      // 同步更新模型列表中的状态
      const { setModelListData } = useModelListStore.getState();
      setModelListData((draft) => draft.map((item) => (item.id === id ? { ...item, status: IN_PROGRESS, currentDownload: 0 } : item)));

      modelDownloadStream(paramsTemp, {
        onmessage: (parsedData: any) => {
          console.log('parsedData---------------->', parsedData);
          const { completedsize, progress, status, totalsize, error } = parsedData;

          // 处理错误情况
          if (error) {
            updateDownloadStatus(id, modelType, {
              status: error.includes('aborted') ? PAUSED : FAILED,
            });
            return;
          }

          // 准备基础更新数据
          const baseUpdates = {
            ...(progress && { currentDownload: progress }),
            ...(completedsize && { completedsize }),
            ...(totalsize && { totalsize }),
          };

          // 根据状态更新下载项
          if (status === 'success') {
            updateDownloadStatus(id, modelType, {
              ...baseUpdates,
              currentDownload: 100,
              status: COMPLETED,
              completedsize: totalsize,
              totalsize,
              can_select: true,
            });
            setDownloadList((currentList) => currentList.filter((item) => item.status !== COMPLETED));
          } else if (status === 'canceled') {
            updateDownloadStatus(id, modelType, {
              ...baseUpdates,
              status: PAUSED,
            });
          } else if (status === 'error') {
            updateDownloadStatus(id, modelType, {
              ...baseUpdates,
              status: FAILED,
            });
          } else {
            updateDownloadStatus(id, modelType, {
              ...baseUpdates,
              status: IN_PROGRESS,
            });
          }
        },
        onerror: (error: any) => {
          updateDownloadStatus(id, modelType, {
            status: FAILED,
          });
        },
      } as any);
    },
    [downloadList, setDownloadList],
  );

  // 暂停下载
  const fetchDownLoadAbort = useCallback(async (data: { model_name: string }, { id, modelType }: any) => {
    try {
      await abortDownload(data);
      updateDownloadStatus(id, modelType, { status: PAUSED });
    } catch (error) {
      console.error('取消或暂停下载失败:', error);
      // 增加错误处理，即使失败也尝试更新UI状态
      updateDownloadStatus(id, modelType, { status: FAILED });
    }
  }, []);

  // 刷新页面时从本地存储中获取下载列表
  useEffect(() => {
    const timeout = setTimeout(() => {
      const downListLocal = getLocalStorageDownList(LOCAL_STORAGE_KEYS.DOWN_LIST);
      if (downListLocal.length > 0) {
        // 将所有 IN_PROGRESS 状态的项目更新为 PAUSED
        const updatedList = downListLocal.map((item: any) => ({
          ...item,
          status: item.status === IN_PROGRESS ? PAUSED : item.status,
        }));
        setDownloadList(updatedList);
      }

      localStorage.removeItem(LOCAL_STORAGE_KEYS.DOWN_LIST);
    }, 150);

    return () => clearTimeout(timeout);
  }, []);

  // 监听浏览器刷新 暂停所有模型下载 并且缓存下载列表
  usePageRefreshListener(() => {
    // 存储正在下载的列表
    localStorage.setItem(LOCAL_STORAGE_KEYS.DOWN_LIST, JSON.stringify(downloadingItems));

    // 更新所有下载中的项目状态为暂停
    if (downListRef.current?.length > 0) {
      const data = downListRef.current.map((item) => ({
        ...item,
        status: item.status === IN_PROGRESS ? PAUSED : item.status,
      }));
      localStorage.setItem(LOCAL_STORAGE_KEYS.DOWN_LIST, JSON.stringify(data));
    }
  });

  const intervalRef = useRef<any>(null);
  // 监听下载列表，处理已完成的下载项，作为兜底处理
  useEffect(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    if (downloadList.length === 0) return;
    const hasCompletedItems = downloadList.some((item) => item.status === COMPLETED);
    if (!hasCompletedItems) return;
    // 创建定时器，定期检查并清理已完成的下载项
    intervalRef.current = setInterval(() => {
      console.log('2 秒执行定时器，处理已完成的下载项');
      setDownloadList((currentList) => {
        const hasCompletedItems = currentList.some((item) => item.status === COMPLETED);
        // 如果所有项目都已完成，清空列表并停止定时器
        if (currentList.length > 0 && currentList.every((item) => item.status === COMPLETED)) {
          clearInterval(intervalRef.current);
          intervalRef.current = null;
          return [];
        }
        if (hasCompletedItems) {
          return currentList.filter((item) => item.status !== COMPLETED);
        }
        return currentList;
      });
    }, 2000);

    // 确保组件卸载时清理定时器
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    };
  }, [downloadList]);

  return { fetchDownloadStart, fetchDownLoadAbort };
};
