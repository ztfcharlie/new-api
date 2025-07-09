package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/lang"
	"one-api/model"
	relayconstant "one-api/relay/constant"
	"one-api/service"
	"one-api/setting"
	"one-api/setting/ratio_setting"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ModelRequest struct {
	Model string `json:"model"`
}

func Distribute() func(c *gin.Context) {
	return func(c *gin.Context) {
		allowIpsMap := common.GetContextKeyStringMap(c, constant.ContextKeyTokenAllowIps)
		if len(allowIpsMap) != 0 {
			clientIp := c.ClientIP()
			if _, ok := allowIpsMap[clientIp]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, lang.T(c, "distributor.error.ip_not_allowed"))
				return
			}
		}
		var channel *model.Channel
		channelId, ok := common.GetContextKey(c, constant.ContextKeyTokenSpecificChannelId)
		modelRequest, shouldSelectChannel, err := getModelRequest(c)
		if err != nil {
			abortWithOpenAiMessage(c, http.StatusBadRequest, fmt.Sprintf(lang.T(c, "distributor.error.invalid_request"), err.Error()))
			return
		}
		userGroup := common.GetContextKeyString(c, constant.ContextKeyUserGroup)
		tokenGroup := common.GetContextKeyString(c, constant.ContextKeyTokenGroup)
		if tokenGroup != "" {
			// check common.UserUsableGroups[userGroup]
			if _, ok := setting.GetUserUsableGroups(userGroup)[tokenGroup]; !ok {
				abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf(lang.T(c, "distributor.error.token_group_disabled"), tokenGroup))
				return
			}
			// check group in common.GroupRatio
			if !ratio_setting.ContainsGroupRatio(tokenGroup) {
				if tokenGroup != "auto" {
					abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf(lang.T(c, "distributor.error.group_deprecated"), tokenGroup))
					return
				}
			}
			userGroup = tokenGroup
		}
		common.SetContextKey(c, constant.ContextKeyUsingGroup, userGroup)
		if ok {
			id, err := strconv.Atoi(channelId.(string))
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, lang.T(c, "distributor.error.invalid_channel_id"))
				return
			}
			channel, err = model.GetChannelById(id, true)
			if err != nil {
				abortWithOpenAiMessage(c, http.StatusBadRequest, lang.T(c, "distributor.error.invalid_channel_id"))
				return
			}
			if channel.Status != common.ChannelStatusEnabled {
				abortWithOpenAiMessage(c, http.StatusForbidden, lang.T(c, "distributor.error.channel_disabled"))
				return
			}
		} else {
			// Select a channel for the user
			// check token model mapping
			modelLimitEnable := common.GetContextKeyBool(c, constant.ContextKeyTokenModelLimitEnabled)
			if modelLimitEnable {
				s, ok := common.GetContextKey(c, constant.ContextKeyTokenModelLimit)
				var tokenModelLimit map[string]bool
				if ok {
					tokenModelLimit = s.(map[string]bool)
				} else {
					tokenModelLimit = map[string]bool{}
				}
				if tokenModelLimit != nil {
					if _, ok := tokenModelLimit[modelRequest.Model]; !ok {
						abortWithOpenAiMessage(c, http.StatusForbidden, fmt.Sprintf(lang.T(c, "distributor.error.model_not_allowed"), modelRequest.Model))
						return
					}
				} else {
					abortWithOpenAiMessage(c, http.StatusForbidden, lang.T(c, "distributor.error.no_model_allowed"))
					return
				}
			}

			if shouldSelectChannel {
				var selectGroup string
				channel, selectGroup, err = model.CacheGetRandomSatisfiedChannel(c, userGroup, modelRequest.Model, 0)
				if err != nil {
					showGroup := userGroup
					if userGroup == "auto" {
						showGroup = fmt.Sprintf("auto(%s)", selectGroup)
					}
					message := fmt.Sprintf(lang.T(c, "distributor.error.no_channel"), showGroup, modelRequest.Model)
					// 如果错误，但是渠道不为空，说明是数据库一致性问题
					if channel != nil {
						common.SysError(fmt.Sprintf(lang.T(c, "not_exist_channel_id"), channel.Id))
						message = lang.T(c, "distributor.error.db_consistency")
					}
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, message)
					return
				}
				if channel == nil {
					abortWithOpenAiMessage(c, http.StatusServiceUnavailable, fmt.Sprintf(lang.T(c, "distributor.error.no_channel_db"), userGroup, modelRequest.Model))
					return
				}
			}
		}
		common.SetContextKey(c, constant.ContextKeyRequestStartTime, time.Now())
		SetupContextForSelectedChannel(c, channel, modelRequest.Model)
		c.Next()
	}
}

