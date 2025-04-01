import React, { useEffect, useState, useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { getFooterHTML, getSystemName } from '../helpers';
import { Layout, Tooltip } from '@douyinfe/semi-ui';
import { StyleContext } from '../context/Style/index.js';

const FooterBar = () => {
  const { t } = useTranslation();
  const systemName = getSystemName();
  const [footer, setFooter] = useState(getFooterHTML());
  const [styleState] = useContext(StyleContext);
  let remainCheckTimes = 5;

  const loadFooter = () => {
    let footer_html = localStorage.getItem('footer_html');
    if (footer_html) {
      setFooter(footer_html);
    }
  };

  const defaultFooter = (
    <div className='custom-footer'>
      <a
        href='https://www.burncloud.com'
        target='_blank'
        rel='noreferrer'
      >
        Burncloud Inc.
      </a>
    </div>
  );

  useEffect(() => {
    const timer = setInterval(() => {
      if (remainCheckTimes <= 0) {
        clearInterval(timer);
        return;
      }
      remainCheckTimes--;
      loadFooter();
    }, 200);
    return () => clearTimeout(timer);
  }, []);

  return (
    <div style={{ textAlign: 'center' }}>
      {footer ? (
        <div
          className='custom-footer'
          dangerouslySetInnerHTML={{ __html: footer }}
        ></div>
      ) : (
        defaultFooter
      )}
    </div>
  );
};

export default FooterBar;
