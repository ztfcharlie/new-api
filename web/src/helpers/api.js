import { getUserIdFromLocalStorage, showError } from './utils';
import axios from 'axios';

export let API = axios.create({
  baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
    ? import.meta.env.VITE_REACT_APP_SERVER_URL
    : '',
  headers: {
    'New-API-User': getUserIdFromLocalStorage(),
    'Cache-Control': 'no-store'
  }
});

export function updateAPI() {
  API = axios.create({
    baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
      ? import.meta.env.VITE_REACT_APP_SERVER_URL
      : '',
    headers: {
      'New-API-User': getUserIdFromLocalStorage(),
      'Cache-Control': 'no-store'
    }
  });
}



// 添加请求拦截器
API.interceptors.request.use(
  function (config) {
    config.headers["I18n-Next-lng"] = localStorage.getItem('i18nextLng')
    return config;
  },
  function (error) {
    return Promise.reject(error);
  }
);


API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  },
);
