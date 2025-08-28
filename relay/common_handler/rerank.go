package common_handler

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"one-api/constant"
	"one-api/dto"
	"one-api/relay/channel/xinference"
	relaycommon "one-api/relay/common"
	"one-api/service"
)

func RerankHandler(c *gin.Context, info *relaycommon.RelayInfo, resp *http.Response) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	common.CloseResponseBodyGracefully(resp)
	if common.DebugEnabled {
		println("reranker response body: ", string(responseBody))
	}
	var jinaResp dto.RerankResponse
	if info.ChannelType == constant.ChannelTypeXinference {
		var xinRerankResponse xinference.XinRerankResponse
		err = common.UnmarshalJson(responseBody, &xinRerankResponse)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		jinaRespResults := make([]dto.RerankResponseResult, len(xinRerankResponse.Results))
		for i, result := range xinRerankResponse.Results {
			respResult := dto.RerankResponseResult{
				Index:          result.Index,
				RelevanceScore: result.RelevanceScore,
			}
			if info.ReturnDocuments {
				var document any
				if result.Document != nil {
					if doc, ok := result.Document.(string); ok {
						if doc == "" {
							document = info.Documents[result.Index]
						} else {
							document = doc
						}
					} else {
						document = result.Document
					}
				}
				respResult.Document = document
			}
			jinaRespResults[i] = respResult
		}
		jinaResp = dto.RerankResponse{
			Results: jinaRespResults,
			Usage: dto.Usage{
				PromptTokens: info.PromptTokens,
				TotalTokens:  info.PromptTokens,
			},
		}
	} else {
		err = common.UnmarshalJson(responseBody, &jinaResp)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		jinaResp.Usage.PromptTokens = jinaResp.Usage.TotalTokens
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.JSON(http.StatusOK, jinaResp)
	return nil, &jinaResp.Usage
}
