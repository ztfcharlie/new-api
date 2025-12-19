package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/gin-gonic/gin"
)

type ModerationInput struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageUrl *struct {
		Url string `json:"url"`
	} `json:"image_url,omitempty"`
}

type ModerationRequest struct {
	Model string            `json:"model"`
	Input []ModerationInput `json:"input"`
}

type ModerationResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Flagged    bool            `json:"flagged"`
		Categories map[string]bool `json:"categories"`
	} `json:"results"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func abortWithModerationError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "new_api_error",
			"param":   "",
			"code":    "sensitive_words_detected",
		},
	})
}

func OpenAIModeration() gin.HandlerFunc {
	return func(c *gin.Context) {
		// FORCE DEBUG LOG
		fmt.Printf("[Moderation-Debug] Middleware invoked. Path: %s, Enabled: %v\n", c.Request.URL.Path, common.ModerationEnabled)

		if !common.ModerationEnabled {
			return
		}

		// Only check Chat endpoints
		// We check for suffixes to match standard OpenAI paths
		path := c.Request.URL.Path
		isChat := strings.HasSuffix(path, "/chat/completions") || strings.HasSuffix(path, "/messages") // OpenAI & Claude

		if !isChat {
			fmt.Printf("[Moderation-Debug] Path not matched: %s\n", path)
			return
		}

		common.SysLog(fmt.Sprintf("Moderation middleware triggered for path: %s", path))

		var inputs []ModerationInput

		// Parse Request
		if isChat {
			var chatReq dto.GeneralOpenAIRequest
			if err := common.UnmarshalBodyReusable(c, &chatReq); err != nil {
				common.SysError(fmt.Sprintf("Moderation: failed to unmarshal body: %v", err))
				abortWithModerationError(c, http.StatusBadRequest, "Invalid request body for moderation")
				return
			}

			// Extract user content
			for _, msg := range chatReq.Messages {
				if msg.Role == "user" {
					contentList := msg.ParseContent()
					for _, content := range contentList {
						if content.Type == dto.ContentTypeText {
							inputs = append(inputs, ModerationInput{
								Type: "text",
								Text: content.Text,
							})
						}
					}
				}
			}
		}

		// If no inputs found (empty request?), just proceed
		if len(inputs) == 0 {
			common.SysLog("Moderation: no text inputs found to check")
			c.Next()
			return
		}

		common.SysLog(fmt.Sprintf("Moderation: checking %d inputs", len(inputs)))

		// Prepare Moderation Request
		modReq := ModerationRequest{
			Model: common.ModerationModel,
			Input: inputs,
		}

		reqBody, err := json.Marshal(modReq)
		if err != nil {
			abortWithModerationError(c, http.StatusInternalServerError, "Failed to build moderation request")
			return
		}

		// Call Moderation API
		// Prepare URL
		url := common.ModerationBaseURL
		if strings.HasSuffix(url, "/") {
			url += "v1/moderations"
		} else {
			if !strings.HasSuffix(url, "/v1/moderations") {
				url += "/v1/moderations"
			}
		}

		common.SysLog(fmt.Sprintf("Moderation: sending request to %s", url))

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			abortWithModerationError(c, http.StatusInternalServerError, "Failed to create moderation request")
			return
		}

		req.Header.Set("Content-Type", "application/json")
		if common.ModerationKey != "" {
			req.Header.Set("Authorization", "Bearer "+common.ModerationKey)
		}

		// Wait for moderation result as configured (default 1000ms)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(common.ModerationTimeout)*time.Millisecond)
		defer cancel()
		req = req.WithContext(ctx)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			common.SysError(fmt.Sprintf("Moderation API failed: %v", err))
			// Timed out or failed, reject request as requested
			abortWithModerationError(c, http.StatusServiceUnavailable, "Content moderation service timed out or failed")
			return
		}
		defer resp.Body.Close()

		common.SysLog(fmt.Sprintf("Moderation: API returned status %d", resp.StatusCode))

		if resp.StatusCode != http.StatusOK {
			common.SysError(fmt.Sprintf("Moderation API returned status: %d", resp.StatusCode))
			abortWithModerationError(c, http.StatusServiceUnavailable, fmt.Sprintf("Content moderation failed with status %d", resp.StatusCode))
			return
		}

		var modResp ModerationResponse
		if err := json.NewDecoder(resp.Body).Decode(&modResp); err != nil {
			abortWithModerationError(c, http.StatusInternalServerError, "Failed to parse moderation response")
			return
		}

		if modResp.Error != nil {
			abortWithModerationError(c, http.StatusBadRequest, "Moderation API Error: "+modResp.Error.Message)
			return
		}

		// Check results
		for _, res := range modResp.Results {
			if res.Flagged {
				// Build reason
				var reasons []string
				for cat, flagged := range res.Categories {
					if flagged {
						reasons = append(reasons, cat)
					}
				}
				common.SysLog(fmt.Sprintf("Moderation: content flagged! Reasons: %v", reasons))
				// Use standard sensitive word error message format
				abortWithModerationError(c, http.StatusBadRequest, fmt.Sprintf("敏感词检测失败: %s", strings.Join(reasons, ", ")))
				return
			}
		}
		
		common.SysLog("Moderation: content passed")
		c.Next()
	}
}
