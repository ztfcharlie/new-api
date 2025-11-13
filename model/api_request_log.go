package model

import (
	"github.com/google/uuid"
)

type ApiRequestLog struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId      int    `json:"user_id" gorm:"index"`
	RequestId   string `json:"request_id" gorm:"type:varchar(255);uniqueIndex"`
	CreatedAt   int64  `json:"created_at" gorm:"index"`

	// 请求信息
	RequestMethod string `json:"request_method" gorm:"type:varchar(10)"`
	RequestPath   string `json:"request_path" gorm:"type:varchar(500)"`
	RequestHeaders string `json:"request_headers" gorm:"type:text"`
	RequestParams  string `json:"request_params" gorm:"type:text"`
	RequestBody    string `json:"request_body" gorm:"type:text"`
	ClientIP       string `json:"client_ip" gorm:"type:varchar(45);index"`

	// 响应信息
	ResponseStatus  int    `json:"response_status"`
	ResponseHeaders string `json:"response_headers" gorm:"type:text"`
	ResponseBody    string `json:"response_body" gorm:"type:text"`

	// 转发给大模型API的信息
	UpstreamRequestHeaders string `json:"upstream_request_headers" gorm:"type:text"`
	UpstreamRequestBody    string `json:"upstream_request_body" gorm:"type:text"`

	// 额外信息
	ModelName      string `json:"model_name" gorm:"type:varchar(100);index"`
	ChannelId      int    `json:"channel_id" gorm:"index"`
	TokenName      string `json:"token_name" gorm:"type:varchar(100)"`
	ProcessingTime int64  `json:"processing_time"`
	ErrorMessage   string `json:"error_message" gorm:"type:text"`
}

func GenerateRequestId() string {
	return uuid.New().String()
}