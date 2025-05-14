import { ModelDataItem } from '@/types';
// 授权弹窗
export interface IModelAuthorize {
  apiHost: string;
  apiKey: string;
}

export type IModelAuthType = 'config' | 'update';

export interface IModelAuth {
  visible: boolean;
  type: IModelAuthType;
  modelData: ModelDataItem;
}
