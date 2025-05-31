package claude

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	relaycommon "one-api/relay/common"
	"one-api/relay/helper"
	"one-api/service"
	"one-api/setting/model_setting"
	"strings"

	"github.com/gin-gonic/gin"
)

func stopReasonClaude2OpenAI(reason string) string {
	switch reason {
	case "stop_sequence":
		return "stop"
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "max_tokens"
	case "tool_use":
		return "tool_calls"
	default:
		return reason
	}
}

func RequestOpenAI2ClaudeComplete(textRequest dto.GeneralOpenAIRequest) *dto.ClaudeRequest {

	claudeRequest := dto.ClaudeRequest{
		Model:         textRequest.Model,
		Prompt:        "",
		StopSequences: nil,
		Temperature:   textRequest.Temperature,
		TopP:          textRequest.TopP,
		TopK:          textRequest.TopK,
		Stream:        textRequest.Stream,
	}
	if claudeRequest.MaxTokensToSample == 0 {
		claudeRequest.MaxTokensToSample = 4096
	}
	prompt := ""
	for _, message := range textRequest.Messages {
		if message.Role == "user" {
			prompt += fmt.Sprintf("\n\nHuman: %s", message.Content)
		} else if message.Role == "assistant" {
			prompt += fmt.Sprintf("\n\nAssistant: %s", message.Content)
		} else if message.Role == "system" {
			if prompt == "" {
				prompt = message.StringContent()
			}
		}
	}
	prompt += "\n\nAssistant:"
	claudeRequest.Prompt = prompt
	return &claudeRequest
}

