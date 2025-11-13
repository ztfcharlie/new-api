package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// RateLimit429Config 429监控配置
type RateLimit429Config struct {
	Enabled        bool   `json:"enabled"`
	Threshold      int    `json:"threshold"`       // 错误次数阈值
	EmailRecipients string `json:"email_recipients"` // 邮件收件人，用逗号分隔
	StatDuration   int    `json:"stat_duration"`   // 统计时长（分钟）
}

// RateLimit429Alert 429告警信息
type RateLimit429Alert struct {
	ChannelId   int    `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ModelName   string `json:"model_name"`
	TotalErrors int    `json:"total_errors"`
	RateLimitErrors int `json:"rate_limit_errors"`
	StatStartTime int64 `json:"stat_start_time"`
	StatEndTime   int64 `json:"stat_end_time"`
}

// GetRateLimit429Config 获取429监控配置
func GetRateLimit429Config() RateLimit429Config {
	config := RateLimit429Config{
		Enabled:        false,
		Threshold:      200,
		EmailRecipients: "burncloud@gmail.com,858377817@qq.com",
		StatDuration:   1,
	}

	// 从Option系统读取配置
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()

	if val, ok := common.OptionMap["RateLimit429MonitorEnabled"]; ok {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Enabled = enabled
		}
	}

	if val, ok := common.OptionMap["RateLimit429Threshold"]; ok {
		if threshold, err := strconv.Atoi(val); err == nil {
			config.Threshold = threshold
		}
	}

	if val, ok := common.OptionMap["RateLimit429EmailRecipients"]; ok {
		if val != "" {
			config.EmailRecipients = val
		}
	}

	if val, ok := common.OptionMap["RateLimit429StatDuration"]; ok {
		if duration, err := strconv.Atoi(val); err == nil {
			config.StatDuration = duration
		}
	}

	return config
}

// CheckRateLimit429Stats 检查429统计（主函数）
func CheckRateLimit429Stats() {
	config := GetRateLimit429Config()
	common.SysLog(fmt.Sprintf("429监控检查 - 启用状态: %v, 阈值: %d", config.Enabled, config.Threshold))

	// 如果监控未启用或阈值为0，则直接返回
	if !config.Enabled || config.Threshold <= 0 {
		common.SysLog("429监控未启用或阈值为0，跳过检查")
		return
	}

	now := time.Now()
	statStartTime := now.Add(-time.Duration(config.StatDuration) * time.Minute).Unix()
	common.SysLog(fmt.Sprintf("检查时间窗口: %s 到 %s (开始时间戳: %d)",
		time.Unix(statStartTime, 0).Format("15:04:05"), now.Format("15:04:05"), statStartTime))

	// 1. 统计最近N分钟内每个渠道和模型的总错误数（type=5）
	totalErrorStats, err := getTotalErrorStats(statStartTime)
	if err != nil {
		common.SysLog("Failed to get total error stats: " + err.Error())
		return
	}

	common.SysLog(fmt.Sprintf("找到 %d 个渠道-模型组合有错误记录", len(totalErrorStats)))

	// 2. 筛选超过阈值的记录，进一步统计429错误
	var alerts []RateLimit429Alert
	for channelId, modelName := range totalErrorStats {
		for model, totalErrors := range modelName {
			common.SysLog(fmt.Sprintf("检查渠道 %d - 模型 %s: 总错误数 %d, 阈值 %d", channelId, model, totalErrors, config.Threshold))
			if totalErrors >= config.Threshold {
				// 统计该渠道和模型的429错误数
				rateLimitErrors, err := getRateLimit429Errors(channelId, model, statStartTime)
				if err != nil {
					common.SysLog(fmt.Sprintf("Failed to get 429 errors for channel %d, model %s: %v", channelId, model, err))
					continue
				}

				common.SysLog(fmt.Sprintf("渠道 %d - 模型 %s: 429错误数 %d, 阈值 %d", channelId, model, rateLimitErrors, config.Threshold))
				// 检查429错误数是否达到阈值
				if rateLimitErrors >= config.Threshold {
					common.SysLog(fmt.Sprintf("✅ 触发告警条件，创建告警记录: 渠道 %d - 模型 %s", channelId, model))
					// 获取渠道名称
					channelName := getChannelName(channelId)

					alert := RateLimit429Alert{
						ChannelId:      channelId,
						ChannelName:    channelName,
						ModelName:      model,
						TotalErrors:    totalErrors,
						RateLimitErrors: rateLimitErrors,
						StatStartTime:  statStartTime,
						StatEndTime:    now.Unix(),
					}
					alerts = append(alerts, alert)
					common.SysLog(fmt.Sprintf("📊 已添加告警到列表，当前告警数量: %d", len(alerts)))
				} else {
					common.SysLog(fmt.Sprintf("❌ 未触发告警条件: 429错误数 %d < 1", rateLimitErrors))
				}
			}
		}
	}

	// 3. 保存统计信息并发送告警
	if len(alerts) > 0 {
		processAlerts(alerts, config)
	}
}

// getTotalErrorStats 统计总错误数（type=5），按渠道和模型分组
func getTotalErrorStats(startTime int64) (map[int]map[string]int, error) {
	var results []struct {
		ChannelId int    `json:"channel_id"`
		ModelName string `json:"model_name"`
		Count     int    `json:"count"`
	}

	err := model.LOG_DB.Table("logs").
		Select("channel_id, model_name, COUNT(*) as count").
		Where("type = 5 AND created_at >= ?", startTime).
		Group("channel_id, model_name").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[int]map[string]int)
	for _, result := range results {
		if _, ok := stats[result.ChannelId]; !ok {
			stats[result.ChannelId] = make(map[string]int)
		}
		stats[result.ChannelId][result.ModelName] = result.Count
	}

	return stats, nil
}

// getRateLimit429Errors 统计429错误数（type=5且content包含429）
func getRateLimit429Errors(channelId int, modelName string, startTime int64) (int, error) {
	var count int64
	err := model.LOG_DB.Model(&model.Log{}).
		Where("type = 5 AND channel_id = ? AND model_name = ? AND created_at >= ? AND content LIKE ?",
			channelId, modelName, startTime, "%429%").
		Count(&count).Error

	return int(count), err
}

// getChannelName 获取渠道名称
func getChannelName(channelId int) string {
	var channel model.Channel
	err := model.DB.First(&channel, channelId).Error
	if err != nil {
		return fmt.Sprintf("Channel_%d", channelId)
	}
	return channel.Name
}

// processAlerts 处理告警信息
func processAlerts(alerts []RateLimit429Alert, config RateLimit429Config) {
	var statsToSave []model.RateLimit429Stat

	for _, alert := range alerts {
		// 检查是否已经发送过邮件
		emailSent, err := model.CheckAndMarkEmailSent(alert.ChannelId, alert.ModelName, alert.StatStartTime)
		if err != nil {
			common.SysLog(fmt.Sprintf("Failed to check email status for channel %d, model %s: %v",
				alert.ChannelId, alert.ModelName, err))
			continue
		}

		// 构建错误详情
		errorDetails, _ := json.Marshal(alert)

		// 创建统计记录
		stat := model.RateLimit429Stat{
			ChannelId:          alert.ChannelId,
			ChannelName:        alert.ChannelName,
			ModelName:          alert.ModelName,
			StatStartTime:      alert.StatStartTime,
			CreatedAt:          time.Now().Unix(),
			TotalErrorCount:    alert.TotalErrors,
			RateLimit429Count:  alert.RateLimitErrors,
			EmailSent:          emailSent,
			StatDurationMinutes: config.StatDuration,
			ErrorDetails:       string(errorDetails),
		}

		statsToSave = append(statsToSave, stat)

		// 如果还没有发送邮件，则加入发送列表
		if !emailSent {
			// 将在后面统一发送
		}
	}

	// 保存统计信息
	if len(statsToSave) > 0 {
		for _, stat := range statsToSave {
			if err := model.AddRateLimit429Stat(&stat); err != nil {
				common.SysLog("Failed to save rate limit 429 stat: " + err.Error())
			}
		}
	}

	// 发送邮件告警（只发送未发送过的新告警）
	newAlerts := make([]RateLimit429Alert, 0)
	for _, alert := range alerts {
		emailSent, _ := model.CheckAndMarkEmailSent(alert.ChannelId, alert.ModelName, alert.StatStartTime)
		if !emailSent {
			newAlerts = append(newAlerts, alert)
		}
	}

	if len(newAlerts) > 0 {
		if err := sendRateLimit429AlertEmail(newAlerts, config); err != nil {
			common.SysLog("Failed to send rate limit 429 alert email: " + err.Error())
		}
	}
}

// sendRateLimit429AlertEmail 发送429告警邮件
func sendRateLimit429AlertEmail(alerts []RateLimit429Alert, config RateLimit429Config) error {
	if config.EmailRecipients == "" {
		return nil // 没有配置收件人，不发送邮件
	}

	// 构建邮件内容
	subject := "【429错误告警】检测到大量速率限制错误"

	emailBody := `
<html>
<head>
    <meta charset="UTF-8">
    <title>429错误告警</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .alert-box { border: 1px solid #ff6b6b; background-color: #fff5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .detail { margin: 10px 0; }
        .highlight { color: #d63031; font-weight: bold; }
        .footer { color: #666; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <h2>🚨 429错误告警通知</h2>
    <p>系统检测到以下渠道和模型出现了大量的速率限制错误（429），请及时处理：</p>
`

	for _, alert := range alerts {
		emailBody += fmt.Sprintf(`
    <div class="alert-box">
        <div class="detail"><strong>渠道ID:</strong> %d</div>
        <div class="detail"><strong>渠道名称:</strong> %s</div>
        <div class="detail"><strong>模型名称:</strong> %s</div>
        <div class="detail"><strong>总错误数:</strong> <span class="highlight">%d</span></div>
        <div class="detail"><strong>429错误数:</strong> <span class="highlight">%d</span></div>
        <div class="detail"><strong>统计时间:</strong> %s ~ %s</div>
    </div>
`, alert.ChannelId, alert.ChannelName, alert.ModelName,
   alert.TotalErrors, alert.RateLimitErrors,
   time.Unix(alert.StatStartTime, 0).Format("2006-01-02 15:04:05"),
   time.Unix(alert.StatEndTime, 0).Format("2006-01-02 15:04:05"))
	}

	emailBody += fmt.Sprintf(`
    <div class="footer">
        <p>配置的告警阈值: %d 错误/分钟</p>
        <p>统计时长: %d 分钟</p>
        <p>发送时间: %s</p>
        <p>此邮件由系统自动发送，请勿回复。</p>
    </div>
</body>
</html>
`, config.Threshold, config.StatDuration, time.Now().Format("2006-01-02 15:04:05"))

	// 解析收件人列表
	recipients := strings.Split(config.EmailRecipients, ",")
	for i, recipient := range recipients {
		recipients[i] = strings.TrimSpace(recipient)
	}

	// 发送邮件给所有收件人
	for _, recipient := range recipients {
		if recipient == "" {
			continue
		}
		if err := common.SendEmail(subject, recipient, emailBody); err != nil {
			common.SysLog(fmt.Sprintf("Failed to send email to %s: %v", recipient, err))
			// 不直接返回错误，尝试发送给其他收件人
		}
	}

	return nil
}

// StartRateLimit429Monitor 启动429监控定时任务
func StartRateLimit429Monitor() {
	common.SysLog("Starting rate limit 429 monitor...")

	go func() {
		ticker := time.NewTicker(1 * time.Minute) // 每分钟执行一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				func() {
					defer func() {
						if r := recover(); r != nil {
							common.SysLog("Rate limit 429 monitor panic: " + fmt.Sprintf("%v", r))
						}
					}()
					CheckRateLimit429Stats()
				}()
			}
		}
	}()
}