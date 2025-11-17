package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

const (
	// MySQL TEXT字段最大长度约为64KB，留一些buffer
	MaxTextFieldLength = 60000
	// 数据表最大记录数
	MaxLogRecords = 10000
)

// responseBodyWriter 用于捕获响应体
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// truncateString 截取字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[TRUNCATED]"
}

// safeJSONMarshal 安全的JSON序列化，避免panic
func safeJSONMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}

// isValidUTF8Text 检查数据是否为有效的UTF-8文本
func isValidUTF8Text(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// 检查是否包含gzip magic bytes (0x1F 0x8B)
	if len(data) >= 2 && data[0] == 0x1F && data[1] == 0x8B {
		return false
	}

	// 使用utf8.ValidString检查是否为有效的UTF-8
	return utf8.ValidString(string(data))
}

// containsBase64 检查字符串是否包含base64编码的内容
func containsBase64(content string) bool {
	// 移除空白字符后检查是否为有效的base64
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "\r", "")
	content = strings.ReplaceAll(content, "\t", "")

	if len(content) < 4 {
		return false
	}

	// 检查是否是base64格式：只包含base64字符且长度是4的倍数
	base64Pattern := `^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`
	matched, _ := regexp.MatchString(base64Pattern, content)

	if matched {
		// 额外检查：如果内容较长且完全是base64格式，很可能是base64编码
		if len(content) > 20 {
			return true
		}
	}

	return false
}

// processRequestBody 处理请求体，如果包含base64内容则替换为占位符
func processRequestBody(data []byte) string {
	content := string(data)

	// 首先检查是否为有效的UTF-8文本
	if !isValidUTF8Text(data) {
		return "it is not a text contents"
	}

	// 检查是否包含base64内容
	if containsBase64(content) {
		return "it contains base64 content"
	}

	return content
}

// processResponseBody 处理响应体，如果不是文本内容则替换为占位符
func processResponseBody(data []byte) string {
	if !isValidUTF8Text(data) {
		return "it is not a text contents"
	}
	return string(data)
}

