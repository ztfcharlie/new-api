package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/gin-gonic/gin"
)

type RequestLog struct {
	Id        int       `json:"id" gorm:"primaryKey"`
	UserId    int       `json:"user_id" gorm:"index"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	Content   string    `json:"content" gorm:"type:longtext"`
	Username  string `json:"username" gorm:"index;default:''"`
	TokenName string `json:"token_name" gorm:"index;default:''"`
	ModelName string `json:"model_name" gorm:"index;default:''"`
	ChannelId int    `json:"channel_id" gorm:"index"`
	TokenId        int    `json:"token_id" gorm:"index;default:0"`
	Group          string `json:"group" gorm:"index"`
	Ip             string `json:"ip" gorm:"index;default:''"`
	AzureRequestId string `json:"azure_request_id" gorm:"index;default:''"`
	RejectFlage    *int64 `json:"reject_flage" gorm:"default:null"`
}

func RecordRequestLog(c *gin.Context, userId int, channelId int, modelName string, tokenName string, content string, tokenId int, group string) {
	if !common.LogArtifactsEnabled {
		return
	}
	
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

	azureRequestId := ""
	var rejectFlage *int64

	// Check if it is Azure
	if c != nil && c.GetInt("channel_type") == constant.ChannelTypeAzure {
		// Get Request ID
		if val := c.GetString("apim-request-id"); val != "" {
			azureRequestId = val
		} else if val := c.GetString("x-ms-request-id"); val != "" {
			azureRequestId = val
		} else if val := c.GetString("x-ms-client-request-id"); val != "" {
			azureRequestId = val
		}

		// Get Reject Flag
		if val, exists := c.Get("reject_flage"); exists {
			if ts, ok := val.(int64); ok {
				rejectFlage = &ts
			}
		}
	}

	logEntry := &RequestLog{
		UserId:         userId,
		CreatedAt:      time.Now().UTC(),
		Content:        content,
		Username:       username,
		TokenName:      tokenName,
		ModelName:      modelName,
		ChannelId:      channelId,
		TokenId:        tokenId,
		Group:          group,
		Ip:             ip,
		AzureRequestId: azureRequestId,
		RejectFlage:    rejectFlage,
	}

	err := LOG_DB.Create(logEntry).Error
	if err != nil {
		logger.LogError(c, "failed to record request log: "+err.Error())
	}
}
