package common

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)


// LogUpstreamRequest 记录转发给大模型API的请求头和请求体
func LogUpstreamRequest(c *gin.Context, headers interface{}, body io.Reader) {
	// 检查是否启用API请求日志
	if !ApiRequestLogEnabled {
		return
	}

	// 从context中获取上游请求信息存储对象
	upstreamInfoInterface, exists := c.Get("upstream_request_info")
	if !exists {
		return
	}

	// 类型断言
	upstreamInfo, ok := upstreamInfoInterface.(*UpstreamRequestInfo)
	if !ok {
		return
	}

	// 记录请求头
	if headers != nil {
		var headersJSON string
		switch h := headers.(type) {
		case map[string]string:
			// 过滤敏感信息
			filteredHeaders := make(map[string]string)
			for k, v := range h {
				lowerKey := strings.ToLower(k)
				if lowerKey == "authorization" ||
					lowerKey == "cookie" ||
					lowerKey == "set-cookie" ||
					lowerKey == "api-key" ||
					lowerKey == "x-api-key" {
					// 对敏感信息进行脱敏处理
					if len(v) > 8 {
						filteredHeaders[k] = v[:4] + "***" + v[len(v)-4:]
					} else {
						filteredHeaders[k] = "***"
					}
				} else if strings.Contains(lowerKey, "goog-api-key") {
					// 对 Google API Key 进行脱敏处理
					if len(v) > 8 {
						filteredHeaders[k] = v[:4] + "***" + v[len(v)-4:]
					} else {
						filteredHeaders[k] = "***"
					}
				} else {
					filteredHeaders[k] = v
				}
			}
			headersJSON = safeJSONMarshal(filteredHeaders)
		default:
			headersJSON = safeJSONMarshal(headers)
		}
		upstreamInfo.Headers = headersJSON
	}

	// 记录请求体
	if body != nil {
		// 尝试读取请求体内容
		bodyBytes, err := io.ReadAll(body)
		if err == nil && len(bodyBytes) > 0 {
			// 关闭原始body
			if c, ok := body.(io.ReadCloser); ok {
				c.Close()
			}
			upstreamInfo.Body = string(bodyBytes)
		}
	}
}

// LogUpstreamRequestWithBodyBytes 记录转发给大模型API的请求头和请求体（使用字节切片）
func LogUpstreamRequestWithBodyBytes(c *gin.Context, headers interface{}, bodyBytes []byte) {
	// 检查是否启用API请求日志
	if !ApiRequestLogEnabled {
		return
	}

	// 从context中获取上游请求信息存储对象
	upstreamInfoInterface, exists := c.Get("upstream_request_info")
	if !exists {
		return
	}

	// 类型断言
	upstreamInfo, ok := upstreamInfoInterface.(*UpstreamRequestInfo)
	if !ok {
		return
	}

	// 记录请求头
	if headers != nil {
		var headersJSON string
		switch h := headers.(type) {
		case map[string]string:
			// 过滤敏感信息
			filteredHeaders := make(map[string]string)
			for k, v := range h {
				lowerKey := strings.ToLower(k)
				if lowerKey == "authorization" ||
					lowerKey == "cookie" ||
					lowerKey == "set-cookie" ||
					lowerKey == "api-key" ||
					lowerKey == "x-api-key" {
					// 对敏感信息进行脱敏处理
					if len(v) > 8 {
						filteredHeaders[k] = v[:4] + "***" + v[len(v)-4:]
					} else {
						filteredHeaders[k] = "***"
					}
				} else if strings.Contains(lowerKey, "goog-api-key") {
					// 对 Google API Key 进行脱敏处理
					if len(v) > 8 {
						filteredHeaders[k] = v[:4] + "***" + v[len(v)-4:]
					} else {
						filteredHeaders[k] = "***"
					}
				} else {
					filteredHeaders[k] = v
				}
			}
			headersJSON = safeJSONMarshal(filteredHeaders)
		default:
			headersJSON = safeJSONMarshal(headers)
		}
		upstreamInfo.Headers = headersJSON
	}

	// 记录请求体
	if len(bodyBytes) > 0 {
		upstreamInfo.Body = string(bodyBytes)
	}
}

// safeJSONMarshal 安全的JSON序列化，避免panic
func safeJSONMarshal(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}