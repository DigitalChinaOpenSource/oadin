import { IModelDataItem } from '@/types';
// 授权弹窗
export interface IModelAuthorize {
  apiHost: string;
  apiKey: string;
}

export type IModelAuthType = 'config' | 'update';

export interface IModelAuth {
  visible: boolean;
  type: IModelAuthType;
  modelData: IModelDataItem;
}

export interface IModelPathSpaceRes {
  free_size: number;
  total_size: number;
  usage_size: number;
}
