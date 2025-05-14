package common

import (
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	relayconstant "one-api/relay/constant"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ThinkingContentInfo struct {
	IsFirstThinkingContent  bool
	SendLastThinkingContent bool
	HasSentThinkingContent  bool
}

const (
	LastMessageTypeNone     = "none"
	LastMessageTypeText     = "text"
	LastMessageTypeTools    = "tools"
	LastMessageTypeThinking = "thinking"
)

type ClaudeConvertInfo struct {
	LastMessagesType string
	Index            int
	Usage            *dto.Usage
	FinishReason     string
	Done             bool
}

const (
	RelayFormatOpenAI = "openai"
	RelayFormatClaude = "claude"
	RelayFormatGemini = "gemini"
)

type RerankerInfo struct {
	Documents       []any
	ReturnDocuments bool
}

type BuildInToolInfo struct {
	ToolName          string
	CallCount         int
	SearchContextSize string
}

type ResponsesUsageInfo struct {
	BuiltInTools map[string]*BuildInToolInfo
}

type RelayInfo struct {
	ChannelType       int
	ChannelId         int
	TokenId           int
	TokenKey          string
	UserId            int
	Group             string
	TokenUnlimited    bool
	StartTime         time.Time
	FirstResponseTime time.Time
	isFirstResponse   bool
	//SendLastReasoningResponse bool
	ApiType           int
	IsStream          bool
	IsPlayground      bool
	UsePrice          bool
	RelayMode         int
	UpstreamModelName string
	OriginModelName   string
	//RecodeModelName      string
	RequestURLPath       string
	ApiVersion           string
	PromptTokens         int
	ApiKey               string
	Organization         string
	BaseUrl              string
	SupportStreamOptions bool
	ShouldIncludeUsage   bool
	IsModelMapped        bool
	ClientWs             *websocket.Conn
	TargetWs             *websocket.Conn
	InputAudioFormat     string
	OutputAudioFormat    string
	RealtimeTools        []dto.RealTimeTool
	IsFirstRequest       bool
	AudioUsage           bool
	ReasoningEffort      string
	ChannelSetting       map[string]interface{}
	ParamOverride        map[string]interface{}
	UserSetting          map[string]interface{}
	UserEmail            string
	UserQuota            int
	RelayFormat          string
	SendResponseCount    int
	ChannelCreateTime    int64
	ThinkingContentInfo
	*ClaudeConvertInfo
	*RerankerInfo
	*ResponsesUsageInfo
}

// 定义支持流式选项的通道类型
var streamSupportedChannels = map[int]bool{
	common.ChannelTypeOpenAI:     true,
	common.ChannelTypeAnthropic:  true,
	common.ChannelTypeAws:        true,
	common.ChannelTypeGemini:     true,
	common.ChannelCloudflare:     true,
	common.ChannelTypeAzure:      true,
	common.ChannelTypeVolcEngine: true,
	common.ChannelTypeOllama:     true,
	common.ChannelTypeXai:        true,
	common.ChannelTypeDeepSeek:   true,
	common.ChannelTypeBaiduV2:    true,
}

func GenRelayInfoWs(c *gin.Context, ws *websocket.Conn) *RelayInfo {
	info := GenRelayInfo(c)
	info.ClientWs = ws
	info.InputAudioFormat = "pcm16"
	info.OutputAudioFormat = "pcm16"
	info.IsFirstRequest = true
	return info
}

func GenRelayInfoClaude(c *gin.Context) *RelayInfo {
	info := GenRelayInfo(c)
	info.RelayFormat = RelayFormatClaude
	info.ShouldIncludeUsage = false
	info.ClaudeConvertInfo = &ClaudeConvertInfo{
		LastMessagesType: LastMessageTypeNone,
	}
	return info
}

func GenRelayInfoRerank(c *gin.Context, req *dto.RerankRequest) *RelayInfo {
	info := GenRelayInfo(c)
	info.RelayMode = relayconstant.RelayModeRerank
	info.RerankerInfo = &RerankerInfo{
		Documents:       req.Documents,
		ReturnDocuments: req.GetReturnDocuments(),
	}
	return info
}

func GenRelayInfoResponses(c *gin.Context, req *dto.OpenAIResponsesRequest) *RelayInfo {
	info := GenRelayInfo(c)
	info.RelayMode = relayconstant.RelayModeResponses
	info.ResponsesUsageInfo = &ResponsesUsageInfo{
		BuiltInTools: make(map[string]*BuildInToolInfo),
	}
	if len(req.Tools) > 0 {
		for _, tool := range req.Tools {
			info.ResponsesUsageInfo.BuiltInTools[tool.Type] = &BuildInToolInfo{
				ToolName:  tool.Type,
				CallCount: 0,
			}
			switch tool.Type {
			case dto.BuildInToolWebSearchPreview:
				if tool.SearchContextSize == "" {
					tool.SearchContextSize = "medium"
				}
				info.ResponsesUsageInfo.BuiltInTools[tool.Type].SearchContextSize = tool.SearchContextSize
			}
		}
	}
	info.IsStream = req.Stream
	return info
}

func GenRelayInfo(c *gin.Context) *RelayInfo {
	channelType := c.GetInt("channel_type")
	channelId := c.GetInt("channel_id")
	channelSetting := c.GetStringMap("channel_setting")
	paramOverride := c.GetStringMap("param_override")

	tokenId := c.GetInt("token_id")
	tokenKey := c.GetString("token_key")
	userId := c.GetInt("id")
	group := c.GetString("group")
	tokenUnlimited := c.GetBool("token_unlimited_quota")
	startTime := c.GetTime(constant.ContextKeyRequestStartTime)
	// firstResponseTime = time.Now() - 1 second

	apiType, _ := relayconstant.ChannelType2APIType(channelType)

	info := &RelayInfo{
		UserQuota:         c.GetInt(constant.ContextKeyUserQuota),
		UserSetting:       c.GetStringMap(constant.ContextKeyUserSetting),
		UserEmail:         c.GetString(constant.ContextKeyUserEmail),
		isFirstResponse:   true,
		RelayMode:         relayconstant.Path2RelayMode(c.Request.URL.Path),
		BaseUrl:           c.GetString("base_url"),
		RequestURLPath:    c.Request.URL.String(),
		ChannelType:       channelType,
		ChannelId:         channelId,
		TokenId:           tokenId,
		TokenKey:          tokenKey,
		UserId:            userId,
		Group:             group,
		TokenUnlimited:    tokenUnlimited,
		StartTime:         startTime,
		FirstResponseTime: startTime.Add(-time.Second),
		OriginModelName:   c.GetString("original_model"),
		UpstreamModelName: c.GetString("original_model"),
		//RecodeModelName:   c.GetString("original_model"),
		IsModelMapped:     false,
		ApiType:           apiType,
		ApiVersion:        c.GetString("api_version"),
		ApiKey:            strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		Organization:      c.GetString("channel_organization"),
		ChannelSetting:    channelSetting,
		ChannelCreateTime: c.GetInt64("channel_create_time"),
		ParamOverride:     paramOverride,
		RelayFormat:       RelayFormatOpenAI,
		ThinkingContentInfo: ThinkingContentInfo{
			IsFirstThinkingContent:  true,
			SendLastThinkingContent: false,
		},
	}
	if strings.HasPrefix(c.Request.URL.Path, "/pg") {
		info.IsPlayground = true
		info.RequestURLPath = strings.TrimPrefix(info.RequestURLPath, "/pg")
		info.RequestURLPath = "/v1" + info.RequestURLPath
	}
	if info.BaseUrl == "" {
		info.BaseUrl = common.ChannelBaseURLs[channelType]
	}
	if info.ChannelType == common.ChannelTypeAzure {
		info.ApiVersion = GetAPIVersion(c)
	}
	if info.ChannelType == common.ChannelTypeVertexAi {
		info.ApiVersion = c.GetString("region")
	}
	if streamSupportedChannels[info.ChannelType] {
		info.SupportStreamOptions = true
	}
	// responses 模式不支持 StreamOptions
	if relayconstant.RelayModeResponses == info.RelayMode {
		info.SupportStreamOptions = false
	}
	return info
}

func (info *RelayInfo) SetPromptTokens(promptTokens int) {
	info.PromptTokens = promptTokens
}

func (info *RelayInfo) SetIsStream(isStream bool) {
	info.IsStream = isStream
}

func (info *RelayInfo) SetFirstResponseTime() {
	if info.isFirstResponse {
		info.FirstResponseTime = time.Now()
		info.isFirstResponse = false
	}
}

func (info *RelayInfo) HasSendResponse() bool {
	return info.FirstResponseTime.After(info.StartTime)
}

type TaskRelayInfo struct {
	*RelayInfo
	Action       string
	OriginTaskID string

	ConsumeQuota bool
}

func GenTaskRelayInfo(c *gin.Context) *TaskRelayInfo {
	info := &TaskRelayInfo{
		RelayInfo: GenRelayInfo(c),
	}
	return info
}
