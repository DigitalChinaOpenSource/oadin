import { IModelSourceType } from '@/types/model.ts';

export interface IModelSelectCardItem {
  name: string;
  avatar: string;
  class: string[];
  id: string;
}
export interface SearchParams {
  keyword?: string;
  page?: number;
  size?: number;
}

export type searchFunc = (params?: SearchParams) => Promise<ICardDeatilItem[]>;

export interface ICardDeatilItem {
  id: string;
  avatar: string;
  name: string;
  class: string[];
  desc?: string;
  source: IModelSourceType;
  update_time: number;
  flavor: string;
}
