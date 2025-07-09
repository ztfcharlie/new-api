import React, { useEffect, useState } from 'react';
import { Card, Spin } from '@douyinfe/semi-ui';

import { API, showError } from '../../helpers/index.js';
import { useTranslation } from 'react-i18next';
import RequestRateLimit from '../../pages/Setting/RateLimit/SettingsRequestRateLimit.js';

const RateLimitSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    ModelRequestRateLimitEnabled: false,
    ModelRequestRateLimitCount: 0,
    ModelRequestRateLimitSuccessCount: 1000,
    ModelRequestRateLimitDurationMinutes: 1,
    ModelRequestRateLimitGroup: '',
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key === 'ModelRequestRateLimitGroup') {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }

        if (item.key.endsWith('Enabled')) {
          newInputs[item.key] = item.value === 'true' ? true : false;
        } else {
          newInputs[item.key] = item.value;
        }
      });

      setInputs(newInputs);
    } else {
      showError(message);
    }
  };
  async function onRefresh() {
    try {
      setLoading(true);
      await getOptions();
      // showSuccess('刷新成功');
    } catch (error) {
      showError('刷新失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    onRefresh();
  }, []);

  return (
    <>
      <Spin spinning={loading} size='large'>
        {/* AI请求速率限制 */}
        <Card style={{ marginTop: '10px' }}>
          <RequestRateLimit options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default RateLimitSetting;