func getModelRequest(c *gin.Context) (*ModelRequest, bool, error) {
	var modelRequest ModelRequest
	shouldSelectChannel := true
	var err error
	if strings.Contains(c.Request.URL.Path, "/mj/") {
		relayMode := relayconstant.Path2RelayModeMidjourney(c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeMidjourneyTaskFetch ||
			relayMode == relayconstant.RelayModeMidjourneyTaskFetchByCondition ||
			relayMode == relayconstant.RelayModeMidjourneyNotify ||
			relayMode == relayconstant.RelayModeMidjourneyTaskImageSeed {
			shouldSelectChannel = false
		} else {
			midjourneyRequest := dto.MidjourneyRequest{}
			err = common.UnmarshalBodyReusable(c, &midjourneyRequest)
			if err != nil {
				return nil, false, err
			}
			midjourneyModel, mjErr, success := service.GetMjRequestModel(relayMode, &midjourneyRequest)
			if mjErr != nil {
				return nil, false, fmt.Errorf(mjErr.Description)
			}
			if midjourneyModel == "" {
				if !success {
					return nil, false, fmt.Errorf(lang.T(c, "distributor.error.invalid_request_model"))
				} else {
					// task fetch, task fetch by condition, notify
					shouldSelectChannel = false
				}
			}
			modelRequest.Model = midjourneyModel
		}
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/suno/") {
		relayMode := relayconstant.Path2RelaySuno(c.Request.Method, c.Request.URL.Path)
		if relayMode == relayconstant.RelayModeSunoFetch ||
			relayMode == relayconstant.RelayModeSunoFetchByID {
			shouldSelectChannel = false
		} else {
			modelName := service.CoverTaskActionToModelName(constant.TaskPlatformSuno, c.Param("action"))
			modelRequest.Model = modelName
		}
		c.Set("platform", string(constant.TaskPlatformSuno))
		c.Set("relay_mode", relayMode)
	} else if strings.Contains(c.Request.URL.Path, "/v1/video/generations") {
		err = common.UnmarshalBodyReusable(c, &modelRequest)
		var platform string
		var relayMode int
		if strings.HasPrefix(modelRequest.Model, "jimeng") {
			platform = string(constant.TaskPlatformJimeng)
			relayMode = relayconstant.Path2RelayJimeng(c.Request.Method, c.Request.URL.Path)
			if relayMode == relayconstant.RelayModeJimengFetchByID {
				shouldSelectChannel = false
			}
		} else {
			platform = string(constant.TaskPlatformKling)
			relayMode = relayconstant.Path2RelayKling(c.Request.Method, c.Request.URL.Path)
			if relayMode == relayconstant.RelayModeKlingFetchByID {
				shouldSelectChannel = false
			}
		}
		c.Set("platform", platform)
		c.Set("relay_mode", relayMode)
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1beta/models/") || strings.HasPrefix(c.Request.URL.Path, "/v1/models/") {
		// Gemini API 路径处理: /v1beta/models/gemini-2.0-flash:generateContent
		relayMode := relayconstant.RelayModeGemini
		modelName := extractModelNameFromGeminiPath(c.Request.URL.Path)
		if modelName != "" {
			modelRequest.Model = modelName
		}
		c.Set("relay_mode", relayMode)
	} else if !strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") && !strings.HasPrefix(c.Request.URL.Path, "/v1/images/edits") {
		err = common.UnmarshalBodyReusable(c, &modelRequest)
	}
	if err != nil {
		return nil, false, errors.New(fmt.Sprintf(lang.T(c, "distributor.error.invalid_request_with"), err.Error()))
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		//wss://api.openai.com/v1/realtime?model=gpt-4o-realtime-preview-2024-10-01
		modelRequest.Model = c.Query("model")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/moderations") {
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	}
	if strings.HasSuffix(c.Request.URL.Path, "embeddings") {
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations") {
		modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "dall-e")
	} else if strings.HasPrefix(c.Request.URL.Path, "/v1/images/edits") {
		modelRequest.Model = common.GetStringIfEmpty(c.PostForm("model"), "gpt-image-1")
	}
	if strings.HasPrefix(c.Request.URL.Path, "/v1/audio") {
		relayMode := relayconstant.RelayModeAudioSpeech
		if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/speech") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "tts-1")
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranslation
		} else if strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions") {
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, c.PostForm("model"))
			modelRequest.Model = common.GetStringIfEmpty(modelRequest.Model, "whisper-1")
			relayMode = relayconstant.RelayModeAudioTranscription
		}
		c.Set("relay_mode", relayMode)
	}
	return &modelRequest, shouldSelectChannel, nil
}

