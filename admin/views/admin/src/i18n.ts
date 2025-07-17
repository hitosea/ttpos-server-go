import { createI18n } from 'vue-i18n';
//
import zh from './locales/zh.json';
import zhTW from './locales/zh-TW.json';
import en from './locales/en.json';
import th from './locales/th.json';
import ja from './locales/ja.json';
import ko from './locales/ko.json';
import my from './locales/my.json';
import tr from './locales/tr.json';
import sv from './locales/sv.json';
//
export const LANGUAGE = 'cashier_admin_language'.toUpperCase();

//
export const i18n = createI18n({
  globalInjection: true, // 注册全局$t
  legacy: false,
  fallbackWarn: false,
  missingWarn: false,
  locale: getLanguage(), // 获取已设置的语言
  fallbackLocale: 'en', //备用语言环境
  messages: {
    en,
    zh: zh,
    zhtw: zhTW,
    th, // 泰语
    ja, // 日语
    ko, // 韩语
    my, // 缅甸语
    tr, // 土耳其语
    sv, // 瑞典语
  },
});

// 语言类型
export enum LanguageEnum {
  en = 'English',
  zh = '简体中文',
  zhtw = '繁體中文',
  th = 'ภาษาไทย', // 泰语
  ja = '日本語', // 日语
  ko = '한국어', // 韩语
  my = 'မြန်မာဘာသာ', // 缅甸语
  tr = 'Türkçe', // 土耳其语
  sv = 'Svenska', // 瑞典语
}

// 语言类型
export type LocaleType = keyof typeof LanguageEnum;

// 获取语言列表
export const languageData = () => {
  const arr: { lang: LocaleType; text: string }[] = [];
  let key: LocaleType;
  for (key in LanguageEnum) {
    if (Object.prototype.hasOwnProperty.call(LanguageEnum, key)) {
      arr.push({
        lang: key,
        text: LanguageEnum[key],
      });
    }
  }
  //
  return arr;
};

// 获取语言
export function getLanguage(): string {
  return localStorage.getItem(LANGUAGE) || 'en';
}

// 设置语言
export function setLanguage(lang?: LocaleType) {
  if (!lang || i18n.global.locale.value == lang) return;
  // 设置语言
  i18n.global.locale.value = lang;
  // 设置语言
  localStorage.setItem(LANGUAGE, lang);
  // 重新加载页面
  location.reload();
}

//
export const $t = i18n.global.t;
