package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

type RequestLog struct {
	Id        int    `json:"id" gorm:"primaryKey"`
	UserId    int    `json:"user_id" gorm:"index"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index"`
	Content   string `json:"content" gorm:"type:longtext"`
	Username  string `json:"username" gorm:"index;default:''"`
	TokenName string `json:"token_name" gorm:"index;default:''"`
	ModelName string `json:"model_name" gorm:"index;default:''"`
	ChannelId int    `json:"channel_id" gorm:"index"`
	TokenId   int    `json:"token_id" gorm:"index;default:0"`
	Group     string `json:"group" gorm:"index"`
	Ip        string `json:"ip" gorm:"index;default:''"`
}

func RecordRequestLog(c *gin.Context, userId int, channelId int, modelName string, tokenName string, content string, tokenId int, group string) {
	if !common.LogArtifactsEnabled {
		return
	}
	common.SysLog(fmt.Sprintf("RecordRequestLog: Preparing to insert log for user %d, content length: %d", userId, len(content)))
	
	username := c.GetString("username")
	if username == "" && userId != 0 {
		username, _ = GetUsernameById(userId, false)
	}

	// Determine if IP logging is needed
	needRecordIp := false
	if settingMap, err := GetUserSetting(userId, false); err == nil {
		if settingMap.RecordIpLog {
			needRecordIp = true
		}
	}
	
	ip := ""
	if needRecordIp && c != nil {
		ip = c.ClientIP()
	}

	logEntry := &RequestLog{
		UserId:    userId,
		CreatedAt: common.GetTimestamp(),
		Content:   content,
		Username:  username,
		TokenName: tokenName,
		ModelName: modelName,
		ChannelId: channelId,
		TokenId:   tokenId,
		Group:     group,
		Ip:        ip,
	}

	err := LOG_DB.Create(logEntry).Error
	if err != nil {
		logger.LogError(c, "failed to record request log: "+err.Error())
	}
}
