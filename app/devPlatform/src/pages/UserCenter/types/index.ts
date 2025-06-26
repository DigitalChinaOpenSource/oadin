export type IUserType = 'person' | 'enterprise';

export interface IAccountInfo {
  id?: string;
  userName?: string;
  type: IUserType;
  avatarUrl?: string;
  email?: string;
  phone: string;
  enterpriseName?: string; // 企业用户特有字段
  isRealNameAuth?: boolean; // 是否实名认证
  isEnterpriseAuth?: boolean; // 是否企业认证
  wechatBind?: boolean; // 是否绑定微信
  wechatInfo?: Record<string, any>;
}

export interface IAccountInfoProps {
  setUserInfo: (userInfo: IAccountInfo) => void;
  accountInfo: IAccountInfo;
}