func RequestOpenAI2ClaudeMessage(textRequest dto.GeneralOpenAIRequest) (*dto.ClaudeRequest, error) {
	claudeTools := make([]dto.Tool, 0, len(textRequest.Tools))

	for _, tool := range textRequest.Tools {
		if params, ok := tool.Function.Parameters.(map[string]any); ok {
			claudeTool := dto.Tool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
			}
			claudeTool.InputSchema = make(map[string]interface{})
			if params["type"] != nil {
				claudeTool.InputSchema["type"] = params["type"].(string)
			}
			claudeTool.InputSchema["properties"] = params["properties"]
			claudeTool.InputSchema["required"] = params["required"]
			for s, a := range params {
				if s == "type" || s == "properties" || s == "required" {
					continue
				}
				claudeTool.InputSchema[s] = a
			}
			claudeTools = append(claudeTools, claudeTool)
		}
	}

	claudeRequest := dto.ClaudeRequest{
		Model:         textRequest.Model,
		MaxTokens:     textRequest.MaxTokens,
		StopSequences: nil,
		Temperature:   textRequest.Temperature,
		TopP:          textRequest.TopP,
		TopK:          textRequest.TopK,
		Stream:        textRequest.Stream,
		Tools:         claudeTools,
	}

	if claudeRequest.MaxTokens == 0 {
		claudeRequest.MaxTokens = uint(model_setting.GetClaudeSettings().GetDefaultMaxTokens(textRequest.Model))
	}

	if model_setting.GetClaudeSettings().ThinkingAdapterEnabled &&
		strings.HasSuffix(textRequest.Model, "-thinking") {

		// 因为BudgetTokens 必须大于1024
		if claudeRequest.MaxTokens < 1280 {
			claudeRequest.MaxTokens = 1280
		}

		// BudgetTokens 为 max_tokens 的 80%
		claudeRequest.Thinking = &dto.Thinking{
			Type:         "enabled",
			BudgetTokens: int(float64(claudeRequest.MaxTokens) * model_setting.GetClaudeSettings().ThinkingAdapterBudgetTokensPercentage),
		}
		// TODO: 临时处理
		// https://docs.anthropic.com/en/docs/build-with-claude/extended-thinking#important-considerations-when-using-extended-thinking
		claudeRequest.TopP = 0
		claudeRequest.Temperature = common.GetPointer[float64](1.0)
		claudeRequest.Model = strings.TrimSuffix(textRequest.Model, "-thinking")
	}

	if textRequest.Stop != nil {
		// stop maybe string/array string, convert to array string
		switch textRequest.Stop.(type) {
		case string:
			claudeRequest.StopSequences = []string{textRequest.Stop.(string)}
		case []interface{}:
			stopSequences := make([]string, 0)
			for _, stop := range textRequest.Stop.([]interface{}) {
				stopSequences = append(stopSequences, stop.(string))
			}
			claudeRequest.StopSequences = stopSequences
		}
	}
	formatMessages := make([]dto.Message, 0)
	lastMessage := dto.Message{
		Role: "tool",
	}
	for i, message := range textRequest.Messages {
		if message.Role == "" {
			textRequest.Messages[i].Role = "user"
		}
		fmtMessage := dto.Message{
			Role:    message.Role,
			Content: message.Content,
		}
		if message.Role == "tool" {
			fmtMessage.ToolCallId = message.ToolCallId
		}
		if message.Role == "assistant" && message.ToolCalls != nil {
			fmtMessage.ToolCalls = message.ToolCalls
		}
		if lastMessage.Role == message.Role && lastMessage.Role != "tool" {
			if lastMessage.IsStringContent() && message.IsStringContent() {
				content, _ := json.Marshal(strings.Trim(fmt.Sprintf("%s %s", lastMessage.StringContent(), message.StringContent()), "\""))
				fmtMessage.Content = content
				// delete last message
				formatMessages = formatMessages[:len(formatMessages)-1]
			}
		}
		if fmtMessage.Content == nil {
			content, _ := json.Marshal("...")
			fmtMessage.Content = content
		}
		formatMessages = append(formatMessages, fmtMessage)
		lastMessage = fmtMessage
	}

	claudeMessages := make([]dto.ClaudeMessage, 0)
	isFirstMessage := true
	for _, message := range formatMessages {
		if message.Role == "system" {
			if message.IsStringContent() {
				claudeRequest.System = message.StringContent()
			} else {
				contents := message.ParseContent()
				content := ""
				for _, ctx := range contents {
					if ctx.Type == "text" {
						content += ctx.Text
					}
				}
				claudeRequest.System = content
			}
		} else {
			if isFirstMessage {
				isFirstMessage = false
				if message.Role != "user" {
					// fix: first message is assistant, add user message
					claudeMessage := dto.ClaudeMessage{
						Role: "user",
						Content: []dto.ClaudeMediaMessage{
							{
								Type: "text",
								Text: common.GetPointer[string]("..."),
							},
						},
					}
					claudeMessages = append(claudeMessages, claudeMessage)
				}
			}
			claudeMessage := dto.ClaudeMessage{
				Role: message.Role,
			}
			if message.Role == "tool" {
				if len(claudeMessages) > 0 && claudeMessages[len(claudeMessages)-1].Role == "user" {
					lastMessage := claudeMessages[len(claudeMessages)-1]
					if content, ok := lastMessage.Content.(string); ok {
						lastMessage.Content = []dto.ClaudeMediaMessage{
							{
								Type: "text",
								Text: common.GetPointer[string](content),
							},
						}
					}
					lastMessage.Content = append(lastMessage.Content.([]dto.ClaudeMediaMessage), dto.ClaudeMediaMessage{
						Type:      "tool_result",
						ToolUseId: message.ToolCallId,
						Content:   message.Content,
					})
					claudeMessages[len(claudeMessages)-1] = lastMessage
					continue
				} else {
					claudeMessage.Role = "user"
					claudeMessage.Content = []dto.ClaudeMediaMessage{
						{
							Type:      "tool_result",
							ToolUseId: message.ToolCallId,
							Content:   message.Content,
						},
					}
				}
			} else if message.IsStringContent() && message.ToolCalls == nil {
				claudeMessage.Content = message.StringContent()
			} else {
				claudeMediaMessages := make([]dto.ClaudeMediaMessage, 0)
				for _, mediaMessage := range message.ParseContent() {
					claudeMediaMessage := dto.ClaudeMediaMessage{
						Type: mediaMessage.Type,
					}
					if mediaMessage.Type == "text" {
						claudeMediaMessage.Text = common.GetPointer[string](mediaMessage.Text)
					} else {
						imageUrl := mediaMessage.GetImageMedia()
						claudeMediaMessage.Type = "image"
						claudeMediaMessage.Source = &dto.ClaudeMessageSource{
							Type: "base64",
						}
						// 判断是否是url
						if strings.HasPrefix(imageUrl.Url, "http") {
							// 是url，获取图片的类型和base64编码的数据
							fileData, err := service.GetFileBase64FromUrl(imageUrl.Url)
							if err != nil {
								return nil, fmt.Errorf("get file base64 from url failed: %s", err.Error())
							}
							claudeMediaMessage.Source.MediaType = fileData.MimeType
							claudeMediaMessage.Source.Data = fileData.Base64Data
						} else {
							_, format, base64String, err := service.DecodeBase64ImageData(imageUrl.Url)
							if err != nil {
								return nil, err
							}
							claudeMediaMessage.Source.MediaType = "image/" + format
							claudeMediaMessage.Source.Data = base64String
						}
					}
					claudeMediaMessages = append(claudeMediaMessages, claudeMediaMessage)
				}
				if message.ToolCalls != nil {
					for _, toolCall := range message.ParseToolCalls() {
						inputObj := make(map[string]any)
						if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &inputObj); err != nil {
							common.SysError("tool call function arguments is not a map[string]any: " + fmt.Sprintf("%v", toolCall.Function.Arguments))
							continue
						}
						claudeMediaMessages = append(claudeMediaMessages, dto.ClaudeMediaMessage{
							Type:  "tool_use",
							Id:    toolCall.ID,
							Name:  toolCall.Function.Name,
							Input: inputObj,
						})
					}
				}
				claudeMessage.Content = claudeMediaMessages
			}
			claudeMessages = append(claudeMessages, claudeMessage)
		}
	}
	claudeRequest.Prompt = ""
	claudeRequest.Messages = claudeMessages
	return &claudeRequest, nil
}
func StreamResponseClaude2OpenAI(reqMode int, claudeResponse *dto.ClaudeResponse) *dto.ChatCompletionsStreamResponse {
	// 创建一个新的 ChatCompletionsStreamResponse 结构体实例
	var response dto.ChatCompletionsStreamResponse
	// 设置响应对象类型为 "chat.completion.chunk"
	response.Object = "chat.completion.chunk"
	// 设置模型名称
	response.Model = claudeResponse.Model
	// 初始化 Choices 列表
	response.Choices = make([]dto.ChatCompletionsStreamResponseChoice, 0)
	// 初始化工具调用列表
	tools := make([]dto.ToolCallResponse, 0)
	fcIdx := 0
	if claudeResponse.Index != nil {
		fcIdx = *claudeResponse.Index - 1
		if fcIdx < 0 {
			fcIdx = 0
		}
	}
	var choice dto.ChatCompletionsStreamResponseChoice

	// 如果请求模式是完成模式
	if reqMode == RequestModeCompletion {
		// 设置内容字符串为 Claude 响应的 Completion 字段
		choice.Delta.SetContentString(claudeResponse.Completion)
		// 获取并转换停止原因
		finishReason := stopReasonClaude2OpenAI(claudeResponse.StopReason)
		// 如果停止原因不为 "null"，则设置 FinishReason
		if finishReason != "null" {
			choice.FinishReason = &finishReason
		}
	} else {
		// 根据 Claude 响应类型进行处理
		if claudeResponse.Type == "message_start" {
			// 设置响应 ID 和模型名称
			response.Id = claudeResponse.Message.Id
			response.Model = claudeResponse.Message.Model
			// 设置角色为 "assistant"
			choice.Delta.SetContentString("")
			choice.Delta.Role = "assistant"
		} else if claudeResponse.Type == "content_block_start" {
			// 检查 ContentBlock 是否存在
			if claudeResponse.ContentBlock != nil {
				// 如果 ContentBlock 类型是 "tool_use"，则添加到工具调用列表
				if claudeResponse.ContentBlock.Type == "tool_use" {
					tools = append(tools, dto.ToolCallResponse{
						Index: common.GetPointer(fcIdx),
						ID:    claudeResponse.ContentBlock.Id,
						Type:  "function",
						Function: dto.FunctionResponse{
							Name:      claudeResponse.ContentBlock.Name,
							Arguments: "",
						},
					})
				}
			} else {
				// 如果 ContentBlock 不存在，返回 nil
				return nil
			}
		} else if claudeResponse.Type == "content_block_delta" {
			// 检查 Delta 是否存在
			if claudeResponse.Delta != nil {
				// 设置索引
				choice.Index = *claudeResponse.Index
				// 设置 Delta 内容
				choice.Delta.Content = claudeResponse.Delta.Text
				// 根据 Delta 类型进行处理
				switch claudeResponse.Delta.Type {
				case "input_json_delta":
					// 添加工具调用响应
					tools = append(tools, dto.ToolCallResponse{
						Type:  "function",
						Index: common.GetPointer(fcIdx),
						Function: dto.FunctionResponse{
							Arguments: *claudeResponse.Delta.PartialJson,
						},
					})
				case "signature_delta":
					// 对加密内容不处理，设置空行
					signatureContent := "\n"
					choice.Delta.ReasoningContent = &signatureContent
				case "thinking_delta":
					// 设置思考内容
					thinkingContent := claudeResponse.Delta.Thinking
					choice.Delta.ReasoningContent = &thinkingContent
				}
			}
		} else if claudeResponse.Type == "message_delta" {
			// 获取并转换停止原因
			finishReason := stopReasonClaude2OpenAI(*claudeResponse.Delta.StopReason)
			// 如果停止原因不为 "null"，则设置 FinishReason
			if finishReason != "null" {
				choice.FinishReason = &finishReason
			}
		} else if claudeResponse.Type == "message_stop" {
			// 如果是 message_stop 类型，返回 nil
			return nil
		} else {
			// 如果是其他类型，返回 nil
			return nil
		}
	}

	// 如果有工具调用，设置工具调用列表
	if len(tools) > 0 {
		choice.Delta.Content = nil // 兼容其他 OpenAI 派生应用
		choice.Delta.ToolCalls = tools
	}
	// 将 choice 添加到 Choices 列表
	response.Choices = append(response.Choices, choice)

	// 返回构建的响应
	return &response
}

