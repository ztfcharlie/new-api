package relay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/lang"
	"one-api/model"
	relaycommon "one-api/relay/common"
	relayconstant "one-api/relay/constant"
	"one-api/relay/helper"
	"one-api/service"
	"one-api/setting"
	"strings"

	"github.com/gin-gonic/gin"
)

func getAndValidImageRequest(c *gin.Context, info *relaycommon.RelayInfo) (*dto.ImageRequest, error) {
	imageRequest := &dto.ImageRequest{}

	switch info.RelayMode {
	case relayconstant.RelayModeImagesEdits:
		_, err := c.MultipartForm()
		if err != nil {
			return nil, err
		}
		formData := c.Request.PostForm
		imageRequest.Prompt = formData.Get("prompt")
		imageRequest.Model = formData.Get("model")
		imageRequest.N = common.String2Int(formData.Get("n"))
		imageRequest.Quality = formData.Get("quality")
		imageRequest.Size = formData.Get("size")

		if imageRequest.Model == "gpt-image-1" {
			if imageRequest.Quality == "" {
				imageRequest.Quality = "standard"
			}
		}
		if imageRequest.N == 0 {
			imageRequest.N = 1
		}
	default:
		err := common.UnmarshalBodyReusable(c, imageRequest)
		if err != nil {
			return nil, err
		}

		if imageRequest.Model == "" {
			imageRequest.Model = "dall-e-3"
		}

		if strings.Contains(imageRequest.Size, "×") {
			return nil, errors.New("size an unexpected error occurred in the parameter, please use 'x' instead of the multiplication sign '×'")
		}

		// Not "256x256", "512x512", or "1024x1024"
		if imageRequest.Model == "dall-e-2" || imageRequest.Model == "dall-e" {
			if imageRequest.Size != "" && imageRequest.Size != "256x256" && imageRequest.Size != "512x512" && imageRequest.Size != "1024x1024" {
				return nil, errors.New("size must be one of 256x256, 512x512, or 1024x1024 for dall-e-2 or dall-e")
			}
			if imageRequest.Size == "" {
				imageRequest.Size = "1024x1024"
			}
		} else if imageRequest.Model == "dall-e-3" {
			if imageRequest.Size != "" && imageRequest.Size != "1024x1024" && imageRequest.Size != "1024x1792" && imageRequest.Size != "1792x1024" {
				return nil, errors.New("size must be one of 1024x1024, 1024x1792 or 1792x1024 for dall-e-3")
			}
			if imageRequest.Quality == "" {
				imageRequest.Quality = "standard"
			}
			if imageRequest.Size == "" {
				imageRequest.Size = "1024x1024"
			}
		} else if imageRequest.Model == "gpt-image-1" {
			if imageRequest.Quality == "" {
				imageRequest.Quality = "auto"
			}
		}

		if imageRequest.Prompt == "" {
			return nil, errors.New("prompt is required")
		}

		if imageRequest.N == 0 {
			imageRequest.N = 1
		}
	}

	if setting.ShouldCheckPromptSensitive() {
		words, err := service.CheckSensitiveInput(imageRequest.Prompt)
		if err != nil {
			common.LogWarn(c, fmt.Sprintf(lang.T(c, "image.log.sensitive_words"), strings.Join(words, ",")))
			return nil, err
		}
	}
	return imageRequest, nil
}

