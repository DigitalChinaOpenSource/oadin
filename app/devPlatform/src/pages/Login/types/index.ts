import type { UploadFile } from 'antd';

export interface IBasePhoneFormProps {
  phone: string;
  verifyCode: string;
}

export interface IBaseEnterpriseFormProps {
  email: string;
  password: string;
}
export interface IEnterpriseCreateFormValues {
  enterpriseName: string;
  email: string;
  emailCode: string;
  password: string;
  surePassword: string;
  phone: string;
  smsCode: string;
  agreed: boolean;
}

export interface IImageUploadProps {
  title?: string;
  maxSize?: number; // 单位：MB
  accept?: string[];
  height?: number | string;
  value?: UploadFile[];
  onChange?: (fileList: UploadFile[]) => void;
  name?: string;
  rules?: any[];
  action?: string; // 添加上传地址
  bgIcon?: string;
}

export interface IWechatInfoProps {
  valid: boolean;
  user: Record<string, any>;
  token: string;
}

export interface IBaseRequestResProps {
  code: number;
  message: string;
  data: any;
}
