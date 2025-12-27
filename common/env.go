package common

import (
	"fmt"
	"os"
	"strconv"
)

var ModerationEnabled bool
var ModerationBaseURL string
var ModerationKey string
var ModerationModel string
var ModerationTimeout int // ms

var ImageModerationEnabled bool

var AzureContentFilterEnabled bool
var AzureContentFilterEndpoint string
var AzureContentFilterKey string
var AzureContentFilterHarmLevel int      // -1: Zero tolerance, >=0: Allowed level
var AzureContentFilterImageHarmLevel int // Separate level for images

var DisableNormalLog bool
var SaveAllImages bool // If true, save all uploaded images to logs/rejected_images/

// https://docs.cohere.com/docs/safety-modes Type; NONE/CONTEXTUAL/STRICT

func GetEnvOrDefault(env string, defaultValue int) int {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	num, err := strconv.Atoi(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %d", env, err.Error(), defaultValue))
		return defaultValue
	}
	return num
}

func GetEnvOrDefaultString(env string, defaultValue string) string {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	return os.Getenv(env)
}

func GetEnvOrDefaultBool(env string, defaultValue bool) bool {
	if env == "" || os.Getenv(env) == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(os.Getenv(env))
	if err != nil {
		SysError(fmt.Sprintf("failed to parse %s: %s, using default value: %t", env, err.Error(), defaultValue))
		return defaultValue
	}
	return b
}