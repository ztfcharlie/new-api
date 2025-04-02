import { t } from 'i18next';
import React from 'react';
import { Message } from 'semantic-ui-react';

const NotFound = () => (
  <>
    <Message negative>
      <Message.Header>{t('页面不存在')}</Message.Header>
      <p>{t('请检查你的浏览器地址是否正确')}</p>
    </Message>
  </>
);

export default NotFound;
