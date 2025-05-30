export interface IServiceProviderDataItem {
  /** 认证密钥（例如 API Key）*/
  auth_key: string;
  /**
   * 认证类型
   * - "none": 无认证
   * - "api_key": 使用 API Key
   */
  auth_type: 'none' | 'api_key';
  /** 创建时间（ISO 8601 格式）*/
  created_at: string;
  /** 描述信息 */
  desc: string;
  /** 提供商类型*/
  flavor: 'tencent';
  /** 模型信息（可为空） */
  models: null;
  /** 扩展属性（JSON 字符串或对象） */
  properties: string | Record<string, any>;
  /** 提供商名称 */
  provider_name: 'remote_tencent_text_to_image';
  /** 服务名称（如 text_to_image） */
  service_name: string;
  /** 服务来源（如 remote） */
  service_source: string;
  /**
   * 状态码
   * - 0: 启用
   * - 1: 禁用
   * - 其他: 自定义状态
   */
  status: number;
  /** 最后更新时间（ISO 8601 格式）*/
  updated_at: string;
}

export interface IProviderDetailParams {
  /** 服务提供者名称 */
  provider_name: string;
  page?: number;
  page_size?: number;
  // 仅在provider_name包含smartvision时需要
  env_type?: string;
}

export interface ISupportModel {
  avatar: string;
  name: string;
  params_size: number;
  class: string[];
  flavor: string;
  api_flavor: string;
  input_length: number;
  output_length: number;
  is_downloaded: boolean;
}

export interface IProviderDetailData {
  id: number;
  provider_name: string;
  service_name: string;
  service_source: 'local' | 'remote';
  desc: string;
  method: string;
  url: string;
  auth_type: string;
  auth_key: string;
  flavor: string;
  extra_headers: string; // 可以是字符串化的 JSON 对象
  extra_json_body: string; // 可以是字符串化的 JSON 对象
  properties: string; // 同样是字符串化的 JSON 对象
  status: number; // 假设 0 表示禁用，1 表示启用
  created_at: string; // ISO8601 时间格式
  updated_at: string; // ISO8601 时间格式
  support_model_list: ISupportModel[];
  page: number;
  page_size: number;
  total_count: number;
  total_page: number;
}