func ResponseClaude2OpenAI(reqMode int, claudeResponse *dto.ClaudeResponse) *dto.OpenAITextResponse {
	choices := make([]dto.OpenAITextResponseChoice, 0)
	fullTextResponse := dto.OpenAITextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Object:  "chat.completion",
		Created: common.GetTimestamp(),
	}
	var responseText string
	var responseThinking string
	if len(claudeResponse.Content) > 0 {
		responseText = claudeResponse.Content[0].GetText()
		responseThinking = claudeResponse.Content[0].Thinking
	}
	tools := make([]dto.ToolCallResponse, 0)
	thinkingContent := ""

	if reqMode == RequestModeCompletion {
		content, _ := json.Marshal(strings.TrimPrefix(claudeResponse.Completion, " "))
		choice := dto.OpenAITextResponseChoice{
			Index: 0,
			Message: dto.Message{
				Role:    "assistant",
				Content: content,
				Name:    nil,
			},
			FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
		}
		choices = append(choices, choice)
	} else {
		fullTextResponse.Id = claudeResponse.Id
		for _, message := range claudeResponse.Content {
			switch message.Type {
			case "tool_use":
				args, _ := json.Marshal(message.Input)
				tools = append(tools, dto.ToolCallResponse{
					ID:   message.Id,
					Type: "function", // compatible with other OpenAI derivative applications
					Function: dto.FunctionResponse{
						Name:      message.Name,
						Arguments: string(args),
					},
				})
			case "thinking":
				// 加密的不管， 只输出明文的推理过程
				thinkingContent = message.Thinking
			case "text":
				responseText = message.GetText()
			}
		}
	}
	choice := dto.OpenAITextResponseChoice{
		Index: 0,
		Message: dto.Message{
			Role: "assistant",
		},
		FinishReason: stopReasonClaude2OpenAI(claudeResponse.StopReason),
	}
	choice.SetStringContent(responseText)
	if len(responseThinking) > 0 {
		choice.ReasoningContent = responseThinking
	}
	if len(tools) > 0 {
		choice.Message.SetToolCalls(tools)
	}
	choice.Message.ReasoningContent = thinkingContent
	fullTextResponse.Model = claudeResponse.Model
	choices = append(choices, choice)
	fullTextResponse.Choices = choices
	return &fullTextResponse
}

