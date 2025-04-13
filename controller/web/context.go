package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// web的上下文
type webContext struct {
	*gin.Context
}

// 获取web的上下文
func getWebContext(c *gin.Context) *webContext {
	return &webContext{c}
}

// 显示错误
func (c *webContext) abortError(code int, msg string) {
	c.HTML(code, "error.html", gin.H{
		"msg": msg,
	})
}

// 渲染模板
func (c *webContext) render(template string, data gin.H) {
	c.HTML(http.StatusOK, template, data)
}
