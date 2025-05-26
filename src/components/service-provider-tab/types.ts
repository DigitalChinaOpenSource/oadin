export interface IServiceProviderDataItem {
  /**
   * 认证密钥（例如 API Key）
   */
  auth_key: string;

  /**
   * 认证类型
   * - "none": 无认证
   * - "api_key": 使用 API Key
   */
  auth_type: 'none' | 'api_key';

  /**
   * 创建时间（ISO 8601 格式）
   */
  created_at: string;

  /**
   * 描述信息
   */
  desc: string;

  /**
   * 提供商类型（这里是 tencent）
   */
  flavor: 'tencent';

  /**
   * 模型信息（可为空）
   */
  models: null;

  /**
   * 扩展属性（JSON 字符串或对象）
   */
  properties: string | Record<string, any>;

  /**
   * 提供商名称
   */
  provider_name: 'remote_tencent_text_to_image';

  /**
   * 服务名称（如 text_to_image）
   */
  service_name: string;

  /**
   * 服务来源（如 remote）
   */
  service_source: string;

  /**
   * 状态码
   * - 0: 启用
   * - 1: 禁用
   * - 其他: 自定义状态
   */
  status: number;

  /**
   * 最后更新时间（ISO 8601 格式）
   */
  updated_at: string;
}
