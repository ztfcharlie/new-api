import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { useSetTheme, useTheme } from '../context/Theme';
import { useTranslation } from 'react-i18next';

import { API, getLogo, getSystemName, isMobile, showSuccess } from '../helpers';
import '../index.css';

import fireworks from 'react-fireworks';

import {
  IconClose,
  IconHelpCircle,
  IconHome,
  IconIndentLeft,
  IconComment,
  IconKey, 
  IconMenu,
  IconNoteMoneyStroked,
  IconPriceTag,
  IconUser,
  IconLanguage,
  IconInfoCircle,
  IconCreditCard,
  IconTerminal,
} from '@douyinfe/semi-icons';
import { Avatar, Button, Dropdown, Layout, Nav, Switch, Tag } from '@douyinfe/semi-ui';
import { stringToColor } from '../helpers/render';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { StyleContext } from '../context/Style/index.js';
import { StatusContext } from '../context/Status/index.js';

// 添加自定义仪表盘图标组件
const IconMenuStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24" // 保持viewBox为24x24以保持图标比例
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <line x1="4" x2="20" y1="12" y2="12"></line>
    <line x1="4" x2="20" y1="6" y2="6"></line>
    <line x1="4" x2="20" y1="18" y2="18"></line>
  </svg>
);

const DashboardIcon = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <rect width="7" height="9" x="3" y="3" rx="1"></rect>
    <rect width="7" height="5" x="14" y="3" rx="1"></rect>
    <rect width="7" height="9" x="14" y="12" rx="1"></rect>
    <rect width="7" height="5" x="3" y="16" rx="1"></rect>
  </svg>
);

const IconHomeStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"></path>
    <path d="M3 10a2 2 0 0 1 .709-1.528l7-5.999a2 2 0 0 1 2.582 0l7 5.999A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
  </svg>
);

const IconPriceTagStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <path d="M12.586 2.586A2 2 0 0 0 11.172 2H4a2 2 0 0 0-2 2v7.172a2 2 0 0 0 .586 1.414l8.704 8.704a2.426 2.426 0 0 0 3.42 0l6.58-6.58a2.426 2.426 0 0 0 0-3.42z"></path>
    <circle cx="7.5" cy="7.5" r=".5" fill="currentColor"></circle>
  </svg>
);

const IconHelpCircleStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <circle cx="12" cy="12" r="10"></circle>
    <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
    <path d="M12 17h.01"></path>
  </svg>
);

const IconInfoCircleStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <circle cx="12" cy="12" r="10"></circle>
    <path d="M12 16v-4"></path>
    <path d="M12 8h.01"></path>
  </svg>
);

const IconUserStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"></path>
    <circle cx="12" cy="7" r="4"></circle>
  </svg>
);

const IconKeyStroked = ({ style }) => (
  <svg 
    xmlns="http://www.w3.org/2000/svg" 
    width="18" 
    height="18" 
    viewBox="0 0 24 24"
    fill="none" 
    stroke="currentColor" 
    strokeWidth="1" 
    strokeLinecap="round" 
    strokeLinejoin="round"
    style={style}
  >
    <path d="M16.555 3.843l3.602 3.602a2.877 2.877 0 0 1 0 4.069l-2.643 2.643a2.877 2.877 0 0 1-4.069 0l-.301-.301l-6.558 6.558a2 2 0 0 1-1.239.578l-.175.008h-1.172a1 1 0 0 1-.993-.883l-.007-.117v-1.172a2 2 0 0 1 .467-1.284l.119-.13l.414-.414h2v-2h2v-2l2.144-2.144l-.301-.301a2.877 2.877 0 0 1 0-4.069l2.643-2.643a2.877 2.877 0 0 1 4.069 0z"/>
  </svg>
);


// 自定义顶部栏样式
const headerStyle = {
  boxShadow: '0 2px 10px rgba(0, 0, 0, 0.1)',
  borderBottom: '1px solid var(--semi-color-border)',
  background: 'var(--semi-color-bg-0)',
  transition: 'all 0.3s ease',
  width: '100%'
};

// 自定义顶部栏按钮样式
const headerItemStyle = {
  borderRadius: '4px',
  margin: '0 4px',
  transition: 'all 0.3s ease'
};

