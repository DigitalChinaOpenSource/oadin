import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import { create } from 'zustand/index';
import { persist } from 'zustand/middleware';

interface IMcpDownItem {
  mcpDetail: McpDetailType | null;
  error?: string | Record<string, any>;
  downStatus: 'downloading' | 'success' | 'error';
  startTime?: number; // 新增，记录downloading的开始时间
}

interface IDownStatusObj {
  id: string;
  downStatus: 'downloading' | 'success' | 'error';
  error?: string | Record<string, any>;
}

interface IMcpDownloadStore {
  mcpDownloadList: IMcpDownItem[];
  delMcpDownloadItem: (id: string | number) => void;
  mcpAddModalShow: boolean;
  setMcpAddModalShow: (show: boolean) => void;
  addMcpDownloadItemInit: (mcpItem: IMcpDownItem) => void;
  addMcpDownloadItem: (item: IDownStatusObj) => void;
}

const useMcpDownLoadStore = create<IMcpDownloadStore>()(
  persist(
    (set, get) => ({
      mcpDownloadList: [],
      mcpAddModalShow: false,
      addMcpDownloadItemInit: (mcpItem: IMcpDownItem) => {
        const currentList = get().mcpDownloadList;
        const currentItem = currentList.find((item) => item.mcpDetail?.id === mcpItem.mcpDetail?.id);
        if (!currentItem) {
          // 如果当前列表中没有该 mcpItem，则添加
          set({ mcpDownloadList: [...currentList, mcpItem] });
        } else {
          set({
            mcpDownloadList: currentList.map((item) => {
              return item.mcpDetail?.id === mcpItem.mcpDetail?.id ? { ...currentItem, downStatus: 'downloading' } : item;
            }),
          });
        }
      },
      addMcpDownloadItem: (downloadItem) => {
        const currentList = get().mcpDownloadList;
        const currentItem = currentList.find((item) => item.mcpDetail?.id === downloadItem.id);
        // 如果找到已有的正在下载的mcp，则修改当前的选项为对应状态，若是成功并在2s之后删除该项
        if (currentItem) {
          set({
            mcpDownloadList: currentList.map((item) => {
              return item.mcpDetail?.id === downloadItem.id ? { ...currentItem, ...downloadItem } : item;
            }),
          });
          if (downloadItem.downStatus === 'error') {
            set({ mcpAddModalShow: true });
          }
          if (downloadItem.downStatus === 'success') {
            setTimeout(() => {
              set({ mcpDownloadList: currentList.filter((item) => item.mcpDetail?.id !== downloadItem.id) });
            }, 2000);
          }
        }
      },
      delMcpDownloadItem: (id) => {
        const currentList = get().mcpDownloadList;
        set({ mcpDownloadList: currentList.filter((item) => item.mcpDetail?.id !== id) });
      },
      setMcpAddModalShow: (show) => {
        set({ mcpAddModalShow: show });
      },
    }),
    {
      name: 'mcp-download-store', // 本地存储的 key
      partialize: (state) => ({
        mcpDownloadList: state.mcpDownloadList,
        // 如需持久化 mcpDetail，可加上 mcpDetail: state.mcpDetail
      }),
    },
  ),
);

// 定时检测downloading超时（如5分钟=300000ms），超时则置为success
// const CHECK_INTERVAL = 30 * 1000; // 30秒检测一次
// const TIMEOUT = 3 * 60 * 1000; // 5分钟
//
// setInterval(() => {
//   const store = useMcpDownLoadStore.getState();
//   const now = Date.now();
//   let changed = false;
//   const newList = store.mcpDownloadList.map((item) => {
//     if (item.downStatus === 'downloading' && item.startTime && now - item.startTime > TIMEOUT) {
//       changed = true;
//       return { ...item, downStatus: 'success' };
//     }
//     return item;
//   });
//   if (changed) {
//     useMcpDownLoadStore.setState({ mcpDownloadList: newList as IMcpDownItem[] });
//   }
// }, CHECK_INTERVAL);

export default useMcpDownLoadStore;
