import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsPaymentGateway(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    PayAddress: '',
    EpayId: '',
    EpayKey: '',
    Price: 7.3,
    RmbPrice: 7.3,
    MinTopUp: 1,
    TopupGroupRatio: '',
    CustomCallbackAddress: '',
    PayMethods: '',
    // ... 现有的字段 ...
    StripeKey: '',
    StripeWebHookKey: '',
    // 添加 Coinbase 和 PayPal 的配置项
    CoinbaseKey: '',
    CoinbaseWebHookKey: '',
    PaypalKey: '',
    PaypalWebHookKey: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        PayAddress: props.options.PayAddress || '',
        EpayId: props.options.EpayId || '',
        EpayKey: props.options.EpayKey || '',
        Price: props.options.Price !== undefined ? parseFloat(props.options.Price) : 7.3,
        RmbPrice: props.options.RmbPrice !== undefined ? parseFloat(props.options.RmbPrice) : 7.3,
        MinTopUp: props.options.MinTopUp !== undefined ? parseFloat(props.options.MinTopUp) : 1,
        TopupGroupRatio: props.options.TopupGroupRatio || '',
        CustomCallbackAddress: props.options.CustomCallbackAddress || '',
        PayMethods: props.options.PayMethods || '',
        // ... 现有的字段 ...
        StripeKey: props.options.StripeKey || '',
        StripeWebHookKey: props.options.StripeWebHookKey || '',
        // 添加 Coinbase 和 PayPal 的配置项
        CoinbaseKey: props.options.CoinbaseKey || '',
        CoinbaseWebHookKey: props.options.CoinbaseWebHookKey || '',
        PaypalKey: props.options.PaypalKey || '',
        PaypalWebHookKey: props.options.PaypalWebHookKey || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitPayAddress = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
      if (!verifyJSON(inputs.TopupGroupRatio)) {
        showError(t('充值分组倍率不是合法的 JSON 字符串'));
        return;
      }
    }

    if (originInputs['PayMethods'] !== inputs.PayMethods) {
      if (!verifyJSON(inputs.PayMethods)) {
        showError(t('充值方式设置不是合法的 JSON 字符串'));
        return;
      }
    }

    setLoading(true);
    try {
      const options = [
        { key: 'PayAddress', value: removeTrailingSlash(inputs.PayAddress) },
      ];

      if (inputs.EpayId !== '') {
        options.push({ key: 'EpayId', value: inputs.EpayId });
      }
      if (inputs.EpayKey !== undefined && inputs.EpayKey !== '') {
        options.push({ key: 'EpayKey', value: inputs.EpayKey });
      }
      if (inputs.Price !== '') {
        options.push({ key: 'Price', value: inputs.Price.toString() });
      }
      if (inputs.RmbPrice !== undefined && inputs.RmbPrice !== '') {
        options.push({ key: 'RmbPrice', value: inputs.RmbPrice.toString() });
      }
      if (inputs.MinTopUp !== '') {
        options.push({ key: 'MinTopUp', value: inputs.MinTopUp.toString() });
      }
      if (inputs.CustomCallbackAddress !== '') {
        options.push({
          key: 'CustomCallbackAddress',
          value: inputs.CustomCallbackAddress,
        });
      }
      if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
        options.push({ key: 'TopupGroupRatio', value: inputs.TopupGroupRatio });
      }
      if (originInputs['PayMethods'] !== inputs.PayMethods) {
        options.push({ key: 'PayMethods', value: inputs.PayMethods });
      }

      if (inputs.StripeKey) {
        options.push({ key: 'StripeKey', value: inputs.StripeKey });
      }
      if (inputs.StripeWebHookKey !== undefined && inputs.StripeWebHookKey !== '') {
        options.push({ key: 'StripeWebHookKey', value: inputs.StripeWebHookKey });
      }
      if (inputs.CoinbaseKey) {
        options.push({ key: 'CoinbaseKey', value: inputs.CoinbaseKey });
      }
      if (inputs.CoinbaseWebHookKey !== undefined && inputs.CoinbaseWebHookKey !== '') {
        options.push({ key: 'CoinbaseWebHookKey', value: inputs.CoinbaseWebHookKey });
      }
      if (inputs.PaypalKey) {
        options.push({ key: 'PaypalKey', value: inputs.PaypalKey });
      }
      if (inputs.PaypalWebHookKey !== undefined && inputs.PaypalWebHookKey !== '') {
        options.push({ key: 'PaypalWebHookKey', value: inputs.PaypalWebHookKey });
      }



      // 发送请求
      const requestQueue = options.map(opt =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        })
      );

      const results = await Promise.all(requestQueue);

      // 检查所有请求是否成功
      const errorResults = results.filter(res => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach(res => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        // 更新本地存储的原始值
        setOriginInputs({ ...inputs });
        props.refresh && props.refresh();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('支付设置')}>
          <Text>
            {t('（当前仅支持易支付接口，默认使用上方服务器地址作为回调地址！）')}
          </Text>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='PayAddress'
                label={t('支付地址')}
                placeholder={t('例如：https://yourdomain.com')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='EpayId'
                label={t('易支付商户ID')}
                placeholder={t('例如：0001')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='EpayKey'
                label={t('易支付商户密钥')}
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CustomCallbackAddress'
                label={t('回调地址')}
                placeholder={t('例如：https://yourdomain.com')}
              />
            </Col>
            
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='Price'
                precision={2}
                label={t('充值价格（x元/美金）')}
                placeholder={t('例如：7，就是7元/美金')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='MinTopUp'
                label={t('最低充值美元数量')}
                placeholder={t('例如：2，就是最低充值2$')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.InputNumber
                field='RmbPrice'
                label='充值价格（x人民币/美金）'
                placeholder="例如：1，就是1人民币/美金"
              />
            </Col>
          </Row>

          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='StripeKey'
                label='StripeKey'
                placeholder=""
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='StripeWebHookKey'
                label='StripeWebHookKey'
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
          </Row>

          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CoinbaseKey'
                label='CoinbaseKey'
                placeholder=""
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='CoinbaseWebHookKey'
                label='CoinbaseWebHookKey'
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
          </Row>

          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='PaypalKey'
                label='PaypalKey'
                placeholder=""
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='PaypalWebHookKey'
                label='PaypalWebHookKey'
                placeholder={t('敏感信息不会发送到前端显示')}
                type='password'
              />
            </Col>
          </Row>
          


          <Form.TextArea
            field='TopupGroupRatio'
            label={t('充值分组倍率')}
            placeholder={t('为一个 JSON 文本，键为组名称，值为倍率')}
            autosize
          />
          <Form.TextArea
            field='PayMethods'
            label={t('充值方式设置')}
            placeholder={t('为一个 JSON 文本')}
            autosize
          />
          <Button onClick={submitPayAddress}>{t('更新支付设置')}</Button>
        </Form.Section>
      </Form>
    </Spin>
  );
} 