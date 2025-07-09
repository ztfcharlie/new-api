import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar.js';
import App from '../../App.js';
import FooterBar from './Footer.js';
import { ToastContainer } from 'react-toastify';
import React, { useContext, useEffect, useState } from 'react';
import { StyleContext } from '../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { API, getLogo, getSystemName, showError,setStatusData } from '../helpers/index.js';
import { setStatusData } from '../helpers/data.js';
import { UserContext } from '../context/User/index.js';
import { StatusContext } from '../context/Status/index.js';
import Loading from './Loading';
import cookie from 'js-cookie';
const { Sider, Content, Header, Footer } = Layout;

import { LocaleProvider } from '@douyinfe/semi-ui';
import zh_CN from '@douyinfe/semi-ui/lib/es/locale/source/zh_CN';
import en_GB from '@douyinfe/semi-ui/lib/es/locale/source/en_GB';

const PageLayout = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const { state: styleState } = useStyle();
  const { i18n } = useTranslation();
  const location = useLocation();

  const shouldHideFooter = location.pathname === '/console/playground' || location.pathname.startsWith('/console/chat');

  const shouldInnerPadding = location.pathname.includes('/console') &&
    !location.pathname.startsWith('/console/chat') &&
    location.pathname !== '/console/playground';

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
    loadStatus().catch(console.error);
    let systemName = getSystemName();
    if (systemName) {
      //document.title = systemName;
    }
    let logo = getLogo();
    if (logo) {
      let linkElement = document.querySelector("link[rel~='icon']");
      if (linkElement) {
        linkElement.href = logo;
      }
    }
    // 从localStorage获取上次使用的语言
    const savedLang = localStorage.getItem('i18nextLng');
    if (savedLang) {
      i18n.changeLanguage(savedLang);
    }
  }, [i18n]);

  return (
    <Layout
      style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: styleState.isMobile ? 'visible' : 'hidden',
      }}
    >
      <Header
        style={{
          padding: 0,
          height: 'auto',
          lineHeight: 'normal',
          position: 'fixed',
          width: '100%',
          top: 0,
          zIndex: 100,
        }}
      >
        <HeaderBar />
      </Header>
      <Layout
        style={{
          overflow: styleState.isMobile ? 'visible' : 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {styleState.showSider && (
          <Sider
            style={{
              position: 'fixed',
              left: 0,
              top: '64px',
              zIndex: 99,
              border: 'none',
              paddingRight: '0',
              height: 'calc(100vh - 64px)',
            }}
          >
            <SiderBar />
          </Sider>
        )}
        <Layout
          style={{
            marginLeft: styleState.isMobile
              ? '0'
              : styleState.showSider
                ? styleState.siderCollapsed
                  ? '60px'
                  : '180px'
                : '0',
            transition: 'margin-left 0.3s ease',
            flex: '1 1 auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Content
            style={{
              flex: '1 0 auto',
              overflowY: styleState.isMobile ? 'visible' : 'hidden',
              WebkitOverflowScrolling: 'touch',
              padding: shouldInnerPadding ? (styleState.isMobile ? '5px' : '24px') : '0',
              position: 'relative',
            }}
          >
            <App />
          </Content>
          {!shouldHideFooter && (
            <Layout.Footer
              style={{
                flex: '0 0 auto',
                width: '100%',
              }}
            >
              <FooterBar />
            </Layout.Footer>
          )}
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  );
};

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
