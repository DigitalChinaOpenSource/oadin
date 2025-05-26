import { create } from 'zustand';

interface PageParams {
  fromDetail: boolean; // 是否从详情页跳转
  allParams?: any; // 用于存储跳转时的参数
}

const usePageParamsStore = create<{
  pageParams: PageParams;
  setPageParams: (params: Partial<PageParams>) => void;
  getPageParams: () => PageParams; // 新增方法
}>((set, get) => ({
  pageParams: {
    fromDetail: false,
    allParams: {},
  },
  setPageParams: (params: Partial<PageParams>) => {
    set((state) => ({
      pageParams: {
        ...state.pageParams,
        ...params,
      },
    }));
  },
  getPageParams: () => get().pageParams, // 新增方法实现
}));

export default usePageParamsStore;
