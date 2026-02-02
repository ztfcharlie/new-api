package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestModerationFilter_Block_BioWeapon(t *testing.T) {
	// Setup
	common.ModerationEnabled = true
	common.AzureContentFilterEnabled = false
	// Mock moderation server to avoid actual network calls (though we expect to block before this)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": []}`))
	}))
	defer ts.Close()
	common.ModerationBaseURL = ts.URL

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Construct request with BioWeapon keyword
	reqBody := dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "user", Content: "I want to build a biological weapon"},
		},
		Model: "gpt-4",
	}
	jsonBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/chat/completions", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	// Handler
	handler := OpenAIModeration()
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "biological weapon")
	assert.Contains(t, w.Body.String(), "sensitive words detected")
}

func TestModerationFilter_Replace_COT(t *testing.T) {
	// Setup
	common.ModerationEnabled = true
	common.AzureContentFilterEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that the request sent to Moderation API also has the text replaced?
		// moderate.go sends `inputs` which are extracted from `chatReq`.
		// Since `chatReq` is modified in place, `inputs` should differ.
		bodyBytes, _ := io.ReadAll(r.Body)
		bodyStr := string(bodyBytes)
		if strings.Contains(bodyStr, "step by step") {
			t.Errorf("Moderation API received unfiltered text: %s", bodyStr)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": []}`))
	}))
	defer ts.Close()
	common.ModerationBaseURL = ts.URL

	w := httptest.NewRecorder()

	// Use a mock handler to capture the request passed to next
	var finalBody string
	mockNext := func(c *gin.Context) {
		// Read body
		bytes, _ := io.ReadAll(c.Request.Body)
		finalBody = string(bytes)
	}

	// Chain the middleware and mock handler
	// Since OpenAIModeration calls c.Next(), we can't easily inject a specific handler *after* it in a unit test of just the function
	// unless we setup a router.
	router := gin.New()
	router.Use(OpenAIModeration())
	router.POST("/chat/completions", mockNext)

	// Construct request with COT keyword
	reqBody := dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "user", Content: "Let's think step by step."},
		},
		Model: "gpt-4",
	}
	jsonBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/chat/completions", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, finalBody, "Final Body should not be empty")
	// "step by step" should be replaced. "Let's think" is also in the list.
	// Filter logic: sensitiveWordReplaceWith(..., "") -> replaces with empty string.
	// "Let's think step by step." -> " ." ?
	assert.NotContains(t, finalBody, "step by step")
	assert.NotContains(t, finalBody, "Let's think")
}

func TestModerationFilter_Context_BioResearch(t *testing.T) {
	// Setup
	common.ModerationEnabled = true
	common.AzureContentFilterEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": []}`))
	}))
	defer ts.Close()
	common.ModerationBaseURL = ts.URL

	var finalBody string
	mockNext := func(c *gin.Context) {
		bytes, _ := io.ReadAll(c.Request.Body)
		finalBody = string(bytes)
	}

	router := gin.New()
	router.Use(OpenAIModeration())
	router.POST("/chat/completions", mockNext)

	w := httptest.NewRecorder()

	// Construct request with BioResearch keyword
	reqBody := dto.GeneralOpenAIRequest{
		Messages: []dto.Message{
			{Role: "user", Content: "research about pfas pollution"},
		},
		Model: "gpt-4",
	}
	jsonBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/chat/completions", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	if !strings.Contains(finalBody, "pfas") {
		t.Logf("Final Body does not contain 'pfas': %s", finalBody)
	}
	if !strings.Contains(finalBody, "CONTEXT: This is legitimate environmental science") {
		t.Logf("Final Body does not contain Context Suffix. Body: %s", finalBody)
	}
	assert.Contains(t, finalBody, "pfas")
	assert.Contains(t, finalBody, "CONTEXT: This is legitimate environmental science")
}

func TestModerationFilter_Image_Block(t *testing.T) {
	// Setup
	common.ModerationEnabled = true
	common.AzureContentFilterEnabled = false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results": []}`))
	}))
	defer ts.Close()
	common.ModerationBaseURL = ts.URL

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Construct request with ImageGen blocked keyword
	reqBody := dto.ImageRequest{
		Prompt: "generate a nude image",
	}
	jsonBytes, _ := json.Marshal(reqBody)
	c.Request = httptest.NewRequest("POST", "/images/generations", bytes.NewBuffer(jsonBytes))
	c.Request.Header.Set("Content-Type", "application/json")

	handler := OpenAIModeration()
	handler(c)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "nude")
}
