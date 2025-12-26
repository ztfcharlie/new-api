package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
)

var (
	logLock sync.Mutex
)

// RecordModerationLog records the blocked moderation request to a log file.
func RecordModerationLog(c *gin.Context, prompt string, reason string, source string) {
	// Construct log directory using global LogDir setting
	// If common.LogDir is set (via flag), use it. Otherwise default to ./logs
	logDir := "./logs"
	if common.LogDir != nil && *common.LogDir != "" {
		logDir = *common.LogDir
	}
	
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		common.SysError(fmt.Sprintf("Failed to create log directory: %v", err))
		return
	}

	now := time.Now()
	dateStr := now.Format("2006-01-02")
	filename := filepath.Join(logDir, fmt.Sprintf("moderation-%s.log", dateStr))

	// Get User Info
	userID := c.GetInt("id")
	userName := c.GetString("username")
	ip := c.ClientIP()
	path := c.Request.URL.Path

	// Format Log Entry
	// [Time] [IP] [UserID(Name)] [Path] [Source] [Reason] Content
	logEntry := fmt.Sprintf("[%s] [IP:%s] [User:%d(%s)] [Path:%s] [Source:%s] [Reason:%s] Content: %s\n",
		now.Format("15:04:05"),
		ip,
		userID,
		userName,
		path,
		source,
		reason,
		prompt,
	)

	// Write to file (using mutex to prevent concurrent write issues if multiple requests hit at once)
	logLock.Lock()
	defer logLock.Unlock()

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		common.SysError(fmt.Sprintf("Failed to open moderation log file: %v", err))
		return
	}
	defer f.Close()

	if _, err := f.WriteString(logEntry); err != nil {
		common.SysError(fmt.Sprintf("Failed to write to moderation log: %v", err))
	}
}
