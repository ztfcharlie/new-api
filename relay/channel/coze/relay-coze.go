package coze

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"strings"

	"github.com/gin-gonic/gin"
)

func convertCozeChatRequest(c *gin.Context, request dto.GeneralOpenAIRequest) *CozeChatRequest {
	var messages []CozeEnterMessage
	// 将 request的messages的role为user的content转换为CozeMessage
	for _, message := range request.Messages {
		if message.Role == "user" {
			messages = append(messages, CozeEnterMessage{
				Role:    "user",
				Content: message.Content,
				// TODO: support more content type
				ContentType: "text",
			})
		}
	}
	user := request.User
	if user == "" {
		user = helper.GetResponseID(c)
	}
	cozeRequest := &CozeChatRequest{
		BotId:              c.GetString("bot_id"),
		UserId:             user,
		AdditionalMessages: messages,
		Stream:             request.Stream,
	}
	return cozeRequest
}

func cozeChatHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return service.OpenAIErrorWrapperLocal(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// convert coze response to openai response
	var response dto.TextResponse
	var cozeResponse CozeChatDetailResponse
	response.Model = info.UpstreamModelName
	err = json.Unmarshal(responseBody, &cozeResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if cozeResponse.Code != 0 {
		return service.OpenAIErrorWrapper(errors.New(cozeResponse.Msg), fmt.Sprintf("%d", cozeResponse.Code), http.StatusInternalServerError), nil
	}
	// 从上下文获取 usage
	var usage dto.Usage
	usage.PromptTokens = c.GetInt("coze_input_count")
	usage.CompletionTokens = c.GetInt("coze_output_count")
	usage.TotalTokens = c.GetInt("coze_token_count")
	response.Usage = usage
	response.Id = helper.GetResponseID(c)

	var responseContent json.RawMessage
	for _, data := range cozeResponse.Data {
		if data.Type == "answer" {
			responseContent = data.Content
			response.Created = data.CreatedAt
		}
	}
	// 添加 response.Choices
	response.Choices = []dto.OpenAITextResponseChoice{
		{
			Index:        0,
			Message:      dto.Message{Role: "assistant", Content: responseContent},
			FinishReason: "stop",
		},
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, &usage
}

func cozeChatStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	helper.SetEventStreamHeaders(c)
	id := helper.GetResponseID(c)
	var responseText string

	var currentEvent string
	var currentData string
	var usage dto.Usage

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if currentEvent != "" && currentData != "" {
				// handle last event
				handleCozeEvent(c, currentEvent, currentData, &responseText, &usage, id, info)
				currentEvent = ""
				currentData = ""
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(line[6:])
			continue
		}

		if strings.HasPrefix(line, "data:") {
			currentData = strings.TrimSpace(line[5:])
			continue
		}
	}

	// Last event
	if currentEvent != "" && currentData != "" {
		handleCozeEvent(c, currentEvent, currentData, &responseText, &usage, id, info)
	}

	if err := scanner.Err(); err != nil {
		return service.OpenAIErrorWrapper(err, "stream_scanner_error", http.StatusInternalServerError), nil
	}
	helper.Done(c)

	if usage.TotalTokens == 0 {
		usage.PromptTokens = info.PromptTokens
		usage.CompletionTokens, _ = service.CountTextToken("gpt-3.5-turbo", responseText)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return nil, &usage
}

func handleCozeEvent(c *gin.Context, event string, data string, responseText *string, usage *dto.Usage, id string, info *relaycommon.RelayInfo) {
	switch event {
	case "conversation.chat.completed":
		// 将 data 解析为 CozeChatResponseData
		var chatData CozeChatResponseData
		err := json.Unmarshal([]byte(data), &chatData)
		if err != nil {
			common.SysError("error_unmarshalling_stream_response: " + err.Error())
			return
		}

		usage.PromptTokens = chatData.Usage.InputCount
		usage.CompletionTokens = chatData.Usage.OutputCount
		usage.TotalTokens = chatData.Usage.TokenCount

		finishReason := "stop"
		stopResponse := helper.GenerateStopResponse(id, common.GetTimestamp(), info.UpstreamModelName, finishReason)
		helper.ObjectData(c, stopResponse)

	case "conversation.message.delta":
		// 将 data 解析为 CozeChatV3MessageDetail
		var messageData CozeChatV3MessageDetail
		err := json.Unmarshal([]byte(data), &messageData)
		if err != nil {
			common.SysError("error_unmarshalling_stream_response: " + err.Error())
			return
		}

		var content string
		err = json.Unmarshal(messageData.Content, &content)
		if err != nil {
			common.SysError("error_unmarshalling_stream_response: " + err.Error())
			return
		}

		*responseText += content

		openaiResponse := dto.ChatCompletionsStreamResponse{
			Id:      id,
			Object:  "chat.completion.chunk",
			Created: common.GetTimestamp(),
			Model:   info.UpstreamModelName,
		}

		choice := dto.ChatCompletionsStreamResponseChoice{
			Index: 0,
		}
		choice.Delta.SetContentString(content)
		openaiResponse.Choices = append(openaiResponse.Choices, choice)

		helper.ObjectData(c, openaiResponse)

	case "error":
		var errorData CozeError
		err := json.Unmarshal([]byte(data), &errorData)
		if err != nil {
			common.SysError("error_unmarshalling_stream_response: " + err.Error())
			return
		}

		common.SysError(fmt.Sprintf("stream event error: ", errorData.Code, errorData.Message))
	}
}

func checkIfChatComplete(a *Adaptor, c *gin.Context, info *relaycommon.RelayInfo) (error, bool) {
	requestURL := fmt.Sprintf("%s/v3/chat/retrieve", info.BaseUrl)

	requestURL = requestURL + "?conversation_id=" + c.GetString("coze_conversation_id") + "&chat_id=" + c.GetString("coze_chat_id")
	// 将 conversationId和chatId作为参数发送get请求
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return err, false
	}
	err = a.SetupRequestHeader(c, &req.Header, info)
	if err != nil {
		return err, false
	}

	resp, err := doRequest(req, info) // 调用 doRequest
	if err != nil {
		return err, false
	}
	if resp == nil { // 确保在 doRequest 失败时 resp 不为 nil 导致 panic
		return fmt.Errorf("resp is nil"), false
	}
	defer resp.Body.Close() // 确保响应体被关闭

	// 解析 resp 到 CozeChatResponse
	var cozeResponse CozeChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err), false
	}
	err = json.Unmarshal(responseBody, &cozeResponse)
	if err != nil {
		return fmt.Errorf("unmarshal response body failed: %w", err), false
	}
	if cozeResponse.Data.Status == "completed" {
		// 在上下文设置 usage
		c.Set("coze_token_count", cozeResponse.Data.Usage.TokenCount)
		c.Set("coze_output_count", cozeResponse.Data.Usage.OutputCount)
		c.Set("coze_input_count", cozeResponse.Data.Usage.InputCount)
		return nil, true
	} else if cozeResponse.Data.Status == "failed" || cozeResponse.Data.Status == "canceled" || cozeResponse.Data.Status == "requires_action" {
		return fmt.Errorf("chat status: %s", cozeResponse.Data.Status), false
	} else {
		return nil, false
	}
}

