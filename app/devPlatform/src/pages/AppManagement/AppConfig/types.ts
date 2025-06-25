import { IModelDataItem } from '@/types/model.ts';

export interface IModelSelectCardItem {
  name: string;
  avatar: string;
  class: string[];
  id: string;
}
export interface SearchParams {
  searchText?: string;
}

export type searchFunc = (params?: SearchParams) => Promise<IModelDataItem[]>;
