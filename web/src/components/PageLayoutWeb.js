import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import React, { useContext, useEffect, useState } from 'react';
import { StyleContext } from '../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { API, getLogo, getSystemName, showError } from '../helpers/index.js';
import { setStatusData } from '../helpers/data.js';
import { UserContext } from '../context/User/index.js';
import { StatusContext } from '../context/Status/index.js';
import Loading from './Loading';
const { Header } = Layout;

import { LocaleProvider } from '@douyinfe/semi-ui';
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN';
import en_GB from '@douyinfe/semi-ui/lib/es/locale/source/en_GB';

const PageLayout = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const { i18n } = useTranslation();

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {
        statusDispatch({ type: 'set', payload: data });
        setStatusData(data);
      } else {
        showError('Unable to connect to server');
      }
    } catch (error) {
      showError('Failed to load status');
    }
  };

  useEffect(() => {
    loadUser();
    // 从localStorage获取上次使用的语言
    const savedLang = localStorage.getItem('i18nextLng');
    if (savedLang) {
      i18n.changeLanguage(savedLang);
    }
  }, [i18n]);

  
  return (
    <Layout style={{ 
      display: 'flex', 
      flexDirection: 'column',
      overflow: styleState.isMobile ? 'visible' : 'hidden'
    }}>
      <Header style={{ 
        padding: 0, 
        height: 'auto', 
        lineHeight: 'normal', 
        position: styleState.isMobile ? 'sticky' : 'fixed',
        width: '100%', 
        top: 0, 
        zIndex: 100,
        boxShadow: '0 1px 6px rgba(0, 0, 0, 0.08)'
      }}>
        <HeaderBar />
      </Header>
    </Layout>
  )
}

function PageLayoutContainer() {
  const { i18n } = useTranslation();
  const [page, setPage] = useState(true);
  const [_, userDispatch] = useContext(UserContext);
  return page ?
    <LocaleProvider locale={i18n.language == 'en' ? en_GB : zh_CN}>
      <PageLayout />
    </LocaleProvider>
    : <Loading />;
}


export default PageLayoutContainer;
