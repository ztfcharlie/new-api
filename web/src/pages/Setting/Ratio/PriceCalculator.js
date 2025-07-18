import React, { useState, useEffect } from 'react';
import { Input, Button, Form, Typography, Card, Space, Table, Divider, Tooltip } from '@douyinfe/semi-ui';
import { IconInfoCircle } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';

const { Text, Title, Paragraph } = Typography;

const PriceCalculator = ({ options }) => {
  const { t } = useTranslation();
  const [costRatio, setCostRatio] = useState(0.6); // 拿货价相对于国际参考价的比例
  const [taxRate, setTaxRate] = useState(0.06); // 税率默认6%
  const [referencePrice, setReferencePrice] = useState(1); // 国际参考价格
  const [calculatedPrices, setCalculatedPrices] = useState([]);
  const [profitRatios, setProfitRatios] = useState([]);
  const [memberMarkups, setMemberMarkups] = useState({
    "普通会员": 1.40,
    "银牌会员": 1.35,
    "金牌会员": 1.30,
    "铂金会员": 1.25,
    "钻石会员": 1.20,
    "保本": 1.00
  });

  const calculatePrices = () => {
    try {
      // 计算高成本阈值
      const HIGH_COST_THRESHOLD = 1 - taxRate;
      
      // 计算保本价格比例
      const breakEvenRatio = costRatio / (1 - taxRate);
      
      // 计算成本价
      const costPrice = referencePrice * costRatio;
      
      // 判断是否为高成本情况
      const isHighCost = costRatio >= HIGH_COST_THRESHOLD;
      
      const results = [];
      const memberLevels = ["保本", "普通会员", "银牌会员", "金牌会员", "铂金会员", "钻石会员"];
      
      memberLevels.forEach(level => {
        let finalRatio;
        let warning = '';
        
        if (isHighCost || level === "保本") {
          // 高成本情况或保本：使用保本价
          finalRatio = breakEvenRatio;
          if (level === "保本") {
            warning = '这是理论最低价格，仅覆盖成本和税费';
          }
        } else {
          // 正常情况：使用标准加价
          const costBasedRatio = (costPrice * memberMarkups[level]) / referencePrice;
          finalRatio = Math.min(1.0, costBasedRatio);
        }
        
        // 计算实际数据
        const sellingPrice = finalRatio * referencePrice;
        const tax = sellingPrice * taxRate;
        const profit = sellingPrice - tax - costPrice;
        const profitRate = profit / sellingPrice;
        
        // 计算相对于成本价的加价比例
        const markupOverCost = (sellingPrice / costPrice - 1) * 100;
        
        if (profitRate < 0) {
          warning = '警告：当前价格会亏损！';
        } else if (profitRate < 0.02 && !isHighCost) {
          warning = '警告：利润率较低！';
        }
        
        results.push({
          level,
          finalRatio: finalRatio.toFixed(3),
          finalRatioPercent: `${(finalRatio * 100).toFixed(1)}%`,
          profitRate: `${(profitRate * 100).toFixed(1)}%`,
          profitRateValue: profitRate,
          markupOverCost: `${markupOverCost.toFixed(1)}%`,
          warning,
          sellingPrice: sellingPrice.toFixed(4),
          profit: profit.toFixed(4),
          tax: tax.toFixed(4),
          costPrice: costPrice.toFixed(4)
        });
      });
      
      setCalculatedPrices(results);
      
      // 计算不同折扣率下的利润率变化
      const ratioData = [];
      // 从保本价格开始，以0.05为步长，直到1.0
      let startRatio = Math.ceil(breakEvenRatio * 20) / 20; // 向上取整到下一个0.05
      if (startRatio <= breakEvenRatio) startRatio += 0.05;
      
      for (let ratio = startRatio; ratio <= 1.0; ratio = parseFloat((ratio + 0.05).toFixed(2))) {
        const sellingPrice = ratio * referencePrice;
        const tax = sellingPrice * taxRate;
        const profit = sellingPrice - tax - costPrice;
        const profitRate = profit / sellingPrice;
        
        ratioData.push({
          ratio: ratio.toFixed(2),
          profitRate: (profitRate * 100).toFixed(1) + '%',
          profitValue: profit.toFixed(4),
          sellingPrice: sellingPrice.toFixed(4),
          tax: tax.toFixed(4)
        });
      }
      
      setProfitRatios(ratioData);
    } catch (e) {
      console.error("Error calculating prices:", e);
    }
  };

  const columns = [
    {
      title: t('会员等级'),
      dataIndex: 'level',
    },
    {
      title: t('售价比例'),
      dataIndex: 'finalRatio',
      render: (text, record) => `${text} (${record.finalRatioPercent})`,
    },
    {
      title: t('利润率'),
      dataIndex: 'profitRate',
    },
    {
      title: t('较成本加价'),
      dataIndex: 'markupOverCost',
    },
    {
      title: t('售价'),
      dataIndex: 'sellingPrice',
    },
    {
      title: t('备注'),
      dataIndex: 'warning',
      render: (text) => text ? <Text type={text.includes('警告') ? 'danger' : 'secondary'}>{text}</Text> : null,
    }
  ];

  const profitRatioColumns = [
    {
      title: t('折扣率'),
      dataIndex: 'ratio',
    },
    {
      title: t('利润率'),
      dataIndex: 'profitRate',
      render: (text) => {
        const profitRate = parseFloat(text);
        if (profitRate < 0) return <Text type="danger">{text}</Text>;
        if (profitRate < 2) return <Text type="warning">{text}</Text>;
        return <Text type="success">{text}</Text>;
      }
    },
    {
      title: t('售价'),
      dataIndex: 'sellingPrice',
    },
    {
      title: t('税费'),
      dataIndex: 'tax',
    },
    {
      title: t('利润'),
      dataIndex: 'profitValue',
      render: (text) => {
        const profit = parseFloat(text);
        if (profit < 0) return <Text type="danger">{text}</Text>;
        return <Text type="success">{text}</Text>;
      }
    }
  ];

  const handleMemberMarkupChange = (level, value) => {
    setMemberMarkups(prev => ({
      ...prev,
      [level]: parseFloat(value) || prev[level]
    }));
  };

  useEffect(() => {
    // 初始化时计算一次
    calculatePrices();
  }, []);

  return (
    <div style={{ padding: '20px 0' }}>
      <Card>
        <Text strong style={{ fontSize: '16px', marginBottom: '16px', display: 'block' }}>
          {t('会员价格计算器')}
        </Text>
        
        <Form layout="horizontal">
          <div style={{ display: 'flex', alignItems: 'flex-end', gap: '16px', flexWrap: 'wrap' }}>
            <Form.Input
              field="costRatio"
              label={
                <span>
                  {t('拿货价比例')}
                  <Tooltip content={t('拿货价相对于国际参考价的比例，例如0.6表示60%')}>
                    <IconInfoCircle style={{ marginLeft: '4px' }} />
                  </Tooltip>
                </span>
              }
              placeholder={t('例如：0.6')}
              value={costRatio}
              onChange={(value) => setCostRatio(parseFloat(value) || 0)}
              style={{ width: 120 }}
            />
            
            <Form.Input
              field="taxRate"
              label={t('税率')}
              placeholder={t('例如：0.06')}
              value={taxRate}
              onChange={(value) => setTaxRate(parseFloat(value) || 0)}
              style={{ width: 120 }}
            />
            
            <Form.Input
              field="referencePrice"
              label={t('国际参考价格')}
              placeholder={t('例如：1')}
              value={referencePrice}
              onChange={(value) => setReferencePrice(parseFloat(value) || 1)}
              style={{ width: 120 }}
            />
            
            <Button type="primary" onClick={calculatePrices}>
              {t('计算会员价格')}
            </Button>
          </div>
          
          <Divider margin="12px" />
          
          <div style={{ display: 'flex', width: '100%' }}>
            <Text strong style={{ fontSize: '15px', display: 'block' }}>{t('会员倍率设置')}</Text>
          </div>
          
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px', marginTop: '16px' }}>
            {Object.entries(memberMarkups).map(([level, value]) => (
              <div key={level} style={{ display: 'flex', flexDirection: 'column', width: 'auto' }}>
                <Text>{level}</Text>
                <Input
                  value={value}
                  onChange={(value) => handleMemberMarkupChange(level, value)}
                  style={{ width: 80 }}
                />
              </div>
            ))}
          </div>
        </Form>
      </Card>

      {calculatedPrices.length > 0 && (
        <Card style={{ marginTop: 16 }}>
          <div style={{ marginBottom: 16 }}>
            <Text strong>{t('计算结果')}</Text>
            {costRatio >= (1 - taxRate) && (
              <Text type="warning" style={{ display: 'block', marginTop: 8 }}>
                {t('【高成本保本策略】由于成本较高（≥参考价的')} {((1 - taxRate) * 100).toFixed(1)}%{t('），所有会员统一采用保本价格')}
              </Text>
            )}
            <Text style={{ display: 'block', marginTop: 4 }}>
              {t('保本价格比例：')} {(costRatio / (1 - taxRate)).toFixed(3)} {t('（覆盖成本和税费）')}
            </Text>
          </div>
          
          <Table
            columns={columns}
            dataSource={calculatedPrices}
            pagination={false}
            style={{ marginTop: 16 }}
          />
        </Card>
      )}

      {profitRatios.length > 0 && (
        <Card style={{ marginTop: 16 }}>
          <Text strong>{t('折扣率与利润率关系表')}</Text>
          <div style={{ marginTop: 16 }}>
            <Table
              columns={profitRatioColumns}
              dataSource={profitRatios}
              pagination={false}
              size="small"
            />
          </div>
        </Card>
      )}

      {/* 定价策略说明卡片 */}
      <Card style={{ marginTop: 24 }}>
        <Title heading={3} style={{ marginBottom: 16 }}>定价策略</Title>
        
        <Title heading={4} style={{ marginTop: 16 }}>C端用户</Title>
        <ul style={{ paddingLeft: 20, marginTop: 8 }}>
          <li>按当前的用户组等级给出价格计算</li>
          <li>如何升级到特定等级的用户组，暂未设计策略（可能按累计的充值金额自动升级 或者 管理员后台操作）</li>
        </ul>
        
        <Title heading={4} style={{ marginTop: 20 }}>B端用户</Title>
        
        <Paragraph style={{ marginTop: 8 }}>
          <Text strong>1. 单一模型定价</Text>
        </Paragraph>
        <ul style={{ paddingLeft: 20, marginTop: 4 }}>
          <li>确定模型所属类型组（official/sp）</li>
          <li>与用户沟通确定使用类型组</li>
          <li>通过特殊通道为用户组设置对应类型组的折扣</li>
          <li>不在特殊通道组的分组按正常分组倍率计费</li>
        </ul>
        
        <Paragraph style={{ marginTop: 12 }}>
          <Text strong>2. 整个通道定价（official/sp）</Text>
        </Paragraph>
        <ul style={{ paddingLeft: 20, marginTop: 4 }}>
          <li>先确认用户需求和模型所在组</li>
          <li>若正常倍率无法满足，使用特殊倍率设置</li>
          <li>通过调整模型所在组倍率实现用户折扣</li>
          <li>不在特殊通道组的分组按正常分组倍率计费</li>
        </ul>
        
        <Title heading={4} style={{ marginTop: 20 }}>商品定价基本原则</Title>
        
        <Paragraph style={{ marginTop: 8 }}>
          <Text strong>统一售价原则</Text>
        </Paragraph>
        <ul style={{ paddingLeft: 20, marginTop: 4 }}>
          <li>同一个商品，不同进货渠道，对用户展示统一售价</li>
          <li>避免因供应商差异造成混淆</li>
          <li>维持价格体系简单透明</li>
        </ul>
        
        <Paragraph style={{ marginTop: 12 }}>
          <Text strong>高价基准原则</Text>
        </Paragraph>
        <ul style={{ paddingLeft: 20, marginTop: 4 }}>
          <li>以较高拿货价为基准定价</li>
          <li>确保最不利情况下保持合理利润</li>
          <li>最终价格不高于国际参考价格+6%的税收，义务劳动了</li>
        </ul>
        
        <Title heading={5} style={{ marginTop: 16 }}>注意事项</Title>
        <ul style={{ paddingLeft: 20, marginTop: 8 }}>
          <li>设置特殊倍率时需考虑同组其他用户影响</li>
          <li>特殊倍率在日志中显示为"专属倍率"</li>
        </ul>
      </Card>
    </div>
  );
};

export default PriceCalculator;