export type IUserType = 'person' | 'enterprise';

export interface IAccountInfo {
  id?: string;
  username?: string;
  nickname?: string;
  type: IUserType;
  avatar?: string;
  email?: string;
  phone: string;
  enterpriseName?: string; // 企业用户特有字段
  isRealNameVerified?: boolean; // 是否实名认证
  idCardFront?: string; // 身份证人像面照片
  idCardBack?: string; // 身份证国徽面照片
  isEnterpriseAuth?: boolean; // 是否企业认证
  wechatBind?: boolean; // 是否绑定微信
  wechatInfo?: Record<string, any>;
  wechatName?: string;
  realNameAuth?: Record<string, any>;
}

export interface IAccountInfoProps {
  setUserInfo: (userInfo: IAccountInfo) => void;
  accountInfo: IAccountInfo;
}

export interface IDeleteAccountProps {
  type?: IUserType;
  phone: string;
  smsCode: string;
  email?: string;
  emailCode?: string;
  enterpriseName?: string;
  token?: string | null;
  userId?: string;
  verifyCode?: string;
}
