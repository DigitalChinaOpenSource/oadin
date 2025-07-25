export type IModelSourceType = 'local' | 'remote';

export interface IModelDataItem {
  service_name: string;
  api_flavor: string;
  flavor: string;
  method: string;
  desc: string;
  url: string;
  auth_type: string;
  auth_apply_url: string;
  auth_fields?: string[];
  name: string;
  service_provider_name: string;
  size: string;
  context: number;
  is_recommended: boolean;
  status: number | string;
  avatar: string;
  can_select: boolean;
  class: string[];
  ollama_id: string;
  params_size: number;
  source?: IModelSourceType;
  type?: string;
  id: string;
  provider?: string;
  modelKey?: string;
  currentDownload?: number;
  update_time?: number;
  smartvision_provider?: string;
  smartvision_model_key?: string;
  completedsize?: number;
  totalsize?: number;
  tools?: boolean; // 是否支持工具调用
  think?: boolean; // 是否支持深度思考
  think_switch?: boolean; // 深度思考是否支持开关
}
// 模型列表数据类型
export interface ModelData {
  data: IModelDataItem[];
  page: number;
  page_size: number;
  total: number;
  total_page: number;
}

export interface CredentialParam {
  id: number;
  name: string;
  label: string;
  type: string; // "text" | "password"
  placeholder: string;
  required: number; // 1 表示 true, 0 表示 false
  value: string | null;
  sort: number;
  createTime: number; // 时间戳
  updateTime: number; // 时间戳
}

export interface SmartvisionDataItem {
  id: number;
  name: string;
  avatar: string;
  type: number;
  provider: string;
  modelKey: string;
  introduce: string;
  tags: string[];
  credentialParamsId: string; // IDs 分割的字符串
  credentialParams: CredentialParam[];
}

// 问学列表数据类型
export interface ISmartvisionDataRes {
  data: SmartvisionDataItem[];
}

export interface IModelPathRes {
  path: string;
}

export interface IRequestModelParams {
  // 模型名称
  model_name: string;
  // 服务名称, 当前仅支持（chat/models/generate/embed/text-to-image）。
  service_name: string;
  // 服务来源, local-本地，remote-远程
  service_source: IModelSourceType;
  // 服务提供商
  provider_name?: string;
}

export type ModelSourceType = 'local' | 'remote';

export interface IModelSquareParams {
  flavor?: string;
  // remote时需要传
  // 'dev' | 'product'
  env_type?: string;
  service_source: ModelSourceType;
  page_size?: number;
  page?: number;
  mine?: boolean;
}
