package controller

import (
	"errors"
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/lang"
	"one-api/middleware"
	"one-api/model"
	"one-api/service"
	"one-api/setting"
	"time"

	"github.com/gin-gonic/gin"
)

func Playground(c *gin.Context) {
	var openaiErr *dto.OpenAIErrorWithStatusCode

	defer func() {
		if openaiErr != nil {
			c.JSON(openaiErr.StatusCode, gin.H{
				"error": openaiErr.Error,
			})
		}
	}()

	useAccessToken := c.GetBool("use_access_token")
	if useAccessToken {
		openaiErr = service.OpenAIErrorWrapperLocal(
			errors.New(lang.T(c, "playground.error.access_token")),
			"access_token_not_supported",
			http.StatusBadRequest,
		)
		return
	}

	playgroundRequest := &dto.PlayGroundRequest{}
	err := common.UnmarshalBodyReusable(c, playgroundRequest)
	if err != nil {
		openaiErr = service.OpenAIErrorWrapperLocal(
			err,
			"unmarshal_request_failed",
			http.StatusBadRequest,
		)
		return
	}

	if playgroundRequest.Model == "" {
		openaiErr = service.OpenAIErrorWrapperLocal(
			errors.New(lang.T(c, "playground.error.model_required")),
			"model_required",
			http.StatusBadRequest,
		)
		return
	}
	c.Set("original_model", playgroundRequest.Model)
	group := playgroundRequest.Group
	userGroup := c.GetString("group")

	if group == "" {
		group = userGroup
	} else {
		if !setting.GroupInUserUsableGroups(group) && group != userGroup {
			openaiErr = service.OpenAIErrorWrapperLocal(
				errors.New(lang.T(c, "playground.error.group_not_allowed")),
				"group_not_allowed",
				http.StatusForbidden,
			)
			return
		}
		c.Set("group", group)
	}
	c.Set("token_name", "playground-"+group)
	channel, err := model.CacheGetRandomSatisfiedChannel(group, playgroundRequest.Model, 0)
	if err != nil {
		message := fmt.Sprintf(
			lang.T(c, "playground.error.no_channel"),
			group,
			playgroundRequest.Model,
		)
		openaiErr = service.OpenAIErrorWrapperLocal(
			errors.New(message),
			"get_playground_channel_failed",
			http.StatusInternalServerError,
		)
		return
	}
	middleware.SetupContextForSelectedChannel(c, channel, playgroundRequest.Model)
	c.Set(constant.ContextKeyRequestStartTime, time.Now())
	Relay(c)
}
