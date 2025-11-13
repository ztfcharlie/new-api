/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Form,
  Row,
  Col,
  Typography,
  Modal,
  Banner,
  Spin,
  Card,
  Radio,
  Select,
  Input,
  Table,
  Tag,
  Space,
  Tooltip,
  Divider,
  Toast,
  InputNumber,
} from '@douyinfe/semi-ui';
const { Text, Title } = Typography;
import {
  API,
  showError,
  showSuccess,
  toBoolean,
} from '../../helpers';
import axios from 'axios';
import { useTranslation } from 'react-i18next';

const RateLimit429Setting = () => {
  const { t } = useTranslation();

  // 配置状态
  const [config, setConfig] = useState({
    enabled: false,
    threshold: 200,
    email_recipients: 'burncloud@gmail.com,858377817@qq.com',
    stat_duration: 1,
  });

  // 统计数据
  const [stats, setStats] = useState([]);
  const [summary, setSummary] = useState(null);
  const [loading, setLoading] = useState(false);
  const [statsLoading, setStatsLoading] = useState(false);

  // 分页状态
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
  });

  // 表单引用
  const formRef = useRef();
  const [formKey, setFormKey] = useState(0);

  // 初始化加载配置和统计数据
  useEffect(() => {
    loadConfig();
    loadSummary();
    loadStats();
  }, []);

  const loadConfig = async () => {
    try {
      const res = await API.get('/api/rate_limit_429/config');
      if (res.data.success) {
        setConfig(res.data.data);
        // 强制重新渲染表单以更新初始值
        setFormKey(prev => prev + 1);
      }
    } catch (error) {
      showError('加载配置失败: ' + error.message);
    }
  };

  const loadStats = async (page = 1, pageSize = 20) => {
    setStatsLoading(true);
    try {
      const res = await API.get(`/api/rate_limit_429/stats?page=${page}&page_size=${pageSize}`);
      if (res.data.success) {
        setStats(res.data.data);
        setPagination({
          page: res.data.page,
          pageSize: res.data.page_size,
          total: res.data.total,
        });
      }
    } catch (error) {
      showError('加载统计数据失败: ' + error.message);
    } finally {
      setStatsLoading(false);
    }
  };

  const loadSummary = async () => {
    try {
      const res = await API.get('/api/rate_limit_429/summary');
      if (res.data.success) {
        setSummary(res.data.data);
      }
    } catch (error) {
      showError('加载统计摘要失败: ' + error.message);
    }
  };

  const saveConfig = async (values) => {
    setLoading(true);
    try {
      const res = await API.put('/api/rate_limit_429/config', values);
      if (res.data.success) {
        showSuccess('配置保存成功');
        // 重新从服务器加载配置以确保状态同步
        await loadConfig();
      }
    } catch (error) {
      showError('保存配置失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const cleanupOldStats = async () => {
    try {
      const res = await API.delete('/api/rate_limit_429/cleanup');
      if (res.data.success) {
        showSuccess('旧统计数据清理成功');
        loadStats();
      }
    } catch (error) {
      showError('清理统计数据失败: ' + error.message);
    }
  };

  const handlePageChange = (page, pageSize) => {
    setPagination(prev => ({ ...prev, page, pageSize }));
    loadStats(page, pageSize);
  };

  // 表格列定义
  const columns = [
    {
      title: '渠道ID',
      dataIndex: 'channel_id',
      key: 'channel_id',
      width: 80,
    },
    {
      title: '渠道名称',
      dataIndex: 'channel_name',
      key: 'channel_name',
      width: 150,
    },
    {
      title: '模型名称',
      dataIndex: 'model_name',
      key: 'model_name',
      width: 200,
    },
    {
      title: '总错误数',
      dataIndex: 'total_error_count',
      key: 'total_error_count',
      width: 100,
      render: (text) => (
        <Tag color="orange" size="large">
          {text}
        </Tag>
      ),
    },
    {
      title: '429错误数',
      dataIndex: 'rate_limit_429_count',
      key: 'rate_limit_429_count',
      width: 100,
      render: (text) => (
        <Tag color="red" size="large">
          {text}
        </Tag>
      ),
    },
    {
      title: '统计时间',
      dataIndex: 'stat_start_time',
      key: 'stat_start_time',
      width: 150,
      render: (text) => new Date(text * 1000).toLocaleString(),
    },
    {
      title: '邮件已发送',
      dataIndex: 'email_sent',
      key: 'email_sent',
      width: 100,
      render: (text) => (
        <Tag color={text ? 'green' : 'yellow'} size="large">
          {text ? '是' : '否'}
        </Tag>
      ),
    },
  ];

  return (
    <div>
      {/* 统计摘要卡片 */}
      {summary && (
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col span={6}>
            <Card
              style={{
                background: 'white',
                border: summary.today.error_count > 0 ? '2px solid #ff4d4f' : '1px solid var(--semi-color-border)'
              }}
            >
              <div style={{ textAlign: 'center' }}>
                <Title heading={4}>今日告警</Title>
                <Text style={{ fontSize: 24, fontWeight: 'bold', color: '#ff4d4f' }}>
                  {summary.today.alert_count}
                </Text>
                <br />
                <Text type="secondary">
                  429错误: {summary.today.error_count}
                </Text>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card style={{ background: 'white' }}>
              <div style={{ textAlign: 'center' }}>
                <Title heading={4}>本周告警</Title>
                <Text style={{ fontSize: 24, fontWeight: 'bold' }}>
                  {summary.week.alert_count}
                </Text>
                <br />
                <Text type="secondary">
                  429错误: {summary.week.error_count}
                </Text>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card style={{ background: 'white' }}>
              <div style={{ textAlign: 'center' }}>
                <Title heading={4}>监控状态</Title>
                <Tag
                  color={summary.current_config.enabled ? 'green' : 'red'}
                  size="large"
                  style={{ fontSize: 16, padding: '4px 12px' }}
                >
                  {summary.current_config.enabled ? '已启用' : '已禁用'}
                </Tag>
                <br />
                <Text type="secondary">
                  阈值: {summary.current_config.threshold}
                </Text>
              </div>
            </Card>
          </Col>
          <Col span={6}>
            <Card style={{ background: 'white' }}>
              <div style={{ textAlign: 'center' }}>
                <Title heading={4}>统计时长</Title>
                <Text style={{ fontSize: 24, fontWeight: 'bold' }}>
                  {summary.current_config.stat_duration}
                </Text>
                <br />
                <Text type="secondary">
                  分钟
                </Text>
              </div>
            </Card>
          </Col>
        </Row>
      )}

      {/* 配置面板 */}
      <Card title="429错误监控配置" style={{ marginBottom: 24 }}>
        <Form
          key={formKey}
          ref={formRef}
          onSubmit={(values) => saveConfig(values)}
          labelPosition="top"
          initValues={config}
        >
          <Row gutter={16}>
            <Col span={8}>
              <Form.Switch
                field="enabled"
                label="启用监控"
                extraText="启用后将开始监控429错误"
              />
            </Col>
            <Col span={8}>
              <Form.InputNumber
                field="threshold"
                label="错误阈值"
                extraText="超过此数量将触发告警（0表示不限制）"
                min={0}
                max={10000}
              />
            </Col>
            <Col span={8}>
              <Form.InputNumber
                field="stat_duration"
                label="统计时长(分钟)"
                extraText="统计最近多少分钟内的错误"
                min={1}
                max={60}
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={24}>
              <Form.TextArea
                field="email_recipients"
                label="邮件收件人"
                extraText="多个收件人用逗号分隔"
                rows={2}
              />
            </Col>
          </Row>
          <Row style={{ marginTop: 16 }}>
            <Col span={24}>
              <Space>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                >
                  保存配置
                </Button>
                <Button
                  type="secondary"
                  onClick={loadConfig}
                >
                  重置
                </Button>
                <Button
                  type="warning"
                  onClick={cleanupOldStats}
                >
                  清理旧数据
                </Button>
                <Button
                  type="tertiary"
                  onClick={() => loadStats()}
                >
                  刷新数据
                </Button>
              </Space>
            </Col>
          </Row>
        </Form>
      </Card>

      {/* 统计数据表格 */}
      <Card
        title="429错误统计记录"
        extra={
          <Space>
            <Text type="secondary">
              最近 {pagination.total} 条记录
            </Text>
          </Space>
        }
      >
        <Table
          columns={columns}
          dataSource={stats}
          loading={statsLoading}
          pagination={{
            current: pagination.page,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            pageSizeOpts: [10, 20, 50, 100],
            onChange: handlePageChange,
          }}
          scroll={{ x: 1200 }}
        />
      </Card>
    </div>
  );
};

export default RateLimit429Setting;