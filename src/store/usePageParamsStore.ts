import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface PageParams {
  fromDetail: boolean; // 是否从详情页跳转
  allParams?: any; // 用于存储跳转时的参数
}

const usePageParamsStore = create(
  persist(
    (set, get) => ({
      pageParams: {
        fromDetail: false,
        allParams: {},
      },
      // 服务标签
      tagsDataStore: [],
      setTagsDataStore: (params: any[]) => {
        set({ tagsDataStore: params || [] });
      },
      setPageParams: (params: Partial<PageParams>) => {
        set((state: any) => ({
          pageParams: {
            ...state.pageParams,
            ...params,
          },
        }));
      },
      getPageParams: () => get().pageParams, // 新增方法实现
    }),
    {
      name: 'page-params-store', // 存储在 localStorage 的 key
      partialize: (state: any) => ({
        pageParams: state.pageParams,
        tagsDataStore: state.tagsDataStore,
      }), // 仅持久化部分状态
    },
  ),
);

export default usePageParamsStore;
