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
  // console.info(
  //   downList.filter((item: any) => !(item.canSelect && item.currentDownload === 100)),
  //   '过滤的下载列表',
  // );
  // 验证下载数量限制
  // const hasReachedLimit = downList.filter((item: any) => !(item.canSelect || item.currentDownload === 100)).length > 2;
  const hasReachedLimit = downList.filter((item: any) => item.status === 1).length > 2;
  // 触发限制条件
  if (hasReachedLimit) {
    message.warning('You have reached the maximum limit of downloading 3 models simultaneously. If you want to download a new model, please complete or cancel some existing download tasks first.');
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
  const { setDownloadList } = useModelDownloadStore.getState();
  const { setModelListData } = useModelListStore.getState();
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