func getChatDetail(a *Adaptor, c *gin.Context, info *relaycommon.RelayInfo) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s/v3/chat/message/list", info.BaseUrl)

	requestURL = requestURL + "?conversation_id=" + c.GetString("coze_conversation_id") + "&chat_id=" + c.GetString("coze_chat_id")
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request failed: %w", err)
	}
	err = a.SetupRequestHeader(c, &req.Header, info)
	if err != nil {
		return nil, fmt.Errorf("setup request header failed: %w", err)
	}
	resp, err := doRequest(req, info)
	if err != nil {
		return nil, fmt.Errorf("do request failed: %w", err)
	}
	return resp, nil
}

func doRequest(req *http.Request, info *relaycommon.RelayInfo) (*http.Response, error) {
	var client *http.Client
	var err error // 声明 err 变量
	if proxyURL, ok := info.ChannelSetting["proxy"]; ok {
		client, err = service.NewProxyHttpClient(proxyURL.(string))
		if err != nil {
			return nil, fmt.Errorf("new proxy http client failed: %w", err)
		}
	} else {
		client = service.GetHttpClient()
	}
	resp, err := client.Do(req)
	if err != nil { // 增加对 client.Do(req) 返回错误的检查
		return nil, fmt.Errorf("client.Do failed: %w", err)
	}
	// _ = resp.Body.Close()
	return resp, nil
}
