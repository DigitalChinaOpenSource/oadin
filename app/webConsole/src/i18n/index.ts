import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import zhCN from './locales/zh-CN';
import enUS from './locales/en-US';

i18n.use(initReactI18next).init({
  resources: {
    zh: { translation: zhCN },
    en: { translation: enUS },
  },
  lng: 'zh',
  fallbackLng: 'zh',
  interpolation: {
    escapeValue: false,
  },
});

export interface I18nOptions {
  msg: string;
}
export const getMessageByModel = (key: string, options: I18nOptions) => {
  console.info(options);
  return i18n.t(`modelInfo.${key}`);
};

export default i18n;
