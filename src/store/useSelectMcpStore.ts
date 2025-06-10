import { create } from 'zustand';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';

interface ISelectMcpStore {
  drawerOpenId: string;
  setDrawerOpenId: (id: string) => void;
  selectMcpList: IMcpListItem[];
  setSelectMcpList: (list: IMcpListItem[] | ((currentList: IMcpListItem[]) => IMcpListItem[])) => void;
}

const useSelectMcpStore = create<ISelectMcpStore>((set) => ({
  selectMcpList: [],
  setSelectMcpList: (list: IMcpListItem[] | ((currentList: IMcpListItem[]) => IMcpListItem[])) => {
    if (typeof list === 'function') {
      // 函数式更新，传入当前状态获取新状态
      set((state) => ({ selectMcpList: list(state.selectMcpList) }));
    } else {
      // 直接设置新状态
      set({ selectMcpList: list });
    }
  },
  drawerOpenId: '',
  setDrawerOpenId: (id: string) => {
    if (id) {
      set({
        drawerOpenId: id,
      });
    } else {
      set({
        drawerOpenId: '',
      });
    }
  },
}));

export default useSelectMcpStore;