type ClaudeResponseInfo struct {
	ResponseId   string
	Created      int64
	Model        string
	ResponseText strings.Builder
	Usage        *dto.Usage
}

func FormatClaudeResponseInfo(requestMode int, claudeResponse *dto.ClaudeResponse, oaiResponse *dto.ChatCompletionsStreamResponse, claudeInfo *ClaudeResponseInfo) bool {
	if requestMode == RequestModeCompletion {
		claudeInfo.ResponseText.WriteString(claudeResponse.Completion)
	} else {
		if claudeResponse.Type == "message_start" {
			// message_start, 获取usage
			claudeInfo.ResponseId = claudeResponse.Message.Id
			claudeInfo.Model = claudeResponse.Message.Model
			claudeInfo.Usage.PromptTokens = claudeResponse.Message.Usage.InputTokens
		} else if claudeResponse.Type == "content_block_delta" {
			if claudeResponse.Delta.Text != nil {
				claudeInfo.ResponseText.WriteString(*claudeResponse.Delta.Text)
			}
		} else if claudeResponse.Type == "message_delta" {
			claudeInfo.Usage.CompletionTokens = claudeResponse.Usage.OutputTokens
			if claudeResponse.Usage.InputTokens > 0 {
				claudeInfo.Usage.PromptTokens = claudeResponse.Usage.InputTokens
			}
			claudeInfo.Usage.TotalTokens = claudeInfo.Usage.PromptTokens + claudeResponse.Usage.OutputTokens
		} else if claudeResponse.Type == "content_block_start" {
		} else {
			return false
		}
	}
	if oaiResponse != nil {
		oaiResponse.Id = claudeInfo.ResponseId
		oaiResponse.Created = claudeInfo.Created
		oaiResponse.Model = claudeInfo.Model
	}
	return true
}
func HandleStreamResponseData(c *gin.Context, info *relaycommon.RelayInfo, claudeInfo *ClaudeResponseInfo, data string, requestMode int) *dto.OpenAIErrorWithStatusCode {
	// 声明一个 ClaudeResponse 结构体实例，用于存放解码后的数据
	var claudeResponse dto.ClaudeResponse

	// 将 JSON 字符串解码到 claudeResponse 中
	err := common.DecodeJsonStr(data, &claudeResponse)
	if err != nil {
		// 解码失败时记录系统错误并返回错误信息
		common.SysError("error unmarshalling stream response: " + err.Error())
		return service.OpenAIErrorWrapper(err, "stream_response_error", http.StatusInternalServerError)
	}

	// 检查解码后的响应是否包含错误信息
	if claudeResponse.Error != nil && claudeResponse.Error.Type != "" {
		// 返回包含错误类型和消息的错误响应
		common.LogInfo(c, fmt.Sprintf("ClaudeResponse contains error: %s", claudeResponse.Error.Message))
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Code:    "stream_response_error",
				Type:    claudeResponse.Error.Type,
				Message: claudeResponse.Error.Message,
			},
			StatusCode: http.StatusInternalServerError,
		}
	}

	// 根据 RelayFormat 判断是处理 Claude 格式还是 OpenAI 格式
	if info.RelayFormat == relaycommon.RelayFormatClaude {
		//common.LogInfo(c, "Processing Claude format")

		// 如果请求模式是完成模式
		if requestMode == RequestModeCompletion {
			// 将响应中的 Completion 字段追加到 ResponseText
			claudeInfo.ResponseText.WriteString(claudeResponse.Completion)
			//common.LogInfo(c, fmt.Sprintf("Appended completion: %s", claudeResponse.Completion))
		} else {
			// 处理不同类型的响应数据
			if claudeResponse.Type == "message_start" {
				// 如果是 message_start，更新模型名称和使用情况
				info.UpstreamModelName = claudeResponse.Message.Model
				claudeInfo.Usage.PromptTokens = claudeResponse.Message.Usage.InputTokens
				claudeInfo.Usage.PromptTokensDetails.CachedTokens = claudeResponse.Message.Usage.CacheReadInputTokens
				claudeInfo.Usage.PromptTokensDetails.CachedCreationTokens = claudeResponse.Message.Usage.CacheCreationInputTokens
				claudeInfo.Usage.CompletionTokens = claudeResponse.Message.Usage.OutputTokens
				//common.LogInfo(c, fmt.Sprintf("Processed message_start: Model=%s, PromptTokens=%d", info.UpstreamModelName, claudeInfo.Usage.PromptTokens))
			} else if claudeResponse.Type == "content_block_delta" {
				// 如果是 content_block_delta，将 Delta 中的文本追加到 ResponseText
				claudeInfo.ResponseText.WriteString(claudeResponse.Delta.GetText())
				//common.LogInfo(c, fmt.Sprintf("Appended content_block_delta: %s", claudeResponse.Delta.GetText()))
			} else if claudeResponse.Type == "message_delta" {
				// 如果是 message_delta，更新使用情况
				if claudeResponse.Usage.InputTokens > 0 {
					// 不叠加，只取最新的输入令牌数
					claudeInfo.Usage.PromptTokens = claudeResponse.Usage.InputTokens
				}
				claudeInfo.Usage.CompletionTokens = claudeResponse.Usage.OutputTokens
				claudeInfo.Usage.TotalTokens = claudeInfo.Usage.PromptTokens + claudeInfo.Usage.CompletionTokens
				//common.LogInfo(c, fmt.Sprintf("Processed message_delta: PromptTokens=%d, CompletionTokens=%d", claudeInfo.Usage.PromptTokens, claudeInfo.Usage.CompletionTokens))
			}
		}
		// 处理 Claude 数据块
		//common.LogInfo(c, fmt.Sprintf("Sending to ClaudeChunkData: Type=%s, Data=%s", claudeResponse.Type, data))
		helper.ClaudeChunkData(c, claudeResponse, data)
	} else if info.RelayFormat == relaycommon.RelayFormatOpenAI {
		//common.LogInfo(c, "Processing OpenAI format")

		// 将 Claude 格式的响应转换为 OpenAI 格式
		response := StreamResponseClaude2OpenAI(requestMode, &claudeResponse)

		// 格式化 Claude 响应信息并检查是否成功
		if !FormatClaudeResponseInfo(requestMode, &claudeResponse, response, claudeInfo) {
			//common.LogInfo(c, "Failed to format ClaudeResponseInfo")
			return nil
		}

		// 发送转换后的响应数据
		err = helper.ObjectData(c, response)
		if err != nil {
			// 记录发送失败的错误
			common.LogError(c, "send_stream_response_failed: "+err.Error())
		}
	}
	// 返回 nil 表示成功处理
	//common.LogInfo(c, "HandleStreamResponseData completed successfully")
	return nil
}

