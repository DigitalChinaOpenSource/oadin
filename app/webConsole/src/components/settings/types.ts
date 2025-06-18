export interface ISettings {
  ollamaRegistry: string; // 模型下载源地址
  systemProxy: ISystemProxy; // 系统模型
}

export interface ISystemProxy {
  enabled: boolean; // 是否开启代理
  endpoint?: string; // ip或域名
  username?: string; // 用户名
  password?: string; // 密码
}
