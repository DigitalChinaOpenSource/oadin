//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

import { useCallback, useRef, useEffect, useMemo } from 'react';
import { modelDownloadStream, abortDownload } from './download';
import { usePageRefreshListener, checkIsMaxDownloadCount, updateDownloadStatus } from './util';
import { DOWNLOAD_STATUS, LOCAL_STORAGE_KEYS } from '@/constants';
import { IModelDataItem } from '@/types';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelListStore from '@/store/useModelListStore';
import { getLocalStorageDownList } from '@/utils';
import { IDownParseData } from './types';
import { message } from 'antd';
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
  const { downloadList, setDownloadList, setIsDownloadEmbed } = useModelDownloadStore();
  const { FAILED, IN_PROGRESS, COMPLETED, PAUSED } = DOWNLOAD_STATUS;
  const downListRef = useRef<any[]>([]);
  downListRef.current = downloadList;

  // 计算当前下载中的项目
  const downloadingItems = useMemo(() => downloadList.filter((item) => item.status === IN_PROGRESS), [downloadList]);

  // 开始下载
  const fetchDownloadStart = useCallback(
    (params: IModelDataItem) => {
      const { id, source, service_provider_name, service_name, name, size } = params;

      // 最大下载数量
      const isMaxNum = checkIsMaxDownloadCount({
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
        size,
      };

      // 更新下载列表
      setDownloadList([
        ...downloadList.map((item) => (item.id === id ? { ...item, status: IN_PROGRESS } : item)),
        // 如果不存在则添加新项
        ...(!downloadList.some((item) => item.id === id) ? [{ ...params, status: IN_PROGRESS }] : []),
      ]);

      // 同步更新模型列表中的状态
      const { setModelListData } = useModelListStore.getState();
      setModelListData((draft) => draft.map((item) => (item.id === id ? { ...item, status: IN_PROGRESS, currentDownload: 0 } : item)));

      modelDownloadStream(paramsTemp, {
        onmessage: (parsedData: IDownParseData) => {
          const { completedsize, progress, status, totalsize, error, message: errorMsg } = parsedData;
          // console.log('下载状态更新:', parsedData);

          // 处理错误情况
          if (error) {
            updateDownloadStatus(id, {
              status: error.includes('aborted') ? PAUSED : FAILED,
            });
            return;
          }

          // 准备基础更新数据
          const baseUpdates = {
            ...(progress && { currentDownload: progress > 100 ? 94 : progress }),
            ...(completedsize && { completedsize }),
            ...(totalsize && { totalsize }),
          };

          // 根据状态更新下载项
          if (status === 'success') {
            updateDownloadStatus(id, {
              ...baseUpdates,
              currentDownload: 100,
              status: COMPLETED,
              completedsize: totalsize,
              totalsize,
              can_select: true,
            });

            // 处理特殊逻辑，词嵌入模型下载完成后设置状态
            if (params.name === 'quentinz/bge-large-zh-v1.5:f16') {
              setIsDownloadEmbed(true);
            }
            setTimeout(() => {
              setDownloadList((currentList) => currentList.filter((item) => item.status !== COMPLETED));
            }, 100);
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
            message.open({
              content: errorMsg,
              type: 'error',
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
        onclose: () => {
          // 处理连接关闭逻辑
          const completedItem = downListRef.current.find((item) => item.id === id);
          // console.log('连接关闭，检查下载状态', completedItem);

          if (completedItem && completedItem.status !== COMPLETED) {
            updateDownloadStatus(id, {
              status: FAILED,
            });
          }
        },
      });
    },
    [downloadList, setDownloadList],
  );

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
      const downListLocal = getLocalStorageDownList(LOCAL_STORAGE_KEYS.DOWN_LIST);
      if (downListLocal.length > 0) {
        // 将所有 IN_PROGRESS 状态的项目更新为 PAUSED
        const updatedList = downListLocal.map((item: IModelDataItem) => ({
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
      // console.log('2 秒执行定时器，处理已完成的下载项');
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
