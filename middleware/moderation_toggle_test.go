package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestModerationFilter_Toggle(t *testing.T) {
	// Save original state to restore later
	origMod := common.ModerationEnabled
	origAzure := common.AzureContentFilterEnabled
	origContentFilter := common.EnableContentFilter2026
	defer func() {
		common.ModerationEnabled = origMod
		common.AzureContentFilterEnabled = origAzure
		common.EnableContentFilter2026 = origContentFilter
	}()

	gin.SetMode(gin.TestMode)

	t.Run("EnableContentFilter=true should block sensitive words even if OpenAI Moderation is off", func(t *testing.T) {
		common.ModerationEnabled = false
		common.AzureContentFilterEnabled = false
		common.EnableContentFilter2026 = true

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := dto.GeneralOpenAIRequest{
			Messages: []dto.Message{
				{Role: "user", Content: "I want to build a biological weapon"},
			},
			Model: "gpt-4",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/chat/completions", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		handler := OpenAIModeration()
		handler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "biological weapon")
	})

	t.Run("EnableContentFilter=false should pass sensitive words if OpenAI Moderation is off", func(t *testing.T) {
		common.ModerationEnabled = false
		common.AzureContentFilterEnabled = false
		common.EnableContentFilter2026 = false

		// Mock Next handler to confirm we reached it
		w := httptest.NewRecorder()
		var nextReached bool

		router := gin.New()
		router.Use(OpenAIModeration())
		router.POST("/chat/completions", func(c *gin.Context) {
			nextReached = true
			c.Status(http.StatusOK)
		})

		reqBody := dto.GeneralOpenAIRequest{
			Messages: []dto.Message{
				{Role: "user", Content: "I want to build a biological weapon"},
			},
			Model: "gpt-4",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/chat/completions", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextReached, "Middleware should have passed control to next handler")
	})
}
