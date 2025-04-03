import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enTranslation from './locales/en.json';
import zhTranslation from './locales/zh.json';

// 获取保存的语言设置或默认使用浏览器语言
const savedLanguage = localStorage.getItem('i18nextLng') || navigator.language.split('-')[0];

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: {
        translation: enTranslation
      },
      zh: {
        translation: zhTranslation
      }
    },
    lng: savedLanguage, // 使用保存的语言或浏览器语言
    fallbackLng: 'en', // 当检测到的语言不支持时，回退到英语
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'], // 缓存语言设置到 localStorage
      lookupLocalStorage: 'i18nextLng',
    },
    interpolation: {
      escapeValue: false
    }
  });

// 确保语言变更时保存到 localStorage
i18n.on('languageChanged', (lng) => {
  localStorage.setItem('i18nextLng', lng);
});

export default i18n;