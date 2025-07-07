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

import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
import { IModelDataItem } from '@/types';

// 新增：多标签页同步频道
const channel = new BroadcastChannel('downloadList-sync');

interface IModelDownloadStore {
  downloadList: IModelDataItem[];
  setDownloadList: (list: any[] | ((currentList: any[]) => any[])) => void;
  isDownloadEmbed: boolean;
  setIsDownloadEmbed: (isDownloadEmbed: boolean) => void;
}

const useModelDownloadStore = create(
  persist<IModelDownloadStore>(
    (set, get) => {
      // 监听其它标签页的同步消息
      channel.onmessage = (event) => {
        if (event.data?.type === 'SYNC_DOWNLOAD_LIST') {
          set({ downloadList: event.data.downloadList });
        }
      };
      return {
        downloadList: [],
        setDownloadList: (list: IModelDataItem[] | ((currentList: IModelDataItem[]) => IModelDataItem[])) => {
          let newList: IModelDataItem[];
          if (typeof list === 'function') {
            newList = list(get().downloadList);
          } else {
            newList = list;
          }

          // 对象数组去重
          const uniqueList = Array.from(new Map(newList.map((item) => [JSON.stringify(item), item])).values());
          set({ downloadList: uniqueList });
          // 广播到其它标签页
          channel.postMessage({ type: 'SYNC_DOWNLOAD_LIST', downloadList: uniqueList });
        },
        isDownloadEmbed: false,
        setIsDownloadEmbed: (isDownloadEmbed: boolean) => {
          set({ isDownloadEmbed });
        },
      };
    },
    {
      name: 'download-storage', // 存储中的项目名称，必须是唯一的
      // storage: createJSONStorage(() => sessionStorage), // 使用sessionStorage作为存储
    },
  ),
);

export default useModelDownloadStore;
