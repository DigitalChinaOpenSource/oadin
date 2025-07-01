import { IModelSelectCardItem } from '@/pages/AppManagement/AppConfig/types.ts';

export interface IApplicationParams {
  keyword?: string;
  page?: number;
  size?: number;
}

export interface IApplicationDetail {
  id: string;
  models: IModelSelectCardItem[];
  mcps: IModelSelectCardItem[];
  os: string[];
  acknowledged: boolean;
  name: string;
  appId: string;
  secretKey: string;
}
export interface IModelSelectCardItemRemote {
  name: {
    src: string;
    zh: string;
  };
  icon: string;
  tags: string[];
  id: string;
}
export interface IApplicationRemoteDetail {
  id: string;
  models: IModelSelectCardItemRemote[];
  mcps: IModelSelectCardItemRemote[];
  os: string[];
  acknowledged: boolean;
  name: string;
  appId: string;
  secretKey: string;
}

export interface ISaveApplicationConfigParams {
  id: string;
  models: string[];
  mcps: string[];
  os: string[];
  acknowledged: boolean;
}

export interface IPaginationParams {
  total: number;
  current?: number;
  pageSize?: number;
}
