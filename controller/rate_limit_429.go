package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// GetRateLimit429Stats 获取429统计记录
func GetRateLimit429Stats(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	stats, err := model.GetRecentRateLimit429Stats(pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get rate limit 429 stats: " + err.Error(),
		})
		return
	}

	// 获取总数
	var totalCount int64
	model.DB.Model(&model.RateLimit429Stat{}).Count(&totalCount)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     stats,
		"total":    totalCount,
		"page":     page,
		"page_size": pageSize,
	})
}

// GetRateLimit429Config 获取429监控配置
func GetRateLimit429Config(c *gin.Context) {
	config := service.GetRateLimit429Config()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateRateLimit429Config 更新429监控配置
func UpdateRateLimit429Config(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 更新各项配置
	if val, ok := req["enabled"]; ok {
		if enabled, ok := val.(bool); ok {
			if err := model.UpdateOption("RateLimit429MonitorEnabled", strconv.FormatBool(enabled)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to update enabled status: " + err.Error(),
				})
				return
			}
			// 同时更新内存中的OptionMap
			common.OptionMapRWMutex.Lock()
			common.OptionMap["RateLimit429MonitorEnabled"] = strconv.FormatBool(enabled)
			common.OptionMapRWMutex.Unlock()
		}
	}

	if val, ok := req["threshold"]; ok {
		if threshold, ok := val.(float64); ok {
			if err := model.UpdateOption("RateLimit429Threshold", strconv.Itoa(int(threshold))); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to update threshold: " + err.Error(),
				})
				return
			}
			// 同时更新内存中的OptionMap
			common.OptionMapRWMutex.Lock()
			common.OptionMap["RateLimit429Threshold"] = strconv.Itoa(int(threshold))
			common.OptionMapRWMutex.Unlock()
		}
	}

	if val, ok := req["email_recipients"]; ok {
		if recipients, ok := val.(string); ok {
			if err := model.UpdateOption("RateLimit429EmailRecipients", recipients); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to update email recipients: " + err.Error(),
				})
				return
			}
			// 同时更新内存中的OptionMap
			common.OptionMapRWMutex.Lock()
			common.OptionMap["RateLimit429EmailRecipients"] = recipients
			common.OptionMapRWMutex.Unlock()
		}
	}

	if val, ok := req["stat_duration"]; ok {
		if duration, ok := val.(float64); ok {
			if err := model.UpdateOption("RateLimit429StatDuration", strconv.Itoa(int(duration))); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to update stat duration: " + err.Error(),
				})
				return
			}
			// 同时更新内存中的OptionMap
			common.OptionMapRWMutex.Lock()
			common.OptionMap["RateLimit429StatDuration"] = strconv.Itoa(int(duration))
			common.OptionMapRWMutex.Unlock()
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// GetRateLimit429StatsByChannel 根据渠道ID获取429统计记录
func GetRateLimit429StatsByChannel(c *gin.Context) {
	channelId, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid channel ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	stats, err := model.GetRateLimit429StatsByChannel(channelId, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get channel stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetRateLimit429StatsByTimeRange 根据时间范围获取429统计记录
func GetRateLimit429StatsByTimeRange(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "start_time and end_time are required",
		})
		return
	}

	startTime, err := strconv.ParseInt(startTimeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid start_time format",
		})
		return
	}

	endTime, err := strconv.ParseInt(endTimeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid end_time format",
		})
		return
	}

	stats, err := model.GetRateLimit429StatsByTimeRange(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get time range stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetRateLimit429Summary 获取429统计摘要信息
func GetRateLimit429Summary(c *gin.Context) {
	// 获取今日统计
	todayStart := time.Now().Truncate(24 * time.Hour).Unix()
	todayStats, err := model.GetRateLimit429StatsByTimeRange(todayStart, time.Now().Unix())
	if err != nil {
		todayStats = []model.RateLimit429Stat{}
	}

	// 获取最近7天统计
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()
	weekStats, err := model.GetRateLimit429StatsByTimeRange(sevenDaysAgo, time.Now().Unix())
	if err != nil {
		weekStats = []model.RateLimit429Stat{}
	}

	// 统计总次数
	todayTotal := len(todayStats)
	weekTotal := len(weekStats)

	// 统计今日429错误总数
	today429Total := 0
	for _, stat := range todayStats {
		today429Total += stat.RateLimit429Count
	}

	// 统计最近7天429错误总数
	week429Total := 0
	for _, stat := range weekStats {
		week429Total += stat.RateLimit429Count
	}

	// 获取当前配置
	config := service.GetRateLimit429Config()

	summary := gin.H{
		"today": gin.H{
			"alert_count":   todayTotal,
			"error_count":   today429Total,
			"timestamp":     todayStart,
			"formatted_time": time.Unix(todayStart, 0).Format("2006-01-02"),
		},
		"week": gin.H{
			"alert_count": weekTotal,
			"error_count": week429Total,
			"start_time":  sevenDaysAgo,
			"end_time":    time.Now().Unix(),
		},
		"current_config": config,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// DeleteOldRateLimit429Stats 清理旧的429统计记录
func DeleteOldRateLimit429Stats(c *gin.Context) {
	err := model.DeleteOldRateLimit429Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete old stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Old rate limit 429 stats deleted successfully",
	})
}