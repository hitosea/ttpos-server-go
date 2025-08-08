package middleware

import (
	"ttpos-server-go/app/api/helper"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/config"

	"github.com/gin-gonic/gin"
)

func Internal() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查X-API-KEY
		apiKey := c.GetHeader("X-API-KEY")
		if apiKey != config.JWT.Secret {
			helper.Fail(c, constant.CodeAccessDenied, "Unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}
