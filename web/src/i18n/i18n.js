import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

import enTranslation from './locales/en.json';
import zhTranslation from './locales/zh.json';
import jaTranslation from './locales/ja.json';
import ruTranslation from './locales/ru.json';
import deTranslation from './locales/de.json';
import ptTranslation from './locales/pt.json';
import esTranslation from './locales/es.json';
import frTranslation from './locales/fr.json';
import koTranslation from './locales/ko.json';
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: {
        translation: enTranslation,
      },
      zh: {
        translation: zhTranslation,
      },
      ru: {
        translation: ruTranslation
      },
      de: {
        translation: deTranslation
      },
      pt: {
        translation: ptTranslation
      },
      es: {
        translation: esTranslation
      },
      fr: {
        translation: frTranslation
      },
      ja: {
        translation: jaTranslation
      },
      ko: {
        translation: koTranslation
      }
    },
    fallbackLng: 'zh',
    interpolation: {
      escapeValue: false,
    },
  });

// export const langs = ['en', 'zh', 'ru', 'de', 'pt', 'es', 'fr','ja', "ko"];
export const langs = [
  {
    label: 'English',
    value: 'en',
  },
  {
    label: '中文',
    value: 'zh',
  },
  {
    label: 'Русский',
    value: 'ru',
  },
  {
    label: 'Deutsch',
    value: 'de',
  },
  {
    label: 'Português',
    value: 'pt',
  },
  {
    label: 'Español',
    value: 'es',
  },
  {
    label: 'Français',
    value: 'fr',
  },
  {
    label: '日本語',
    value: 'ja',
  },
  {
    label: '한국어',
    value: 'ko',
  },
];




export default i18n;
