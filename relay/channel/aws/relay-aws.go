package aws

import (
	"encoding/json"                    // 用于JSON编码和解码
	"fmt"                              // 格式化I/O
	"net/http"                         // HTTP客户端和服务器实现
	"one-api/common"                   // 项目中的通用模块
	"one-api/dto"                      // 数据传输对象模块
	"one-api/relay/channel/claude"     // Claude相关的模块
	relaycommon "one-api/relay/common" // 继电器通用模块
	"strings"                          // 字符串操作

	"github.com/gin-gonic/gin" // Gin框架，用于构建Web应用
	"github.com/pkg/errors"    // 错误处理

	"github.com/aws/aws-sdk-go-v2/aws"                          // AWS SDK
	"github.com/aws/aws-sdk-go-v2/credentials"                  // AWS凭证
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"       // Bedrock运行时服务
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types" // Bedrock运行时类型
)

// 创建一个新的AWS客户端
func newAwsClient(c *gin.Context, info *relaycommon.RelayInfo) (*bedrockruntime.Client, error) {
	awsSecret := strings.Split(info.ApiKey, "|") // 从API密钥中提取AWS凭证
	if len(awsSecret) != 3 {
		return nil, errors.New("invalid aws secret key") // 如果密钥格式不正确，返回错误
	}
	ak := awsSecret[0]     // 访问密钥
	sk := awsSecret[1]     // 秘密密钥
	region := awsSecret[2] // 区域
	client := bedrockruntime.New(bedrockruntime.Options{
		Region:      region,                                                                        // 设置区域
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(ak, sk, "")), // 设置凭证
	})

	return client, nil // 返回AWS客户端
}

// 包装错误信息
func wrapErr(err error) *dto.OpenAIErrorWithStatusCode {
	return &dto.OpenAIErrorWithStatusCode{
		StatusCode: http.StatusInternalServerError, // 设置HTTP状态码
		Error: dto.OpenAIError{
			Message: fmt.Sprintf("%s", err.Error()), // 错误信息
		},
	}
}

// 获取AWS区域前缀
func awsRegionPrefix(awsRegionId string) string {
	parts := strings.Split(awsRegionId, "-") // 按“-”分割区域ID
	regionPrefix := ""
	if len(parts) > 0 {
		regionPrefix = parts[0] // 取第一个部分作为前缀
	}
	return regionPrefix
}

// 判断AWS模型是否可以跨区域
func awsModelCanCrossRegion(awsModelId, awsRegionPrefix string) bool {
	regionSet, exists := awsModelCanCrossRegionMap[awsModelId] // 从映射中获取区域集
	return exists && regionSet[awsRegionPrefix]                // 判断区域前缀是否存在于区域集中
}

// 获取跨区域的AWS模型ID
func awsModelCrossRegion(awsModelId, awsRegionPrefix string) string {
	modelPrefix, find := awsRegionCrossModelPrefixMap[awsRegionPrefix] // 获取模型前缀
	if !find {
		return awsModelId // 如果找不到，返回原模型ID
	}
	return modelPrefix + "." + awsModelId // 返回跨区域模型ID
}

// 获取AWS模型ID
func awsModelID(requestModel string) (string, error) {
	if awsModelID, ok := awsModelIDMap[requestModel]; ok {
		return awsModelID, nil // 如果找到模型ID，返回
	}

	return requestModel, nil // 否则返回请求的模型ID
}