func HandleStreamFinalResponse(c *gin.Context, info *relaycommon.RelayInfo, claudeInfo *ClaudeResponseInfo, requestMode int) {
	if info.RelayFormat == relaycommon.RelayFormatClaude {
		if requestMode == RequestModeCompletion {
			claudeInfo.Usage, _ = service.ResponseText2Usage(claudeInfo.ResponseText.String(), info.UpstreamModelName, info.PromptTokens)
		} else {
			// 说明流模式建立失败，可能为官方出错
			if claudeInfo.Usage.PromptTokens == 0 {
				//usage.PromptTokens = info.PromptTokens
			}
			if claudeInfo.Usage.CompletionTokens == 0 {
				claudeInfo.Usage, _ = service.ResponseText2Usage(claudeInfo.ResponseText.String(), info.UpstreamModelName, claudeInfo.Usage.PromptTokens)
			}
		}
	} else if info.RelayFormat == relaycommon.RelayFormatOpenAI {
		if requestMode == RequestModeCompletion {
			claudeInfo.Usage, _ = service.ResponseText2Usage(claudeInfo.ResponseText.String(), info.UpstreamModelName, info.PromptTokens)
		} else {
			if claudeInfo.Usage.PromptTokens == 0 {
				//上游出错
			}
			if claudeInfo.Usage.CompletionTokens == 0 {
				claudeInfo.Usage, _ = service.ResponseText2Usage(claudeInfo.ResponseText.String(), info.UpstreamModelName, claudeInfo.Usage.PromptTokens)
			}
		}
		if info.ShouldIncludeUsage {
			response := helper.GenerateFinalUsageResponse(claudeInfo.ResponseId, claudeInfo.Created, info.UpstreamModelName, *claudeInfo.Usage)
			err := helper.ObjectData(c, response)
			if err != nil {
				common.SysError("send final response failed: " + err.Error())
			}
		}
		helper.Done(c)
	}
}

func ClaudeStreamHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, requestMode int) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	claudeInfo := &ClaudeResponseInfo{
		ResponseId:   fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Created:      common.GetTimestamp(),
		Model:        info.UpstreamModelName,
		ResponseText: strings.Builder{},
		Usage:        &dto.Usage{},
	}
	var err *dto.OpenAIErrorWithStatusCode
	helper.StreamScannerHandler(c, resp, info, func(data string) bool {
		err = HandleStreamResponseData(c, info, claudeInfo, data, requestMode)
		if err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return err, nil
	}

	HandleStreamFinalResponse(c, info, claudeInfo, requestMode)
	return nil, claudeInfo.Usage
}

func HandleClaudeResponseData(c *gin.Context, info *relaycommon.RelayInfo, claudeInfo *ClaudeResponseInfo, data []byte, requestMode int) *dto.OpenAIErrorWithStatusCode {
	var claudeResponse dto.ClaudeResponse
	err := common.DecodeJson(data, &claudeResponse)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "unmarshal_claude_response_failed", http.StatusInternalServerError)
	}
	if claudeResponse.Error != nil && claudeResponse.Error.Type != "" {
		return &dto.OpenAIErrorWithStatusCode{
			Error: dto.OpenAIError{
				Message: claudeResponse.Error.Message,
				Type:    claudeResponse.Error.Type,
				Code:    claudeResponse.Error.Type,
			},
			StatusCode: http.StatusInternalServerError,
		}
	}
	if requestMode == RequestModeCompletion {
		completionTokens, err := service.CountTextToken(claudeResponse.Completion, info.OriginModelName)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "count_token_text_failed", http.StatusInternalServerError)
		}
		claudeInfo.Usage.PromptTokens = info.PromptTokens
		claudeInfo.Usage.CompletionTokens = completionTokens
		claudeInfo.Usage.TotalTokens = info.PromptTokens + completionTokens
	} else {
		claudeInfo.Usage.PromptTokens = claudeResponse.Usage.InputTokens
		claudeInfo.Usage.CompletionTokens = claudeResponse.Usage.OutputTokens
		claudeInfo.Usage.TotalTokens = claudeResponse.Usage.InputTokens + claudeResponse.Usage.OutputTokens
		claudeInfo.Usage.PromptTokensDetails.CachedTokens = claudeResponse.Usage.CacheReadInputTokens
		claudeInfo.Usage.PromptTokensDetails.CachedCreationTokens = claudeResponse.Usage.CacheCreationInputTokens
	}
	var responseData []byte
	switch info.RelayFormat {
	case relaycommon.RelayFormatOpenAI:
		openaiResponse := ResponseClaude2OpenAI(requestMode, &claudeResponse)
		openaiResponse.Usage = *claudeInfo.Usage
		responseData, err = json.Marshal(openaiResponse)
		if err != nil {
			return service.OpenAIErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError)
		}
	case relaycommon.RelayFormatClaude:
		responseData = data
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(http.StatusOK)
	_, err = c.Writer.Write(responseData)
	return nil
}

func ClaudeHandler(c *gin.Context, resp *http.Response, requestMode int, info *relaycommon.RelayInfo) (*dto.OpenAIErrorWithStatusCode, *dto.Usage) {
	claudeInfo := &ClaudeResponseInfo{
		ResponseId:   fmt.Sprintf("chatcmpl-%s", common.GetUUID()),
		Created:      common.GetTimestamp(),
		Model:        info.UpstreamModelName,
		ResponseText: strings.Builder{},
		Usage:        &dto.Usage{},
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return service.OpenAIErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	resp.Body.Close()
	if common.DebugEnabled {
		println("responseBody: ", string(responseBody))
	}
	handleErr := HandleClaudeResponseData(c, info, claudeInfo, responseBody, requestMode)
	if handleErr != nil {
		return handleErr, nil
	}
	return nil, claudeInfo.Usage
}
