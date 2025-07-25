import { useCallback, useEffect, useRef } from 'react';
import { abortDownload, modelDownloadStream } from './download';
import { updateDownloadStatus, usePageRefreshListener, checkIsMaxDownloadCount } from './util';
import { DOWNLOAD_STATUS, LOCAL_STORAGE_KEYS } from '@/constants';
import { IModelDataItem } from '@/types';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelListStore from '@/store/useModelListStore';
import { getLocalStorageDownList } from '@/utils';

import { IDownParseData } from './types';

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
  const tempDownloadList = downloadList.length > 0 ? downloadList : getLocalStorageDownList(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
  downListRef.current = tempDownloadList;

  // 开始下载
  const fetchDownloadStart = (params: IModelDataItem) => {
    const { id, source, service_provider_name, service_name, name } = params;

    // 最大下载数量
    const isMaxNum = checkIsMaxDownloadCount({
      modelList: useModelListStore.getState().myModelsList,
      downList: downListRef.current,
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

    // 更新下载列表 - 使用函数式更新避免状态竞争
    setDownloadList((currentList: IModelDataItem[]) => {
      console.log('开始下载，当前列表长度:', currentList.length, '要添加的模型ID:', id);

      // 检查是否已存在
      const existingIndex = currentList.findIndex((item: IModelDataItem) => item.id === id);

      if (existingIndex >= 0) {
        // 如果存在，更新状态
        console.log('模型已存在，更新状态:', id);
        return currentList.map((item: IModelDataItem) => (item.id === id ? { ...item, status: IN_PROGRESS } : item));
      } else {
        // 如果不存在，添加新项
        console.log('添加新模型到下载列表:', id, '新列表长度将为:', currentList.length + 1);
        return [...currentList, { ...params, status: IN_PROGRESS }];
      }
    });
    // 同步更新所有模型列表中的状态
    const { setModelListData, setMyModelsList, setModelSquareList } = useModelListStore.getState();

    // 更新函数，统一处理更新逻辑
    const updateModelStatus = (draft: IModelDataItem[]) => draft.map((item: IModelDataItem) => (item.id === id ? { ...item, status: IN_PROGRESS, currentDownload: 0 } : item));

    // 更新所有相关状态
    setModelListData(updateModelStatus);
    setMyModelsList(updateModelStatus);
    setModelSquareList(updateModelStatus);

    modelDownloadStream(paramsTemp, {
      onmessage: (parsedData: IDownParseData) => {
        const { completedsize, progress, status, totalsize, error } = parsedData;
        // 处理错误情况
        if (error) {
          console.log('error===>', error);
          updateDownloadStatus(id, {
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
          // 1. 先直接更新全局状态（不使用节流）
          const { setModelListData, setMyModelsList, setModelSquareList } = useModelListStore.getState();
          const completedUpdates = {
            ...baseUpdates,
            currentDownload: 100,
            status: COMPLETED,
            completedsize: totalsize,
            totalsize,
            can_select: true,
          };
          // 更新或添加已完成的模型到列表
          const updateCompletedModel = (draft: IModelDataItem[]) => {
            // 检查模型是否已在列表中
            const existingModelIndex = draft.findIndex((item) => item.id === id);

            // 如果模型已存在，更新它的状态
            if (existingModelIndex >= 0) {
              return draft.map((item: IModelDataItem) => (item.id === id ? { ...item, ...completedUpdates } : item));
            }
            // 如果模型不存在，将它添加到列表中
            else {
              // 从当前下载列表中获取完整的模型信息
              const downloadItem = downloadList.find((item) => item.id === id);
              if (!downloadItem) return draft; // 安全检查

              const completedModel = {
                ...downloadItem, // 使用下载列表中的完整模型信息
                ...completedUpdates, // 添加完成状态的更新
              };
              console.log('添加新完成模型到列表：', completedModel);
              return [...draft, completedModel];
            }
          };

          // 直接更新三个列表
          console.log(`模型下载完成，ID: ${id}，准备更新列表`);

          // 获取当前列表状态
          const currentModelListData = useModelListStore.getState().modelListData;
          const currentMyModelsList = useModelListStore.getState().myModelsList;
          const currentModelSquareList = useModelListStore.getState().modelSquareList;

          console.log(`当前列表状态: 全部模型(${currentModelListData.length}), 我的模型(${currentMyModelsList.length}), 模型广场(${currentModelSquareList.length})`);

          setModelListData(updateCompletedModel);
          setMyModelsList(updateCompletedModel); // 这里会处理模型不存在的情况
          setModelSquareList(updateCompletedModel);

          console.log('模型列表状态更新完成');

          // 2. 然后使用常规方法更新下载状态
          updateDownloadStatus(id, completedUpdates);

          // 发布一个事件，通知模型下载完成
          // 这里可以添加一个自定义事件，但为了简单，我们使用一个console.log
          console.log('模型下载完成，ID:', id, '状态:', completedUpdates);

          // 立即发出一个事件，可以被其他组件捕获
          const downloadCompleteEvent = new CustomEvent('modelDownloadComplete', {
            detail: { id, completedUpdates, timestamp: Date.now() },
          });
          document.dispatchEvent(downloadCompleteEvent);

          // 延时清理已完成的下载项（给用户一些时间看到完成状态）
          setTimeout(() => {
            const currentList = useModelDownloadStore.getState().downloadList;
            const newList = currentList.filter((item) => item.status !== COMPLETED);
            if (newList.length !== currentList.length) {
              console.log(`延时清理已完成的下载项，清理后剩余: ${newList.length} 个`);
              setDownloadList(newList);
            }
          }, 3000); // 3秒后清理
        } else if (status === 'canceled') {
          updateDownloadStatus(id, {
            ...baseUpdates,
            status: PAUSED,
          });
        } else if (status === 'error') {
          updateDownloadStatus(id, {
            ...baseUpdates,
            status: FAILED,
          });
        } else {
          updateDownloadStatus(id, {
            ...baseUpdates,
            status: IN_PROGRESS,
          });
        }
      },
      onerror: (error: Error) => {
        updateDownloadStatus(id, {
          status: FAILED,
        });
      },
    });
  };

  // 暂停下载
  const fetchDownLoadAbort = useCallback(async (data: { model_name: string }, { id }: { id: string }) => {
    try {
      await abortDownload(data);
      updateDownloadStatus(id, { status: PAUSED });
    } catch (error) {
      console.error('取消或暂停下载失败:', error);
      // 增加错误处理，即使失败也尝试更新UI状态
      updateDownloadStatus(id, { status: FAILED });
    }
  }, []);

  // 刷新页面时从本地存储中获取下载列表
  useEffect(() => {
    const timeout = setTimeout(() => {
      const downListLocal = getLocalStorageDownList(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
      if (downListLocal.length > 0) {
        const updatedList = downListLocal.map((item: IModelDataItem) => ({
          ...item,
        }));
        setDownloadList(updatedList);
      }

      // localStorage.removeItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST);
    }, 150);

    return () => clearTimeout(timeout);
  }, []);

  // 监听浏览器刷新 暂停所有模型下载 并且缓存下载列表
  usePageRefreshListener(() => {
    const downloadingItems = tempDownloadList.filter((item: IModelDataItem) => item.status === IN_PROGRESS);
    console.log('usePageRefreshListener===>', downloadingItems);
    // 存储正在下载的列表
    localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(downloadingItems));

    // 更新所有下载中的项目状态为暂停
    if (downListRef.current?.length > 0) {
      const data = downListRef.current.map((item) => ({
        ...item,
        status: item.status === IN_PROGRESS ? PAUSED : item.status,
      }));
      localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(data));
    }
  });

  const intervalRef = useRef<any>(null);
  // 监听下载列表，处理已完成的下载项，作为兜底处理
  useEffect(() => {
    // 清理之前的定时器
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
    
    // 只要有下载列表就创建定时器，不管是否有已完成的项目
    if (downloadList.length === 0) return;
    
    // 创建定时器，定期检查并清理已完成的下载项
    intervalRef.current = setInterval(() => {
      const currentList = useModelDownloadStore.getState().downloadList;
      const completedItems = currentList.filter((item) => item.status === COMPLETED);
      
      // 如果有已完成的项目，就清除它们
      if (completedItems.length > 0) {
        const newList = currentList.filter((item) => item.status !== COMPLETED);
        console.log(`清理已完成的下载项: ${completedItems.length} 个，剩余: ${newList.length} 个`);
        setDownloadList(newList);
      }
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
