import { useCallback, useRef, useEffect, useMemo } from 'react';
import { useImmer } from 'use-immer';
import { startDownLoad, abortDownload } from './download';
import { usePageRefreshListener, checkIsMaxDownloadCount } from './util';
import { DOWNLOAD_STATUS } from '../../constants';
import downLoadCompleteEventBus from '../../utils/downloadEvent';

// 本地存储键名常量
const LOCAL_STORAGE_KEYS = {
  DOWN_LIST: 'downList',
  REFRESH_DOWN_LIST: 'refreshDownList',
};

/**
 * 从本地存储获取下载列表
 * @param {string} key - 本地存储键名
 * @returns {Array} - 解析后的下载列表
 */
function getLocalStorageDownList(key: string) {
  try {
    const data = localStorage.getItem(key);
    return data ? JSON.parse(data) : [];
  } catch (error) {
    console.error(`Failed to parse ${key} from localStorage:`, error);
    return [];
  }
}

/**
 * 更新下载项状态的工具函数
 * @param {Array} list - 当前下载列表
 * @param {Function} setList - 设置下载列表的函数
 * @param {string|number} id - 下载项ID
 * @param {string} modelType - 模型类型
 * @param {Object} updates - 要更新的属性
 */
function updateDownloadItem(setList: any, id: number, modelType: string, updates: any) {
  setList((draft: any) => {
    const item = draft.find((item: any) => item.modelType === modelType && item.id === id);
    if (item) {
      Object.assign(item, updates);
    }
  });
}

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
  const [downList, setDownList] = useImmer<any[]>([]);
  const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
  const downListRef = useRef<any[]>([]);
  downListRef.current = downList;

  // 计算当前下载中的项目
  const downloadingItems = useMemo(() => downList.filter((item) => item.status === IN_PROGRESS), [downList]);

  // 开始下载
  const downLoadStart = useCallback((params: any, downObj: any) => {
    // modelType: 模型提供商类型
    const { id, type, modelType, serviceProviderName, serviceName } = downObj;

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

    // 处理参数
    const paramsTemp = {
      ...params,
      serviceName: serviceName,
      serviceSource: 'local',
      providerName: serviceProviderName || 'local_ollama_chat',
    };

    // 更新下载列表
    setDownList((draft) => {
      const foundItem = draft.find((item) => item.type === type && item.modelType === modelType && item.id === id);
      if (!foundItem) {
        draft.push({ ...downObj, status: IN_PROGRESS });
      } else {
        foundItem.status = IN_PROGRESS;
      }
    });

    // 调用API开始下载
    startDownLoad(paramsTemp, {
      onmessage: (parsedData: any) => {
        const { completedsize, progress, status, totalsize, error } = parsedData;

        // 处理错误情况
        if (error) {
          updateDownloadItem(setDownList, id, modelType, {
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
          updateDownloadItem(setDownList, id, modelType, {
            ...baseUpdates,
            currentDownload: 100,
            status: COMPLETED,
            completedsize: totalsize,
            totalsize,
          });
          downLoadCompleteEventBus.emit('downloadComplete');
        } else if (status === 'canceled') {
          updateDownloadItem(setDownList, id, modelType, {
            ...baseUpdates,
            status: PAUSED,
          });
        } else if (status === 'error') {
          updateDownloadItem(setDownList, id, modelType, {
            ...baseUpdates,
            status: FAILED,
          });
        } else {
          updateDownloadItem(setDownList, id, modelType, {
            ...baseUpdates,
            status: IN_PROGRESS,
          });
        }
      },
      onerror: (error: any) => {
        updateDownloadItem(setDownList, id, modelType, {
          status: FAILED,
        });
      },
    } as any);
  }, []);

  // 暂停下载
  const downLoadAbort = useCallback(
    async (data: any, { id, modelType }: any) => {
      try {
        await abortDownload(data);
        updateDownloadItem(setDownList, id, modelType, { status: PAUSED });
      } catch (error) {
        console.error('Failed to abort download:', error);
        // 增加错误处理，即使失败也尝试更新UI状态
        updateDownloadItem(setDownList, id, modelType, { status: FAILED });
      }
    },
    [setDownList],
  );

  // 刷新页面时从本地存储中获取下载列表
  useEffect(() => {
    const timeout = setTimeout(() => {
      const downListLocal = getLocalStorageDownList(LOCAL_STORAGE_KEYS.DOWN_LIST);
      if (downListLocal.length > 0) setDownList(downListLocal);

      // 将刷新之前正在下载的列表继续下载
      const refreshDownListLocal = getLocalStorageDownList(LOCAL_STORAGE_KEYS.REFRESH_DOWN_LIST);
      if (refreshDownListLocal.length > 0) {
        refreshDownListLocal.forEach((item: any) => {
          downLoadStart(
            { modelName: item.name },
            {
              ...item,
              type: item.type,
              status: IN_PROGRESS,
              modelType: item.modelType,
            },
          );
        });
      }
      localStorage.removeItem(LOCAL_STORAGE_KEYS.REFRESH_DOWN_LIST);
    }, 150);

    return () => clearTimeout(timeout);
  }, [downLoadStart]);

  // 监听浏览器刷新 暂停所有模型下载 并且缓存下载列表
  usePageRefreshListener(() => {
    // 存储正在下载的列表
    localStorage.setItem(LOCAL_STORAGE_KEYS.REFRESH_DOWN_LIST, JSON.stringify(downloadingItems));

    // 更新所有下载中的项目状态为暂停
    const data = downList.map((item) => ({
      ...item,
      status: item.status === IN_PROGRESS ? PAUSED : item.status,
    }));
    localStorage.setItem(LOCAL_STORAGE_KEYS.DOWN_LIST, JSON.stringify(data));
  });

  // 单独监听下载完成事件
  useEffect(() => {
    // 监听下载完成事件，立即处理已完成的项
    const handleDownloadComplete = () => {
      setDownList((currentList) => {
        const hasCompletedItems = currentList.some((item) => item.status === COMPLETED);
        if (hasCompletedItems) {
          return currentList.filter((item) => item.status !== COMPLETED);
        }
        return currentList;
      });
    };

    // 订阅下载完成事件
    downLoadCompleteEventBus.on('downloadComplete', handleDownloadComplete);
    return () => {
      downLoadCompleteEventBus.off('downloadComplete', handleDownloadComplete);
    };
  }, [COMPLETED, setDownList]);

  const intervalRef = useRef<any>(null);
  // 监听下载列表，处理已完成的下载项，作为兜底处理
  useEffect(() => {
    if (intervalRef.current) return;
    if (downList.length === 0) return;

    // 创建定时器，定期检查并清理已完成的下载项
    intervalRef.current = setInterval(() => {
      setDownList((currentList) => {
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
  }, [downList, setDownList]);

  return { downList, setDownList, downLoadStart, downLoadAbort };
};