func ImageHelper(c *gin.Context) *dto.OpenAIErrorWithStatusCode {
	relayInfo := relaycommon.GenRelayInfo(c)

	imageRequest, err := getAndValidImageRequest(c, relayInfo)
	if err != nil {
		common.LogError(c, fmt.Sprintf(lang.T(c, "image.error.get_request"), err.Error()))
		return service.OpenAIErrorWrapper(err, "invalid_image_request", http.StatusBadRequest)
	}

	err = helper.ModelMappedHelper(c, relayInfo)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_mapped_error", http.StatusInternalServerError)
	}

	imageRequest.Model = relayInfo.UpstreamModelName

	priceData, err := helper.ModelPriceHelper(c, relayInfo, len(imageRequest.Prompt), 0)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "model_price_error", http.StatusInternalServerError)
	}
	var preConsumedQuota int
	var quota int
	var userQuota int
	if !priceData.UsePrice {
		// modelRatio 16 = modelPrice $0.04
		// per 1 modelRatio = $0.04 / 16
		// priceData.ModelPrice = 0.0025 * priceData.ModelRatio
		var openaiErr *dto.OpenAIErrorWithStatusCode
		preConsumedQuota, userQuota, openaiErr = preConsumeQuota(c, priceData.ShouldPreConsumedQuota, relayInfo)
		if openaiErr != nil {
			return openaiErr
		}
		defer func() {
			if openaiErr != nil {
				returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
			}
		}()

	} else {
		sizeRatio := 1.0
		// Size
		if imageRequest.Size == "256x256" {
			sizeRatio = 0.4
		} else if imageRequest.Size == "512x512" {
			sizeRatio = 0.45
		} else if imageRequest.Size == "1024x1024" {
			sizeRatio = 1
		} else if imageRequest.Size == "1024x1792" || imageRequest.Size == "1792x1024" {
			sizeRatio = 2
		}

		qualityRatio := 1.0
		if imageRequest.Model == "dall-e-3" && imageRequest.Quality == "hd" {
			qualityRatio = 2.0
			if imageRequest.Size == "1024x1792" || imageRequest.Size == "1792x1024" {
				qualityRatio = 1.5
			}
		}

		// reset model price
		priceData.ModelPrice *= sizeRatio * qualityRatio * float64(imageRequest.N)
		quota = int(priceData.ModelPrice * priceData.GroupRatio * common.QuotaPerUnit)
		userQuota, err = model.GetUserQuota(relayInfo.UserId, false)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "get_user_quota_failed", http.StatusInternalServerError)
		}
		if userQuota-quota < 0 {
			return service.OpenAIErrorWrapperLocal(fmt.Errorf(
				lang.T(c, "image.error.insufficient_quota"),
				common.FormatQuota(userQuota),
				common.FormatQuota(quota),
			), "insufficient_user_quota", http.StatusForbidden)
		}
	}

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf(
			lang.T(c, "image.error.invalid_api_type"),
			relayInfo.ApiType,
		), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(relayInfo)

	var requestBody io.Reader

	convertedRequest, err := adaptor.ConvertImageRequest(c, relayInfo, *imageRequest)
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "convert_request_failed", http.StatusInternalServerError)
	}
	if relayInfo.RelayMode == relayconstant.RelayModeImagesEdits {
		requestBody = convertedRequest.(io.Reader)
	} else {
		jsonData, err := json.Marshal(convertedRequest)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "json_marshal_failed", http.StatusInternalServerError)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	if common.DebugEnabled {
		println(fmt.Sprintf("image request body: %s", requestBody))
	}

	statusCodeMappingStr := c.GetString("status_code_mapping")

	resp, err := adaptor.DoRequest(c, relayInfo, requestBody)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	var httpResp *http.Response
	if resp != nil {
		httpResp = resp.(*http.Response)
		relayInfo.IsStream = relayInfo.IsStream || strings.HasPrefix(httpResp.Header.Get("Content-Type"), "text/event-stream")
		if httpResp.StatusCode != http.StatusOK {
			openaiErr := service.RelayErrorHandler(httpResp, false)
			// reset status code 重置状态码
			service.ResetStatusCode(openaiErr, statusCodeMappingStr)
			return openaiErr
		}
	}

	usage, openaiErr := adaptor.DoResponse(c, httpResp, relayInfo)
	if openaiErr != nil {
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}

	if usage.(*dto.Usage).TotalTokens == 0 {
		usage.(*dto.Usage).TotalTokens = imageRequest.N
	}
	if usage.(*dto.Usage).PromptTokens == 0 {
		usage.(*dto.Usage).PromptTokens = imageRequest.N
	}
	quality := "standard"
	if imageRequest.Quality == "hd" {
		quality = "hd"
	}

	logContent := fmt.Sprintf(lang.T(c, "image.log.size_quality"), imageRequest.Size, quality)
	postConsumeQuota(c, relayInfo, usage.(*dto.Usage), preConsumedQuota, userQuota, priceData, logContent)
	return nil
}
