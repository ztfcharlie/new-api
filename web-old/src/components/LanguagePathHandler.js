import { useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { langs } from '../i18n/i18n';

const LanguagePathHandler = () => {
  const { i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  
  useEffect(() => {
    // 检查当前路径是否是语言路径
    const pathSegments = location.pathname.split('/').filter(Boolean);
    
    if (pathSegments.length > 0) {
      const firstSegment = pathSegments[0];
      
      // 检查第一段是否是支持的语言代码
      if (langs.includes(firstSegment)) {
        // 设置语言
        i18n.changeLanguage(firstSegment);
        localStorage.setItem('i18nextLng', firstSegment);
        
        // 重定向到没有语言代码的路径
        const newPath = '/' + pathSegments.slice(1).join('/');
        navigate(newPath, { replace: true });
      }
    }
  }, [location.pathname, i18n, navigate]);
  
  return null;
};

export default LanguagePathHandler;