// 处理AWS请求
func awsHandler(c *gin.Context, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	awsCli, err := newAwsClient(c, info) // 创建AWS客户端
	if err != nil {
		return wrapErr(errors.Wrap(err, "newAwsClient")), nil // 错误处理
	}

	awsModelId, err := awsModelID(c.GetString("request_model")) // 获取AWS模型ID
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil // 错误处理
	}

	awsRegionPrefix := awsRegionPrefix(awsCli.Options().Region)           // 获取区域前缀
	canCrossRegion := awsModelCanCrossRegion(awsModelId, awsRegionPrefix) // 判断是否可以跨区域
	if canCrossRegion {
		awsModelId = awsModelCrossRegion(awsModelId, awsRegionPrefix) // 获取跨区域模型ID
	}

	awsReq := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(awsModelId),         // 模型ID
		Accept:      aws.String("application/json"), // 接受类型
		ContentType: aws.String("application/json"), // 内容类型
	}

	claudeReq_, ok := c.Get("converted_request") // 获取转换后的请求
	if !ok {
		return wrapErr(errors.New("request not found")), nil // 错误处理
	}
	claudeReq := claudeReq_.(*dto.ClaudeRequest)  // 类型断言
	awsClaudeReq := copyRequest(claudeReq)        // 复制请求
	awsReq.Body, err = json.Marshal(awsClaudeReq) // 序列化请求体
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil // 错误处理
	}

	awsResp, err := awsCli.InvokeModel(c.Request.Context(), awsReq) // 调用AWS模型
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModel")), nil // 错误处理
	}

	claudeInfo := &claude.ClaudeResponseInfo{
		ResponseId:   fmt.Sprintf("chatcmpl-%s", common.GetUUID()), // 响应ID
		Created:      common.GetTimestamp(),                        // 创建时间戳
		Model:        info.UpstreamModelName,                       // 模型名称
		ResponseText: strings.Builder{},                            // 响应文本
		Usage:        &dto.Usage{},                                 // 使用情况
	}

	claude.HandleClaudeResponseData(c, info, claudeInfo, awsResp.Body, RequestModeMessage) // 处理Claude响应数据
	return nil, claudeInfo.Usage                                                           // 返回使用情况
}
func awsStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	awsCli, err := newAwsClient(c, info)
	if err != nil {
		return wrapErr(errors.Wrap(err, "newAwsClient")), nil
	}

	awsModelId, err := awsModelID(c.GetString("request_model"))
	if err != nil {
		return wrapErr(errors.Wrap(err, "awsModelID")), nil
	}

	awsRegionPrefix := awsRegionPrefix(awsCli.Options().Region)
	canCrossRegion := awsModelCanCrossRegion(awsModelId, awsRegionPrefix)
	if canCrossRegion {
		awsModelId = awsModelCrossRegion(awsModelId, awsRegionPrefix)
	}

	awsReq := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(awsModelId),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	}

	claudeReq_, ok := c.Get("converted_request")
	if !ok {
		return wrapErr(errors.New("request not found")), nil
	}
	claudeReq := claudeReq_.(*dto.ClaudeRequest)

	awsClaudeReq := copyRequest(claudeReq)
	awsReq.Body, err = json.Marshal(awsClaudeReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "marshal request")), nil
	}

	awsResp, err := awsCli.InvokeModelWithResponseStream(c.Request.Context(), awsReq)
	if err != nil {
		return wrapErr(errors.Wrap(err, "InvokeModelWithResponseStream")), nil
	}
	stream := awsResp.GetStream()
	defer stream.Close()

	claudeInfo := &claude.ClaudeResponseInfo{
		ResponseId:   fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Created:      common.GetTimestamp(),
		Model:        info.UpstreamModelName,
		ResponseText: strings.Builder{},
		Usage:        &dto.Usage{},
	}

	// 添加计数器
	chunkCount := 0

	for event := range stream.Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			chunkCount++ // 增加计数器
			info.SetFirstResponseTime()
			respErr := claude.HandleStreamResponseData(c, info, claudeInfo, string(v.Value.Bytes), RequestModeMessage)
			if respErr != nil {
				return respErr, nil
			}
			// 记录每个数据块
			//fmt.Printf("Received chunk #%d: %s\n", chunkCount, string(v.Value.Bytes))
		case *types.UnknownUnionMember:
			//fmt.Println("unknown tag:", v.Tag)
			return wrapErr(errors.New("unknown response type")), nil
		default:
			//fmt.Println("union is nil or unknown type")
			return wrapErr(errors.New("nil or unknown response type")), nil
		}
	}

	//fmt.Printf("Total chunks received: %d\n", chunkCount)

	claude.HandleStreamFinalResponse(c, info, claudeInfo, RequestModeMessage)
	return nil, claudeInfo.Usage
}
