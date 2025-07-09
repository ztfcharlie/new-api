import { getUserIdFromLocalStorage, showError } from './utils';
import axios from 'axios';

export let API = axios.create({
  baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
    ? import.meta.env.VITE_REACT_APP_SERVER_URL
    : '',
  headers: {
    'New-API-User': getUserIdFromLocalStorage(),
    'Cache-Control': 'no-store',
  },
});

// 添加请求拦截器
API.interceptors.request.use(
  function (config) {
    const currentLang = localStorage.getItem('i18nextLng');
    // 确保语言值存在
    if (currentLang) {
      config.headers["I18n-Next-lng"] = currentLang;
      // 添加调试日志
      //console.log('Setting language header:', currentLang);
    } else {
      //console.log('No language found in localStorage');
    }
    return config;
  },
  function (error) {
    console.error('Request interceptor error:', error);
    return Promise.reject(error);
  }
);


API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  },
);

export function updateAPI() {
  API = axios.create({
    baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
      ? import.meta.env.VITE_REACT_APP_SERVER_URL
      : '',
    headers: {
      'New-API-User': getUserIdFromLocalStorage(),
      'Cache-Control': 'no-store',
    },
  });
  
  // 重新添加拦截器到新的 API 实例
  API.interceptors.request.use(
    function (config) {
      const currentLang = localStorage.getItem('i18nextLng');
      if (currentLang) {
        config.headers["I18n-Next-lng"] = currentLang;
        //console.log('Setting language header (after update):', currentLang);
      }
      return config;
    },
    function (error) {
      console.error('Request interceptor error:', error);
      return Promise.reject(error);
    }
  );

  API.interceptors.response.use(
    (response) => response,
    (error) => {
      showError(error);
    },
  );
}