// ApiRequestLog 记录API请求和响应的详细信息
func ApiRequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用API请求日志
		if !common.ApiRequestLogEnabled {
			c.Next()
			return
		}

		startTime := time.Now()
		requestId := model.GenerateRequestId()

	// 将request_id存入context，便于后续使用
		c.Set("request_id", requestId)

	// 创建一个用于存储上游请求信息的结构体，并放入context
	upstreamInfo := &common.UpstreamRequestInfo{}
	c.Set("upstream_request_info", upstreamInfo)

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 读取请求头并转换为JSON
		requestHeaders := make(map[string]string)
		for key, values := range c.Request.Header {
			// 过滤掉敏感的header信息
			if strings.ToLower(key) == "authorization" ||
			   strings.ToLower(key) == "cookie" ||
			   strings.ToLower(key) == "set-cookie" {
				continue
			}
			requestHeaders[key] = strings.Join(values, ", ")
		}
		requestHeadersJSON := safeJSONMarshal(requestHeaders)

		// 读取请求参数
		requestParams := make(map[string]interface{})
		for key, values := range c.Request.URL.Query() {
			if len(values) == 1 {
				requestParams[key] = values[0]
			} else {
				requestParams[key] = values
			}
		}
		requestParamsJSON := safeJSONMarshal(requestParams)

		// 创建响应体写入器
		responseBodyWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseBodyWriter

		// 执行请求
		c.Next()

		// 计算处理时间
		processingTime := time.Since(startTime).Milliseconds()

		// 读取响应头并转换为JSON
		responseHeaders := make(map[string]string)
		for key, values := range c.Writer.Header() {
			responseHeaders[key] = strings.Join(values, ", ")
		}
		responseHeadersJSON := safeJSONMarshal(responseHeaders)

		// 获取用户信息
		userIdInt := 0
		if userId, exists := c.Get("id"); exists {
			if uid, ok := userId.(int); ok {
				userIdInt = uid
			}
		}

		// 处理和截取长数据以防止数据库错误
		// 处理请求体，检查是否包含base64内容
		processedRequestBody := processRequestBody(requestBody)
		truncatedRequestBody := truncateString(processedRequestBody, MaxTextFieldLength)
		// 处理响应体，检查是否为文本内容
		processedResponseBody := processResponseBody(responseBodyWriter.body.Bytes())
		truncatedResponseBody := truncateString(processedResponseBody, MaxTextFieldLength)
		truncatedRequestHeaders := truncateString(requestHeadersJSON, MaxTextFieldLength)
		truncatedRequestParams := truncateString(requestParamsJSON, MaxTextFieldLength)
		truncatedResponseHeaders := truncateString(responseHeadersJSON, MaxTextFieldLength)

		// 创建日志记录
		apiLog := &model.ApiRequestLog{
			UserId:          userIdInt,
			RequestId:       requestId,
			CreatedAt:       common.GetTimestamp(),
			RequestMethod:   c.Request.Method,
			RequestPath:     c.Request.URL.Path,
			RequestHeaders:  truncatedRequestHeaders,
			RequestParams:   truncatedRequestParams,
			RequestBody:     truncatedRequestBody,
			ClientIP:        c.ClientIP(),
			ResponseStatus:  c.Writer.Status(),
			ResponseHeaders: truncatedResponseHeaders,
			ResponseBody:    truncatedResponseBody,
			ProcessingTime:  processingTime,
		}

		// 尝试从context中获取额外信息
		if modelName, exists := c.Get("model_name"); exists {
			apiLog.ModelName = modelName.(string)
		}
		if channelId, exists := c.Get("channel_id"); exists {
			apiLog.ChannelId = channelId.(int)
		}
		if tokenNameStr, exists := c.Get("token_name"); exists {
			apiLog.TokenName = tokenNameStr.(string)
		}

		// 从context中获取上游请求信息并添加到日志
		if upstreamInfo, exists := c.Get("upstream_request_info"); exists {
			if info, ok := upstreamInfo.(*common.UpstreamRequestInfo); ok {
				apiLog.UpstreamRequestHeaders = truncateString(info.Headers, MaxTextFieldLength)
				apiLog.UpstreamRequestBody = truncateString(info.Body, MaxTextFieldLength)
			}
		}

		// 如果有错误信息，记录错误
		if len(c.Errors) > 0 {
			errorMessages := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMessages[i] = err.Error()
			}
			apiLog.ErrorMessage = strings.Join(errorMessages, "; ")
		}

		// 异步保存到数据库，避免影响请求性能
		go func() {
			// 添加recover防止panic影响程序
			defer func() {
				if r := recover(); r != nil {
					common.SysLog("panic in api request log: " + string(r.(string)))
				}
			}()

			// 先检查表大小，如果超过限制则清理旧数据
			if err := cleanupOldLogRecords(); err != nil {
				common.SysLog("failed to cleanup old api request logs: " + err.Error())
				// 即使清理失败，也尝试保存新记录
			}

			// 保存新记录
			if err := model.LOG_DB.Create(apiLog).Error; err != nil {
				common.SysLog("failed to record api request log: " + err.Error())
			}
		}()
	}
}

// cleanupOldLogRecords 清理旧的API请求日志记录
func cleanupOldLogRecords() error {
	// 检查当前记录数
	var count int64
	if err := model.LOG_DB.Model(&model.ApiRequestLog{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果记录数超过限制，删除最旧的记录
	if count > MaxLogRecords {
		// 计算需要删除的记录数
		deleteCount := count - MaxLogRecords/2 // 删除一半，避免频繁清理

		// 找到要保留的最小created_at
		var oldestRecord model.ApiRequestLog
		if err := model.LOG_DB.Order("created_at ASC").Offset(int(deleteCount)).First(&oldestRecord).Error; err != nil {
			return err
		}

		// 删除比这个时间点更早的记录
		if err := model.LOG_DB.Where("created_at < ?", oldestRecord.CreatedAt).Delete(&model.ApiRequestLog{}).Error; err != nil {
			return err
		}

		common.SysLog("cleaned up old api request logs, deleted records")
	}

	return nil
}