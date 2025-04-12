import React, { useState, useEffect, useRef, useContext } from 'react';
import { useTranslation } from 'react-i18next';
import { copy, showSuccess, API, showError } from '@/helpers';
import {
  renderQuota,
  getQuotaRatio,
  renderQuotaWithPrompt,
  getQuotaPerUnit,
} from '@/helpers/render';
import {
  Typography,
  Card,
  Button,
  Form,
  Divider,
  Descriptions,
  Modal,
  Input,
  InputNumber,
} from '@douyinfe/semi-ui';
import { IconCreditCard } from '@douyinfe/semi-icons';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { UserContext } from '@/context/User';
import styles from './AffLinkCard.module.scss';
export default (props) => {
  const { t } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const formRef = useRef();
  const [affCode, setAffCode] = useState('');
  const [quotaForInviter, setQuotaForInviter] = useState('');
  const [status, setStatus] = useState({});
  const [openTransfer, setOpenTransfer] = useState(false);
  const [transferAmount, setTransferAmount] = useState(0);

  useEffect(() => {
    getAffLink();
    getUserData();
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
    }
    setTransferAmount(getQuotaPerUnit());
    setQuotaForInviter(status.quota_for_inviter);
  }, []);
  const getUserData = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;

    if (success) {
      props.setUserQuota(data.quota);
      userDispatch({ type: 'login', payload: data });
    } else {
      showError(message);
    }
  };
  const getAffLink = async () => {
    const res = await API.get('/api/user/aff');
    const { success, message, data } = res.data;
    if (success) {
      let link = `${window.location.origin}/register?aff=${data}`;
      setAffCode(data);
      formRef.current.formApi.setValue('affLink', link);
    } else {
      showError(message);
    }
  };
  const handleAffLinkClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(t('邀请链接已复制到剪切板'));
  };

  const transfer = async () => {
    if (transferAmount < getQuotaPerUnit()) {
      showError(t('划转金额最低为') + ' ' + renderQuota(getQuotaPerUnit()));
      return;
    }
    const res = await API.post(`/api/user/aff_transfer`, {
      quota: transferAmount,
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(message);
      setOpenTransfer(false);
      getUserData().then();
    } else {
      showError(message);
    }
  };
  const handleCancel = () => {
    setOpenTransfer(false);
  };
  return (
    <>
      <Modal
        title={t('请输入要划转的数量')}
        visible={openTransfer}
        onOk={transfer}
        onCancel={handleCancel}
        maskClosable={false}
        size={'small'}
        centered={true}
      >
        <div style={{ marginTop: 20 }}>
          <Typography.Text>
            {t('可用额度')}
            {renderQuotaWithPrompt(userState?.user?.aff_quota)}
          </Typography.Text>
          <Input
            style={{ marginTop: 5 }}
            value={userState?.user?.aff_quota}
            disabled={true}
          ></Input>
        </div>
        <div style={{ marginTop: 20 }}>
          <Typography.Text>
            {t('划转额度')}
            {renderQuotaWithPrompt(transferAmount)}{' '}
            {t('最低') + renderQuota(getQuotaPerUnit())}
          </Typography.Text>
          <div>
            <InputNumber
              min={0}
              style={{ marginTop: 5 }}
              value={transferAmount}
              onChange={(value) => setTransferAmount(value)}
              disabled={false}
            ></InputNumber>
          </div>
        </div>
      </Modal>

      {userState?.user ? (
        <Card style={{ width: '500px', padding: '20px' }}>
          <div className='flex items-center justify-between'>
            <Title level={3} style={{ textAlign: 'center' }}>
              {t('邀请返利')}
            </Title>
            <div className='flex items-center'>
              <IconCreditCard />
              {t('返利比例')} {getQuotaRatio(quotaForInviter)}
            </div>
          </div>
          {quotaForInviter > 0 ? (
            <p className='my-2'>
              {t('成功邀请通过电子邮件验证注册的用户将获得')}
              <span style={{ color: 'rgba(var(--semi-red-5), 1)' }}>
                {renderQuota(quotaForInviter)}
              </span>
              {t('的代币奖励。')}
            </p>
          ) : null}

          <Descriptions className={styles['affLink-card']} align='left'>
            <Descriptions.Item itemKey={t('待使用收益')}>
              <span>{renderQuota(userState?.user?.aff_quota)}</span>
              <Button
                type={'secondary'}
                onClick={() => setOpenTransfer(true)}
                size={'small'}
                style={{ marginLeft: 10 }}
              >
                {t('划转')}
              </Button>
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('总收益')}>
              {renderQuota(userState?.user?.aff_history_quota)}
            </Descriptions.Item>
            <Descriptions.Item itemKey={t('邀请人数')}>
              {userState.user?.aff_count}
            </Descriptions.Item>
          </Descriptions>

          <div style={{ marginTop: 20 }}>
            <Divider>{t('邀请信息')}</Divider>
            <Form labelPosition='inset' ref={formRef}>
              <Typography.Text>
                {t('邀请码：')} {affCode}
              </Typography.Text>
              <Form.Input
                field='affLink'
                label={t('邀请链接')}
                placeholder={t('邀请链接')}
                onClick={handleAffLinkClick}
              />
            </Form>
          </div>
        </Card>
      ) : null}
    </>
  );
};