func SetupContextForSelectedChannel(c *gin.Context, channel *model.Channel, modelName string) {
	c.Set("original_model", modelName) // for retry
	if channel == nil {
		return
	}
	c.Set("channel_id", channel.Id)
	c.Set("channel_name", channel.Name)
	c.Set("channel_type", channel.Type)
	c.Set("channel_create_time", channel.CreatedTime)
	c.Set("channel_setting", channel.GetSetting())
	c.Set("param_override", channel.GetParamOverride())
	if nil != channel.OpenAIOrganization && "" != *channel.OpenAIOrganization {
		c.Set("channel_organization", *channel.OpenAIOrganization)
	}
	c.Set("auto_ban", channel.GetAutoBan())
	c.Set("model_mapping", channel.GetModelMapping())
	c.Set("status_code_mapping", channel.GetStatusCodeMapping())
	c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.Key))
	c.Set("base_url", channel.GetBaseURL())
	// TODO: api_version统一
	switch channel.Type {
	case constant.ChannelTypeAzure:
		c.Set("api_version", channel.Other)
	case constant.ChannelTypeVertexAi:
		c.Set("region", channel.Other)
	case constant.ChannelTypeXunfei:
		c.Set("api_version", channel.Other)
	case constant.ChannelTypeGemini:
		c.Set("api_version", channel.Other)
	case constant.ChannelTypeAli:
		c.Set("plugin", channel.Other)
	case constant.ChannelCloudflare:
		c.Set("api_version", channel.Other)
	case constant.ChannelTypeMokaAI:
		c.Set("api_version", channel.Other)
	case constant.ChannelTypeCoze:
		c.Set("bot_id", channel.Other)
	}
}

// extractModelNameFromGeminiPath 从 Gemini API URL 路径中提取模型名
// 输入格式: /v1beta/models/gemini-2.0-flash:generateContent
// 输出: gemini-2.0-flash
func extractModelNameFromGeminiPath(path string) string {
	// 查找 "/models/" 的位置
	modelsPrefix := "/models/"
	modelsIndex := strings.Index(path, modelsPrefix)
	if modelsIndex == -1 {
		return ""
	}

	// 从 "/models/" 之后开始提取
	startIndex := modelsIndex + len(modelsPrefix)
	if startIndex >= len(path) {
		return ""
	}

	// 查找 ":" 的位置，模型名在 ":" 之前
	colonIndex := strings.Index(path[startIndex:], ":")
	if colonIndex == -1 {
		// 如果没有找到 ":"，返回从 "/models/" 到路径结尾的部分
		return path[startIndex:]
	}

	// 返回模型名部分
	return path[startIndex : startIndex+colonIndex]
}