// 自定义顶部栏按钮悬停样式
const headerItemHoverStyle = {
  backgroundColor: 'var(--semi-color-primary-light-default)',
  color: 'var(--semi-color-primary)'
};

// 自定义顶部栏Logo样式
const logoStyle = {
  display: 'flex',
  alignItems: 'center',
  gap: '10px',
  padding: '0 10px',
  height: '100%'
};

// 自定义顶部栏系统名称样式
const systemNameStyle = {
  fontWeight: 'bold',
  fontSize: '18px',
  background: 'linear-gradient(45deg, #f97316, #dc2626)',
  WebkitBackgroundClip: 'text',
  WebkitTextFillColor: 'transparent',
  padding: '0 5px'
};

// 自定义顶部栏按钮图标样式
const headerIconStyle = {
  fontSize: '18px',
  transition: 'all 0.3s ease',
};

// 自定义头像样式
const avatarStyle = {
  margin: '4px',
  cursor: 'pointer',
  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
  transition: 'all 0.3s ease'
};

// 自定义下拉菜单样式
const dropdownStyle = {
  borderRadius: '8px',
  boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)',
  overflow: 'hidden'
};

// 自定义主题切换开关样式
const switchStyle = {
  margin: '0 8px'
};

const themeButtonStyle = {
  width: '36px',
  height: '36px',
  borderRadius: '50%',
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
  cursor: 'pointer',
  border: 'none',
  transition: 'all 0.3s ease',
  margin: '0 8px'
};

