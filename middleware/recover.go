package middleware

import (
	"fmt"
	"net/http"
	"one-api/common"
	"one-api/lang"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RelayPanicRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				common.SysError(fmt.Sprintf("panic detected: %v", err))
				common.SysError(fmt.Sprintf("stacktrace from panic: %s", string(debug.Stack())))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"message": fmt.Sprintf(lang.T(c, "recover.error.panic_detected"), err),
						"type":    "ai_api_panic",
					},
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
