interface WxLoginOptions {
  id: string;
  appid: string;
  scope: string;
  redirect_uri: string;
  state: string;
  self_redirect?: boolean;
  styletype?: string;
  sizetype?: string;
  bgcolor?: string;
  rst?: string;
  style?: string;
  href?: string;
  lang?: string;
  stylelite?: number;
  fast_login?: number;
  onReady?: (ready: boolean) => void;
  onCleanup?: () => void;
}

interface Window {
  WxLogin: new (options: WxLoginOptions) => void;
}
