import React from 'react';
import { useTokenKeys } from '../../components/fetchTokenKeys';
import { useTranslation } from 'react-i18next';
const chat2page = () => {
  const { t } = useTranslation();
  const { keys, chatLink, serverAddress, isLoading } = useTokenKeys();

  const comLink = (key) => {
    if (!chatLink || !serverAddress || !key) return '';
    return `${chatLink}/#/?settings={"key":"sk-${key}","url":"${encodeURIComponent(serverAddress)}"}`;
  };

  if (keys.length > 0) {
    const redirectLink = comLink(keys[0]);
    if (redirectLink) {
      window.location.href = redirectLink;
    }
  }

  return (
    <div>
        <h3>{t('正在加载，请稍候...')}</h3>
    </div>
  );
};

export default chat2page;