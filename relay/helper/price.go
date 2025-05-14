package helper

import (
	"fmt"
	"one-api/common"
	constant2 "one-api/constant"
	"one-api/lang"
	relaycommon "one-api/relay/common"
	"one-api/setting"
	"one-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

type PriceData struct {
	ModelPrice             float64
	ModelRatio             float64
	CompletionRatio        float64
	CacheRatio             float64
	CacheCreationRatio     float64
	ImageRatio             float64
	GroupRatio             float64
	UsePrice               bool
	ShouldPreConsumedQuota int
}

func (p PriceData) ToSetting() string {
	return fmt.Sprintf("ModelPrice: %f, ModelRatio: %f, CompletionRatio: %f, CacheRatio: %f, GroupRatio: %f, UsePrice: %t, CacheCreationRatio: %f, ShouldPreConsumedQuota: %d, ImageRatio: %f", p.ModelPrice, p.ModelRatio, p.CompletionRatio, p.CacheRatio, p.GroupRatio, p.UsePrice, p.CacheCreationRatio, p.ShouldPreConsumedQuota, p.ImageRatio)
}

func ModelPriceHelper(c *gin.Context, info *relaycommon.RelayInfo, promptTokens int, maxTokens int) (PriceData, error) {
	modelPrice, usePrice := operation_setting.GetModelPrice(info.OriginModelName, false)
	groupRatio := setting.GetGroupRatio(info.Group)
	var preConsumedQuota int
	var modelRatio float64
	var completionRatio float64
	var cacheRatio float64
	var imageRatio float64
	var cacheCreationRatio float64
	if !usePrice {
		preConsumedTokens := common.PreConsumedQuota
		if maxTokens != 0 {
			preConsumedTokens = promptTokens + maxTokens
		}
		var success bool
		modelRatio, success = operation_setting.GetModelRatio(info.OriginModelName)
		if !success {
			acceptUnsetRatio := false
			if accept, ok := info.UserSetting[constant2.UserAcceptUnsetRatioModel]; ok {
				b, ok := accept.(bool)
				if ok {
					acceptUnsetRatio = b
				}
			}
			if !acceptUnsetRatio {
				if info.UserId == 1 {
					return PriceData{}, fmt.Errorf(
						lang.T(c, "price.error.model_ratio_admin"),
						info.OriginModelName,
						info.OriginModelName,
					)
				} else {
					return PriceData{}, fmt.Errorf(
						lang.T(c, "price.error.model_ratio_user"),
						info.OriginModelName,
						info.OriginModelName,
					)
				}
			}
		}
		completionRatio = operation_setting.GetCompletionRatio(info.OriginModelName)
		cacheRatio, _ = operation_setting.GetCacheRatio(info.OriginModelName)
		cacheCreationRatio, _ = operation_setting.GetCreateCacheRatio(info.OriginModelName)
		imageRatio, _ = operation_setting.GetImageRatio(info.OriginModelName)
		ratio := modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
	}

	priceData := PriceData{
		ModelPrice:             modelPrice,
		ModelRatio:             modelRatio,
		CompletionRatio:        completionRatio,
		GroupRatio:             groupRatio,
		UsePrice:               usePrice,
		CacheRatio:             cacheRatio,
		ImageRatio:             imageRatio,
		CacheCreationRatio:     cacheCreationRatio,
		ShouldPreConsumedQuota: preConsumedQuota,
	}

	if common.DebugEnabled {
		println(fmt.Sprintf(lang.T(c, "price.debug.helper_result"), priceData.ToSetting()))
	}

	return priceData, nil
}

func ContainPriceOrRatio(modelName string) bool {
	_, ok := operation_setting.GetModelPrice(modelName, false)
	if ok {
		return true
	}
	_, ok = operation_setting.GetModelRatio(modelName)
	if ok {
		return true
	}
	return false
}
