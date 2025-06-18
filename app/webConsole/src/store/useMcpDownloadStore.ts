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
  startDownloadStatusCheck: () => void; // 新增，启动定时检查函数
}

// 只保留必要数据
const filterData = (mcpItem: IMcpDownItem) => {
  if (!mcpItem.mcpDetail) return mcpItem;
  const { serverConfig, serverName, summary, ...rest } = mcpItem.mcpDetail;
  return { ...mcpItem, mcpDetail: { ...rest }, startTime: Date.now() };
};

const useMcpDownLoadStore = create<IMcpDownloadStore>()(
  persist(
    (set, get) => {
      let mcpInterval: NodeJS.Timeout | null = null; // 定时器单例
      return {
        mcpDownloadList: [],
        mcpAddModalShow: false,
        addMcpDownloadItemInit: (mcpItem: IMcpDownItem) => {
          const currentList = get().mcpDownloadList;
          const currentItem = currentList.find((item) => item.mcpDetail?.id === mcpItem.mcpDetail?.id);
          if (!currentItem) {
            // 如果当前列表中没有该 mcpItem，则添加，并记录开始时间
            set({ mcpDownloadList: [...currentList, filterData(mcpItem)] });
          } else {
            set({
              mcpDownloadList: currentList.map((item) => {
                return item.mcpDetail?.id === mcpItem.mcpDetail?.id ? { ...currentItem, downStatus: 'downloading', startTime: Date.now() } : item;
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
        startDownloadStatusCheck: () => {
          if (mcpInterval) return; // 如果定时器已存在，则直接返回
          mcpInterval = setInterval(() => {
            const currentList = get().mcpDownloadList;
            const updatedList = currentList.filter((item) => {
              if (item.downStatus === 'downloading' && item.startTime) {
                const elapsedTime = Date.now() - item.startTime;
                return elapsedTime <= 2 * 60 * 1000 + 5000; // 保留未超过二分钟的项目
              }
              return true; // 保留其他状态的项目
            });
            set({ mcpDownloadList: updatedList });
          }, 30 * 1000); // 每分钟检查一次
        },
      };
    },
    {
      name: 'mcp-download-store', // 本地存储的 key
      partialize: (state) => ({
        mcpDownloadList: state.mcpDownloadList,
      }),
    },
  ),
);

// 启动定时检查
useMcpDownLoadStore.getState().startDownloadStatusCheck();

export default useMcpDownLoadStore;
