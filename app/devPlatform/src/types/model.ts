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
  is_recommended: boolean;
  status: number | string;
  avatar: string;
  can_select: boolean;
  class: string[];
  ollama_id: string;
  params_size: number;
  input_length?: number;
  output_length?: number;
  source?: IModelSourceType;
  type?: string;
  id: string;
  provider?: string;
  modelKey?: string;
  currentDownload?: number;
  is_downloaded?: boolean;
  update_time?: number;
  smartvision_provider?: string;
  smartvision_model_key?: string;
  completedsize?: number;
  totalsize?: number;
  think?: boolean; // 是否支持深度思考
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

export interface IMcpListItem {
  id: string | number;
  // 0 未添加未启用；1 已添加已启用
  status: number;
  envRequired: number;
  name: {
    src: string;
    zh: string;
  };
  abstract: {
    src: string;
    zh: string;
  };
  supplier: string;
  logo: string;
  popularity: number;
  tags: string[];
  hosted: boolean;
  authorized: number;
  updatedAt: number;
  envSchema?: Record<string, Record<string, any>>;
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

export interface IMcpListRequestParams {
  id?: number[];
  /** 搜索关键词，模糊匹配 title 和 abstract */
  keyword?: string;
  /** 类别筛选 */
  category?: string[];
  /** 标签筛选，取交集 */
  tags?: string[];
  /** 部署类型："local" / "hosted" */
  deployment?: 'local' | 'hosted';
  /** 页码，默认 1 */
  page?: number;
  /**  每页条数，默认 10 */
  size?: number;
}

export type cardType = {
  title: string;
  content: string;
  icon?: string;
  tags: string[];
  serviceId: string;
};

export interface ITagsDataItem {
  category: string;
  tags: Record<string, any>[];
}

export interface IPagination {
  current: number;
  pageSize: number;
  total: number;
}
