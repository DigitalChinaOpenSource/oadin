export type IModelSourceType = 'local' | 'remote';

export interface IModelDataItem {
  api_flavor: string;
  avatar: string;
  class: string[];
  description: string;
  flavor: string;
  id: string;
  name: string;
  ollama_id: string;
  params_size: number;
  service_source?: IModelSourceType;
  service_name: string;
  think?: boolean; // 是否支持深度思考
  think_switch?: boolean; // 是否支持深度思考
  size: string;
  updatedAt: number;
}
// 模型列表数据类型
export interface ModelData {
  list: IModelDataItem[];
  total: number;
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

// MCP列表数据类型
export interface McpData {
  list: IMcpListItem[];
  total: number;
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
export interface ITagsDataItem {
  category: string;
  tags: Record<string, any>[];
}
