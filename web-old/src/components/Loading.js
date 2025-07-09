import React from 'react';
import { Spin } from '@douyinfe/semi-ui';

import { useTranslation } from 'react-i18next';

const Loading = ({ prompt: name = 'page' }) => {
  const { t } = useTranslation();
  return (
    <Spin style={{ height: 100 }} spinning={true}>
      {t("加载{name}中...",{name})}
    </Spin>
  );
};

export default Loading;
