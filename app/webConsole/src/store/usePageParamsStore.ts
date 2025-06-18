import { create } from 'zustand';
import { persist, PersistOptions } from 'zustand/middleware';
import { IMcpListRequestParams, IPagination, ITagsDataItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';

interface IPageParams {
  fromDetail: boolean;
  allParams: {
    pagination?: IPagination;
    checkedValues?: Record<string, any[]>;
    postParams?: IMcpListRequestParams;
    searchVal?: string;
  };
}

interface IPageParamsState {
  pageParams: IPageParams;
  tagsDataStore?: ITagsDataItem[];
  setTagsDataStore: (params: ITagsDataItem[]) => void;
  setPageParams: (params: Partial<IPageParams>) => void;
  getPageParams: () => IPageParams;
}

// 定义持久化配置的类型
type PageParamsPersistOptions = PersistOptions<IPageParamsState, unknown>;

const usePageParamsStore = create<IPageParamsState>()(
  persist(
    (set, get) => ({
      pageParams: {
        fromDetail: false,
        allParams: {},
      },
      tagsDataStore: [],
      setTagsDataStore: (params) => {
        set({ tagsDataStore: params || [] });
      },
      setPageParams: (params) => {
        set((state) => ({
          pageParams: {
            ...state.pageParams,
            ...params,
          },
        }));
      },
      getPageParams: () => get().pageParams,
    }),
    {
      name: 'page-params-store',
      partialize: (state) => ({
        pageParams: state.pageParams,
        tagsDataStore: state.tagsDataStore,
      }),
    } as PageParamsPersistOptions,
  ),
);

export default usePageParamsStore;