const HeaderBar = () => {
  const { t, i18n } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  let navigate = useNavigate();
  const savedLang = localStorage.getItem('i18nextLng') || i18n.language;
  const [currentLang, setCurrentLang] = useState(savedLang);

  const systemName = getSystemName();
  const logo = getLogo();
  const currentDate = new Date();
  // enable fireworks on new year(1.1 and 2.9-2.24)
  const isNewYear =
    (currentDate.getMonth() === 0 && currentDate.getDate() === 1);

  // Check if self-use mode is enabled
  const isSelfUseMode = statusState?.status?.self_use_mode_enabled || false;
  const docsLink = statusState?.status?.docs_link || '';
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;
  const mobileButtons = [
    {
      text: null,
      itemKey: 'home',
      to: '/',
      icon: <IconHomeStroked style={headerIconStyle} />,
    },
    {
      text: null,
      itemKey: 'detail',
      to: '/',
      icon: <DashboardIcon  style={headerIconStyle} />,
    },
    {
      text: null,
      itemKey: 'more',
      icon: <IconMenuStroked style={headerIconStyle} />,
      items: [
        {
          text: t('定价'),
          itemKey: 'pricing',
          to: '/pricing',
          icon: <IconPriceTagStroked style={headerIconStyle} />,
        },
        // Only include the docs button if docsLink exists
    /*
    ...(docsLink ? [{
      text: t('文档'),
      itemKey: 'docs',
      isExternal: true,
      externalLink: docsLink,
      icon: <IconHelpCircle style={headerIconStyle} />,
    }] : []),
      */
        {
          text: t('FAQ'),
          itemKey: 'faq',
          to: '/faq',
          icon: <IconHelpCircleStroked style={headerIconStyle} />,
        },
        {
          text: t('关于'),
          itemKey: 'about',
          to: '/about',
          icon: <IconInfoCircleStroked style={headerIconStyle} />,
        },
      ]
    }
  ];
  
  let buttons = styleState.isMobile ? mobileButtons : [
    {
      text: t('首页'),
      itemKey: 'home',
      to: '/',
      icon: <IconHomeStroked style={headerIconStyle} />,
    },
    {
      text: t('控制台'),
      itemKey: 'detail',
      to: '/',
      icon: <DashboardIcon style={headerIconStyle} />,
    },
    {
      text: t('定价'),
      itemKey: 'pricing',
      to: '/pricing',
      icon: <IconPriceTagStroked style={headerIconStyle} />,
    },
    // Only include the docs button if docsLink exists
    /*
    ...(docsLink ? [{
      text: t('文档'),
      itemKey: 'docs',
      isExternal: true,
      externalLink: docsLink,
      icon: <IconHelpCircleStroked style={headerIconStyle} />,
    }] : []),
      */
    {
      text: t('FAQ'),
      itemKey: 'faq',
      to: '/faq',
      icon: <IconHelpCircleStroked style={headerIconStyle} />,
    },
    {
      text: t('关于'),
      itemKey: 'about',
      to: '/about',
      icon: <IconInfoCircleStroked style={headerIconStyle} />,
    },
  ];
  

  async function logout() {
    await API.get('/api/user/logout');
    showSuccess(t('注销成功!'));
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  }

  const handleNewYearClick = () => {
    fireworks.init('root', {});
    fireworks.start();
    setTimeout(() => {
      fireworks.stop();
      setTimeout(() => {
        window.location.reload();
      }, 10000);
    }, 3000);
  };

  const theme = useTheme();
  const setTheme = useSetTheme();

  useEffect(() => {
    if (theme === 'dark') {
      document.body.setAttribute('theme-mode', 'dark');
    } else {
      document.body.removeAttribute('theme-mode');
    }
    // 发送当前主题模式给子页面
    const iframe = document.querySelector('iframe');
    if (iframe) {
      iframe.contentWindow.postMessage({ themeMode: theme }, '*');
    }

    if (isNewYear) {
      console.log('Happy New Year!');
    }
  }, [theme]);

  useEffect(() => {
    const handleLanguageChanged = (lng) => {
      setCurrentLang(lng);
      const iframe = document.querySelector('iframe');
      if (iframe) {
        iframe.contentWindow.postMessage({ lang: lng }, '*');
      }
    };

    i18n.on('languageChanged', handleLanguageChanged);

    return () => {
      i18n.off('languageChanged', handleLanguageChanged);
    };
  }, [i18n]);

  const handleLanguageChange = (lang) => {
    i18n.changeLanguage(lang);
    localStorage.setItem('i18nextLng', lang); // 确保语言设置被保存
    setCurrentLang(lang);
  };

  return (
    <>
      <Layout>
        <div style={{ width: '100%' }}>
          <Nav
            className={'topnav'}
            mode={'horizontal'}
            style={headerStyle}
            itemStyle={headerItemStyle}
            hoverStyle={headerItemHoverStyle}
            renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
              const routerMap = {
                about: '/about',
                login: '/login',
                register: '/register',
                pricing: '/pricing',
                detail: '/detail',
                home: '/',
                chat: '/chat',
                faq: '/faq',  // 添加这一行
              };
              return (
                <div onClick={(e) => {
                  // 添加 FAQ 到不需要设置内边距的页面列表中
                  if (props.itemKey === 'home' || props.itemKey === 'about' || props.itemKey === 'faq' || props.itemKey === 'language') {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
                    styleDispatch({ type: 'SET_SIDER', payload: false });
                  } else {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
                    if (!styleState.isMobile) {
                      styleDispatch({ type: 'SET_SIDER', payload: true });
                    }
                  }
                }}>
                  {props.isExternal ? (
                    <a
                      className="header-bar-text"
                      style={{ textDecoration: 'none' }}
                      href={props.externalLink}
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      {itemElement}
                    </a>
                  ) : (
                    <Link
                      className="header-bar-text"
                      style={{ textDecoration: 'none' }}
                      to={routerMap[props.itemKey]}
                    >
                      {itemElement}
                    </Link>
                  )}
                </div>
              );
            }}
            selectedKeys={[]}
            // items={headerButtons}
            onSelect={(key) => {}}
            header={styleState.isMobile?{
              logo: (
                <div style={{ display: 'flex', alignItems: 'center', position: 'relative' }}>
                  {
                    !styleState.showSider ?
                      <Button icon={<IconMenuStroked />} theme="light" aria-label={t('展开侧边栏')} onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: true })
                      } />:
                      <Button icon={<IconIndentLeft />} theme="light" aria-label={t('闭侧边栏')} onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: false })
                      } />
                  }
                  {(isSelfUseMode || isDemoSiteMode) && (
                    <Tag 
                      color={isSelfUseMode ? 'purple' : 'blue'}
                      style={{ 
                        position: 'absolute',
                        top: '-8px',
                        right: '-15px',
                        fontSize: '0.7rem',
                        padding: '0 4px',
                        height: 'auto',
                        lineHeight: '1.2',
                        zIndex: 1,
                        pointerEvents: 'none'
                      }}
                    >
                      {isSelfUseMode ? t('自用模式') : t('演示站点')}
                    </Tag>
                  )}
                </div>
              ),
            }:{
              logo: (
                <div style={logoStyle}>
                  <img src={logo} alt='logo' style={{ height: '28px' }} />
                </div>
              ),
              text: (
                <div style={{ position: 'relative', display: 'inline-block' }}>
                  <span style={systemNameStyle}>{systemName}</span>
                  {(isSelfUseMode || isDemoSiteMode) && (
                    <Tag 
                      color={isSelfUseMode ? 'purple' : 'blue'}
                      style={{ 
                        position: 'absolute', 
                        top: '-10px', 
                        right: '-25px', 
                        fontSize: '0.7rem',
                        padding: '0 4px',
                        whiteSpace: 'nowrap',
                        zIndex: 1,
                        boxShadow: '0 0 3px rgba(255, 255, 255, 0.7)'
                      }}
                    >
                      {isSelfUseMode ? t('自用模式') : t('演示站点')}
                    </Tag>
                  )}
                </div>
              ),
            }}
            items={buttons}
            footer={
              <>
                {isNewYear && (
                  // happy new year
                  <Dropdown
                    position='bottomRight'
                    render={
                      <Dropdown.Menu style={dropdownStyle}>
                        <Dropdown.Item onClick={handleNewYearClick}>
                          Happy New Year!!!
                        </Dropdown.Item>
                      </Dropdown.Menu>
                    }
                  >
                    <Nav.Item itemKey={'new-year'} text={'🎉'} />
                  </Dropdown>
                )}
                {/* <Nav.Item itemKey={'about'} icon={<IconHelpCircle />} /> */}
                <>
                <button
    onClick={() => {
      const isDark = theme !== 'dark';  // 转换为布尔值
      setTheme(isDark);  // 传递布尔值给 setTheme
    }}
    style={{
      ...themeButtonStyle,
      backgroundColor: theme === 'dark' ? 'var(--semi-color-bg-1)' : 'rgb(243 244 246)',
      color: theme === 'dark' ? 'rgb(31 41 55)' : 'rgb(243 244 246)'
    }}
    type="button"
    role="switch"
    aria-checked={theme === 'dark'}
    autoComplete="off"
>
    {theme === 'dark' ? '☀️' : '🌙'}
</button>
                </>
                <Dropdown
                    position='bottomRight'
                    render={
                      <Dropdown.Menu style={dropdownStyle}>
                        <Dropdown.Item
                          onClick={() => handleLanguageChange('en')}
                          type={currentLang === 'en' ? 'primary' : 'tertiary'}
                        >
                          EN
                        </Dropdown.Item>
                        <Dropdown.Item
                          onClick={() => handleLanguageChange('zh')}
                          type={currentLang === 'zh' ? 'primary' : 'tertiary'}
                        >
                          ZH
                        </Dropdown.Item>
                      </Dropdown.Menu>
                    }
                  >
                    <Nav.Item
                      itemKey={'language'}
                      text={currentLang.toUpperCase()}
                    />
                  </Dropdown>
                {userState.user ? (
                  <>
                    <Dropdown
                      position='bottomRight'
                      render={
                        <Dropdown.Menu style={dropdownStyle}>
                          <Dropdown.Item onClick={logout}>{t('退出')}</Dropdown.Item>
                        </Dropdown.Menu>
                      }
                    >
                      <Avatar
                        size='small'
                        color={stringToColor(userState.user.username)}
                        style={avatarStyle}
                      >
                        {userState.user.username[0]}
                      </Avatar>
                      {styleState.isMobile?null:<Text style={{ marginLeft: '4px', fontWeight: '500' }}>{userState.user.username}</Text>}
                    </Dropdown>
                  </>
                ) : (
                  <>
                    <Nav.Item
                      itemKey={'login'}
                      text={!styleState.isMobile?t('登录'):null}
                      icon={<IconUserStroked style={headerIconStyle} />}
                    />
                    {
                      // Hide register option in self-use mode
                      !styleState.isMobile && !isSelfUseMode && (
                        <Nav.Item
                          itemKey={'register'}
                          text={t('注册')}
                          icon={<IconKeyStroked style={headerIconStyle} />}
                        />
                      )
                    }
                  </>
                )}
              </>
            }
          ></Nav>
        </div>
      </Layout>
    </>
  );
};

export default HeaderBar;
