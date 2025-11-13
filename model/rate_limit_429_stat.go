package model

import (
	"time"
)

// RateLimit429Stat 429错误统计表
type RateLimit429Stat struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ChannelId   int    `json:"channel_id" gorm:"column:channel_id;index:idx_channel_model_time,priority:1"` // 渠道ID
	ChannelName string `json:"channel_name" gorm:"column:channel_name;type:varchar(100)"`                     // 渠道名称
	ModelName   string `json:"model_name" gorm:"column:model_name;type:varchar(100);index:idx_channel_model_time,priority:2"` // 模型名称
	// 统计时间窗口的开始时间（最近1分钟的统计时间）
	StatStartTime int64 `json:"stat_start_time" gorm:"column:stat_start_time;bigint;index:idx_channel_model_time,priority:3"`
	// 记录创建时间
	CreatedAt int64 `json:"created_at" gorm:"column:created_at;bigint"`
	// 总错误数（type=5的所有错误）
	TotalErrorCount int `json:"total_error_count" gorm:"column:total_error_count;default:0"`
	// 429错误数（content包含429的错误）
	RateLimit429Count int `json:"rate_limit_429_count" gorm:"column:rate_limit_429_count;default:0"`
	// 是否已发送邮件通知
	EmailSent bool `json:"email_sent" gorm:"column:email_sent;default:false"`
	// 统计时长（分钟），默认1分钟
	StatDurationMinutes int `json:"stat_duration_minutes" gorm:"column:stat_duration_minutes;default:1"`
	// 错误详情（JSON格式，保存所有相关错误信息）
	ErrorDetails string `json:"error_details" gorm:"column:error_details;type:text"`
}

// TableName 指定表名
func (RateLimit429Stat) TableName() string {
	return "rate_limit_429_stats"
}

// CreateRateLimit429StatTable 创建429统计表
func CreateRateLimit429StatTable() error {
	return DB.AutoMigrate(&RateLimit429Stat{})
}

// AddRateLimit429Stat 添加429统计记录
func AddRateLimit429Stat(stat *RateLimit429Stat) error {
	return DB.Create(stat).Error
}

// GetRecentRateLimit429Stats 获取最近的429统计记录
func GetRecentRateLimit429Stats(limit int, offset int) ([]RateLimit429Stat, error) {
	var stats []RateLimit429Stat
	err := DB.Order("created_at DESC").Limit(limit).Offset(offset).Find(&stats).Error
	return stats, err
}

// GetRateLimit429StatsByChannel 根据渠道ID获取429统计记录
func GetRateLimit429StatsByChannel(channelId int, limit int) ([]RateLimit429Stat, error) {
	var stats []RateLimit429Stat
	err := DB.Where("channel_id = ?", channelId).Order("created_at DESC").Limit(limit).Find(&stats).Error
	return stats, err
}

// DeleteOldRateLimit429Stats 清理旧的429统计记录（保留最近7天）
func DeleteOldRateLimit429Stats() error {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
	return DB.Where("created_at < ?", sevenDaysAgo).Delete(&RateLimit429Stat{}).Error
}

// GetRateLimit429StatsByTimeRange 根据时间范围获取429统计记录
func GetRateLimit429StatsByTimeRange(startTime, endTime int64) ([]RateLimit429Stat, error) {
	var stats []RateLimit429Stat
	err := DB.Where("stat_start_time >= ? AND stat_start_time <= ?", startTime, endTime).
		Order("created_at DESC").Find(&stats).Error
	return stats, err
}

// GetChannel429StatsCount 获取渠道429统计总数
func GetChannel429StatsCount(channelId int) (int64, error) {
	var count int64
	err := DB.Model(&RateLimit429Stat{}).Where("channel_id = ?", channelId).Count(&count).Error
	return count, err
}

// CheckAndMarkEmailSent 检查并标记邮件已发送状态
func CheckAndMarkEmailSent(channelId int, modelName string, statStartTime int64) (bool, error) {
	var stat RateLimit429Stat
	err := DB.Where("channel_id = ? AND model_name = ? AND stat_start_time = ?",
		channelId, modelName, statStartTime).First(&stat).Error
	if err != nil {
		// 如果找不到记录，说明还没有发送过邮件
		return false, nil
	}

	if stat.EmailSent {
		return true, nil // 已经发送过邮件
	}

	// 标记为已发送
	return false, DB.Model(&stat).Update("email_sent", true).Error
}