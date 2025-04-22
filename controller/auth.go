package controller

import (
	"one-api/common"

	"github.com/gin-gonic/gin"
)

// 是否是root用户
func isRootUser(c *gin.Context) bool {
	return c.GetInt("role") == common.RoleRootUser
}
