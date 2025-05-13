// 授权弹窗
export interface IModelAuthorize {
  apiHost: string;
  apiKey: string;
}

export type IModelAuthType = 'config' | 'update';

export type IModelSourceType = 'local' | 'remote' | 'all';

export interface ModelDataItem {
  service_name: string;
  api_flavor: string;
  method: string;
  desc: string;
  url: string;
  auth_type: string;
  auth_apply_url: string;
  auth_fields: null | any;
  name: string;
  service_provider_name: string;
  size: string;
  is_recommended: boolean;
  status: string;
  avatar: string;
  can_select: boolean;
  class: string;
  ollama_id: string;
  params_size: number;
  source?: IModelSourceType;
  type?: string;
}
// 定义模型数据类型
export interface ModelData {
  chat: ModelDataItem[];
}
