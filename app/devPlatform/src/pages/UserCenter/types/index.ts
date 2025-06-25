export type IUserType = 'person' | 'enterprise';

export interface IAccountInfo {
  id?: string;
  userName?: string;
  userType: IUserType;
  avatarUrl?: string;
  email?: string;
  phoneNumber: string;
  companyName?: string; // 企业用户特有字段
  realName?: string; // 实名认证名称
  isRealNameAuth?: boolean; // 是否实名认证
  isEnterpriseAuth?: boolean; // 是否企业认证
  wechatBind?: boolean; // 是否绑定微信
  wechatInfo?: Record<string, any>;
}
